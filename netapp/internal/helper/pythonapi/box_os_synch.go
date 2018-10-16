package pythonapi

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/gobuffalo/packr"
)

// fileSyncResult containes the information for one individual file
type fileSyncResult struct {
	BoxPath   string
	FilePath  string
	Available bool
	Updated   bool
}

// SyncResult contains the box to OS sync results with
// accessor functions Contains/GetFilePath
type SyncResult struct {
	files []fileSyncResult
}

func (sR SyncResult) get(boxPath string) (*fileSyncResult, error) {
	for _, file := range sR.files {
		if file.BoxPath == boxPath {
			return &file, nil
		}
	}

	return nil, fmt.Errorf("no file available for path: %v", boxPath)
}

// FileCount returns the number of available files on the OS
func (sR SyncResult) FileCount() int {
	var fileCnt int
	for _, file := range sR.files {
		if file.Available {
			fileCnt++
		}
	}

	return fileCnt
}

// Contains returns true if provided boxPath exists and is available
func (sR SyncResult) Contains(boxPath string) bool {
	file, err := sR.get(boxPath)
	if err != nil {
		return false
	}

	return file.Available
}

// GetFilePath returns the OS file path for requested boxPath
func (sR SyncResult) GetFilePath(boxPath string) (string, error) {
	file, err := sR.get(boxPath)
	if err != nil {
		return "", err
	}

	if !file.Available {
		return "", fmt.Errorf("api file [%v] could not be extracted to OS", boxPath)
	}
	return file.FilePath, nil
}

// SynchBoxToOS synchronize box folder ./python to provided folder
func SynchBoxToOS(folder string, requiredFiles *[]string) (*SyncResult, error) {

	// create packr box
	box := packr.NewBox("./python")

	// ensure all required files are present in box
	for idx := range *requiredFiles {
		reqFile := (*requiredFiles)[idx]
		if !box.Has(reqFile) {
			return nil, fmt.Errorf("missing python API file [%v] in BOX", reqFile)
		}
	}

	// ensure folder exists on host drive
	os.MkdirAll(folder, os.ModePerm)

	// walk the packr box content and ensure files in folder are identical
	done := make(chan struct{})
	defer close(done)

	resc, errc := createPythonProjectFolder(done, folder, box)
	var errMsgs strings.Builder

	if err := <-errc; err != nil {
		fmt.Fprintf(&errMsgs, "OS python folder issue: %v", err)
	}

	files := []fileSyncResult{}
	for res := range resc {
		if res.err != nil {
			fmt.Fprintf(&errMsgs, "box file [%v] caused: %v", res.boxpath, res.err)
			continue
		}

		// check that all import files are there and available!
		log.Infof("box file [%v] avail: %v updated: %v",
			res.boxpath, res.available, res.updated)
		files = append(files, fileSyncResult{
			res.boxpath, res.ospath, res.available, res.updated})
	}

	if errMsgs.Len() > 0 {
		return nil, errors.New(errMsgs.String())
	}

	return &SyncResult{files}, nil
}

// DirDeleteRecursive delete directory and all its content provided by path
func DirDeleteRecursive(path string) error {
	d, err := os.Open(path)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		log.Infof("removeall on: %v", name)
		err = os.RemoveAll(filepath.Join(path, name))
		if err != nil {
			return err
		}
	}

	return os.RemoveAll(path)
}

//**************************************************************************
//   the actual extraction of the python source to source OS
//   MD5 Sum Source: https://blog.golang.org/pipelines#TOC_8.
//**************************************************************************

// a result is a python source file with its OS availability status
type prjFileResult struct {
	boxpath   string
	ospath    string
	available bool
	updated   bool
	err       error
}

func createPythonProjectFolder(done <-chan struct{},
	folder string, box packr.Box) (<-chan prjFileResult, <-chan error) {

	resc := make(chan prjFileResult)
	errc := make(chan error, 1)
	go func() {
		var wg sync.WaitGroup
		err := box.Walk(func(path string, file packr.File) error {
			fInfo, err := file.FileInfo()
			if err != nil {
				return err
			}

			if !fInfo.Mode().IsRegular() {
				return nil
			}

			wg.Add(1)
			go func() {
				select {
				case resc <- processFilePath(folder, path, box):
				case <-done:
				}
				wg.Done()
			}()

			// abort walk if done is closed... e.g. cancel was triggered
			select {
			case <-done:
				return errors.New("create Python project folder canceled")
			default:
				return nil
			}
		})

		// box.Walk has returned so all calls to wg.Add are done.
		// Start a go routeing to close c once all sends are done.
		go func() {
			wg.Wait()
			close(resc)
		}()

		// add possible errors, no select needed here since errc is buffered
		errc <- err
	}()

	return resc, errc
}

// processFilePath returns a {result} for the extraction of the file
// for given path from box to the host OS at provided osRoot folder
func processFilePath(osRoot string, path string, box packr.Box) prjFileResult {
	// create the OS full file path from path
	osFilePath := filepath.Join(osRoot, path)

	// try to read box file at path to byte[]
	boxFileData, bfdErr := box.MustBytes(path)

	if bfdErr != nil {
		// could not read box file, no goodie
		return prjFileResult{path, osFilePath, false, false, bfdErr}
	}

	// try reading the file from OS
	osFileData, ofdErr := ioutil.ReadFile(osFilePath)

	// flag to indicate if BOX file needs to be written
	wrFileToOS := false

	if ofdErr != nil {
		// either host file corrupt or not existent
		osFileDir, _ := filepath.Split(osFilePath)
		// ensure file folder exists
		os.MkdirAll(osFileDir, os.ModePerm)
		// flag file needs to be written
		wrFileToOS = true
	}

	boxfMD5 := md5.Sum(boxFileData)
	osfMD5 := md5.Sum(osFileData)

	log.Debugf("processing path [%v] box vs OS: %v vs %v", path, boxfMD5, osfMD5)

	if boxfMD5 != osfMD5 {
		// OS file different from package file, needs writing
		wrFileToOS = true
	}

	if wrFileToOS {
		// write box file bytes to file <-- think about permission (read/exec only?)
		fwrErr := ioutil.WriteFile(osFilePath, boxFileData, os.ModePerm)
		return prjFileResult{path, osFilePath, fwrErr == nil, true, fwrErr}
	}

	return prjFileResult{path, osFilePath, true, false, nil}
}
