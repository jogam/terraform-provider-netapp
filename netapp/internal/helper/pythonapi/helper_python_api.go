package pythonapi

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/segmentio/ksuid"
	log "github.com/sirupsen/logrus"

	"github.com/hashicorp/go-plugin"

	pyapi "github.com/jogam/terraform-provider-netapp/netapp/internal/keystore"
)

// NetAppAPI the structure for the Python API interaction
// to be refined access to Python API
type NetAppAPI struct {
	pyapi.PythonAPI
	client     *plugin.Client
	clientID   string
	apiFiles   *SyncResult
	status     string
	statusLock *sync.Mutex
	root       string
	Version    string
}

// Stop must be called before API is stopped being used, e.g. plugin shutdown
func (api NetAppAPI) Stop() error {

	api.statusLock.Lock()

	if api.status == "STOPPED" {
		return nil
	}

	succ, err := api.Shutdown(api.clientID)
	if err == nil && succ {
		api.status = "STOPPED"
	} else {
		log.Errorf("API shutdown returned [%v] with error: %v", succ, err.Error)
		if !succ {
			err = fmt.Errorf("API shutdown returned: %v", succ)
		}
	}

	api.statusLock.Unlock()

	// kill the underlying client
	api.client.Kill()

	return err
}

// func stopAPI(apiFiles *SyncResult, folder string, client *plugin.Client) error {
// 	// get the setup script path
// 	shutdownFilePath, err := apiFiles.GetFilePath("scripts/stop_api.sh")
// 	if err != nil {
// 		return err
// 	}
// 	// execute API virtualenv setup and requirements install
// 	out, err := exec.Command("sh", "-c",
// 		fmt.Sprintf("%v %v", shutdownFilePath, folder)).Output()
// 	if err != nil {
// 		log.Errorf("could not stop API, got: %v", err)
// 		return err
// 	}
// 	log.Infof("stop API returned:\n%v", string(out))

// 	// killing the client
// 	client.Kill()

// 	return nil
// }

var requiredAPIScripts = append([]string{
	"scripts/setup_virtualenv.sh",
	"scripts/start_api.sh",
	"requirements.txt",
}, pyapi.APIScripts...)

func apiUp(apiFolder string) bool {
	if _, err := os.Stat(filepath.Join(
		apiFolder, "API_UP")); !os.IsNotExist(err) {
		return true
	}

	return false
}

func ensureAPISetup(folder string, sdkroot string, syncResult *SyncResult) error {

	// check python version
	out, err := exec.Command("sh", "-c",
		"python -c 'import sys; print(sys.version_info[:])'").Output()
	if err != nil {
		log.Errorf("failed to execute python version command, Python installed?")
		return err
	}
	log.Infof("python version: %v", string(out))

	// check virtualenv installed + version?
	out, err = exec.Command("sh", "-c", "virtualenv --version").Output()
	if err != nil {
		log.Errorf("failed to execute virtualenv version command, virtualenv installed?")
		return err
	}
	log.Infof("virtualenv version: %v", string(out))

	// get the setup script path
	setupFilePath, err := syncResult.GetFilePath("scripts/setup_virtualenv.sh")
	if err != nil {
		return err
	}
	// execute API virtualenv setup and requirements install
	out, err = exec.Command("sh", "-c",
		fmt.Sprintf("%v %v", setupFilePath, folder)).Output()
	if err != nil {
		log.Errorf("could not setup virtualenv, got: %v", err)
		return err
	}
	log.Infof("virtualenv setup returned: %v", string(out))

	return nil
}

// CreateAPI TODO: doc for create API call
func CreateAPI(
	folder string, sdkroot string, regport string,
	apiport string) (*NetAppAPI, error) {

	// check if API is already running via running file
	apiRunning := apiUp(folder)

	// synchronize the packr / python source files to OS filesystem / API folder
	syncResult, err := SynchBoxToOS(folder, &requiredAPIScripts)
	if err != nil {
		return nil, err
	}

	if !apiRunning {
		if err = ensureAPISetup(folder, sdkroot, syncResult); err != nil {
			return nil, err
		}
	}

	// if we'd like to not see logging output...
	//log.SetOutput(ioutil.Discard)

	// get the api startup script
	startupFilePath, err := syncResult.GetFilePath("scripts/start_api.sh")
	if err != nil {
		return nil, err
	}

	// create unique id for this client
	// source: https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html
	clientID := ksuid.New().String()

	// start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: pyapi.Handshake,
		Plugins:         pyapi.PluginMap,
		Cmd: exec.Command("sh", "-c",
			fmt.Sprintf("%v %v %v %v %v %v %v",
				startupFilePath,
				folder, sdkroot, regport, // shift arguments
				pyapi.APIMain, apiport, clientID)),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})

	log.Infof("client [%v] created", clientID)

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		log.Errorf("Plugin start Error: %v", err.Error())
		//stopAPI(syncResult, folder, client)
		client.Kill()
		return nil, err
	}

	log.Info("client protocol created")

	// Request the plugin
	raw, err := rpcClient.Dispense("kv_grpc")
	if err != nil {
		log.Errorf("Plugin dispense Error: %v", err.Error())
		//stopAPI(syncResult, folder, client)
		rpcClient.Close()
		client.Kill()
		return nil, err
	}

	log.Info("client plugin dispensed")

	// We should have a KV store now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	kv := raw.(pyapi.PythonAPI)

	log.Info("client plugin interface taken")

	err = kv.Put("version", "shouldbegetonly")
	if err != nil {
		log.Errorf("version set Error: %v", err.Error())
	}

	version, err := kv.Get("version")
	if err != nil {
		log.Errorf("version get Error: %v", err.Error())
	}

	return &NetAppAPI{
		PythonAPI:  kv,
		client:     client,
		clientID:   clientID,
		status:     "RUNNING",
		statusLock: &sync.Mutex{},
		apiFiles:   syncResult,
		root:       folder,
		Version:    version}, nil
}
