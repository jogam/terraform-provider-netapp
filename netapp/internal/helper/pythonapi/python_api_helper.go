package pythonapi

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/segmentio/ksuid"

	"github.com/hashicorp/go-plugin"

	pyapi "github.com/jogam/terraform-provider-netapp/netapp/internal/grpcapi"
)

// NetAppAPI the structure for the Python API interaction
// to be refined access to Python API
type NetAppAPI struct {
	pyapi.PythonAPI
	client   *plugin.Client
	clientID string
}

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

// Stop must be called before API is stopped being used, e.g. plugin shutdown
func (api NetAppAPI) Stop() error {
	succ, err := api.Shutdown(api.clientID)
	if err != nil {
		log.Printf("[ERROR] API shutdown returned [%v] with error: %s", succ, err)
		if !succ {
			err = fmt.Errorf("API shutdown returned: %v", succ)
		}
	}

	return err
}

func ensureAPISetup(folder string, sdkroot string, syncResult *SyncResult) error {

	// check python version
	out, err := exec.Command("sh", "-c",
		"python -c 'import sys; print(sys.version_info[:])'").Output()
	if err != nil {
		log.Printf("[ERROR] failed to execute python version command, Python installed?")
		return err
	}
	log.Printf("[INFO] python version: %v", string(out))

	// check virtualenv installed + version?
	out, err = exec.Command("sh", "-c", "virtualenv --version").Output()
	if err != nil {
		log.Printf("[ERROR] failed to execute virtualenv version command, virtualenv installed?")
		return err
	}
	log.Printf("[INFO] virtualenv version: %v", string(out))

	// get the setup script path
	setupFilePath, err := syncResult.GetFilePath("scripts/setup_virtualenv.sh")
	if err != nil {
		return err
	}
	// execute API virtualenv setup and requirements install
	out, err = exec.Command("sh", "-c",
		fmt.Sprintf("%v %v", setupFilePath, folder)).Output()
	if err != nil {
		log.Printf("[ERROR] could not setup virtualenv, got: %v", err)
		return err
	}
	log.Printf("[INFO] virtualenv setup returned: %v", string(out))

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

	log.Printf("[INFO] client [%v] created", clientID)

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		log.Printf("[ERROR] Plugin start Error: %s", err)
		//stopAPI(syncResult, folder, client)
		client.Kill()
		return nil, err
	}

	log.Printf("[INFO] client protocol created")

	// Request the plugin
	raw, err := rpcClient.Dispense("grpcapi")
	if err != nil {
		log.Printf("[ERROR] Plugin dispense Error: %s", err)
		//stopAPI(syncResult, folder, client)
		rpcClient.Close()
		client.Kill()
		return nil, err
	}

	log.Printf("[INFO] client plugin dispensed")

	// We should have an API now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	apiplug := raw.(pyapi.PythonAPI)

	log.Printf("[INFO] client plugin interface taken")

	return &NetAppAPI{
		PythonAPI: apiplug,
		client:    client,
		clientID:  clientID,
	}, nil
}
