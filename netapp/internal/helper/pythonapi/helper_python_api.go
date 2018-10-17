package pythonapi

import (
	"fmt"
	"os/exec"
	"sync"

	"context"

	log "github.com/sirupsen/logrus"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/jogam/terraform-provider-netapp/netapp/internal/keystore"
)

// KV is the interface that we're exposing as a plugin.
type pythonapi interface {
	Put(key string, value string) error
	Get(key string) (string, error)
}

// NetAppAPI the structure for the Python API interaction
// to be refined access to Python API
type NetAppAPI struct {
	pythonapi
	client     *plugin.Client
	apiFiles   *SyncResult
	status     string
	statusLock *sync.Mutex
	root       string
	Version    string
}

/* // Get value of given key
func (api NetAppAPI) Get(key string) (string, error) {
	data, err := api.Get(key)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Put a key value pair to the API
func (api NetAppAPI) Put(key string, value string) error {
	return api.Put(key, []byte(value))
} */

// Stop must be called before API is stopped being used, e.g. plugin shutdown
func (api NetAppAPI) Stop() error {

	api.statusLock.Lock()

	if api.status == "STOPPED" {
		return nil
	}

	err := stopAPI(api.apiFiles, api.root, api.client)
	if err == nil {
		api.status = "STOPPED"
	}

	api.statusLock.Unlock()

	return err
}

func stopAPI(apiFiles *SyncResult, folder string, client *plugin.Client) error {
	// get the setup script path
	shutdownFilePath, err := apiFiles.GetFilePath("scripts/stop_api.sh")
	if err != nil {
		return err
	}
	// execute API virtualenv setup and requirements install
	out, err := exec.Command("sh", "-c",
		fmt.Sprintf("%v %v", shutdownFilePath, folder)).Output()
	if err != nil {
		log.Errorf("could not stop API, got: %v", err)
		return err
	}
	log.Infof("stop API returned:\n%v", string(out))

	// killing the client
	client.Kill()

	return nil
}

var requiredAPIScripts = []string{
	"scripts/setup_virtualenv.sh",
	"scripts/start_api.sh",
	"scripts/stop_api.sh",
	"requirements.txt",
	"__init__.py",
	"keystore_pb2_grpc.py",
	"keystore_pb2.py",
	"keystore.py",
	"if_vlan_get.py",
}

// CreateAPI TODO: doc for create API call
func CreateAPI(folder string, sdkroot string) (*NetAppAPI, error) {

	// check python version
	out, err := exec.Command("sh", "-c",
		"python -c 'import sys; print(sys.version_info[:])'").Output()
	if err != nil {
		log.Errorf("failed to execute python version command, Python installed?")
		return nil, err
	}
	log.Infof("python version: %v", string(out))

	// check virtualenv installed + version?
	out, err = exec.Command("sh", "-c", "virtualenv --version").Output()
	if err != nil {
		log.Errorf("failed to execute virtualenv version command, virtualenv installed?")
		return nil, err
	}
	log.Infof("virtualenv version: %v", string(out))

	syncResult, err := SynchBoxToOS(folder, &requiredAPIScripts)
	if err != nil {
		return nil, err
	}

	// get the setup script path
	setupFilePath, err := syncResult.GetFilePath("scripts/setup_virtualenv.sh")
	if err != nil {
		return nil, err
	}
	// execute API virtualenv setup and requirements install
	out, err = exec.Command("sh", "-c",
		fmt.Sprintf("%v %v", setupFilePath, folder)).Output()
	if err != nil {
		log.Errorf("could not setup virtualenv, got: %v", err)
		return nil, err
	}
	log.Infof("virtualenv setup returned: %v", string(out))

	/* 	for _, prjFile := range prjFileResults {
	   		if prjFile.BoxPath == "__init__.py" {
	   			// read version number
	   			versionBytes, err := ioutil.ReadFile(prjFile.FilePath)
	   			if err != nil {
	   				log.Errorf("could not read API version from file [%v]", prjFile.FilePath)
	   			}

	   			versionStr := string(versionBytes)
	   			log.Infof("API version content: %v", versionStr)
	   		}

	   		if filepath.Ext(prjFile.BoxPath) == ".py" {
	   			// got a python file lets replace the sdkroot in the file
	   			fileBytes, err := ioutil.ReadFile(prjFile.FilePath)
	   			if err != nil {
	   				return nil, err
	   			}

	   			re := regexp.MustCompile("LATER")
	   			fileBytes = re.ReplaceAll(fileBytes, []byte(sdkroot))
	   			err = ioutil.WriteFile(prjFile.FilePath, fileBytes, os.ModePerm)
	   			if err != nil {
	   				return nil, err
	   			}

	   		}
	   	}
	*/

	// if we don't want to see the plugin logs.
	//log.SetOutput(ioutil.Discard)

	// get the api startup script
	startupFilePath, err := syncResult.GetFilePath("scripts/start_api.sh")
	if err != nil {
		return nil, err
	}

	// start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: Handshake,
		Plugins:         PluginMap,
		Cmd: exec.Command("sh", "-c",
			fmt.Sprintf("%v %v %v",
				startupFilePath, folder, "keystore.py")),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})

	log.Info("client created")

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		log.Errorf("Plugin start Error: %v", err.Error())
		stopAPI(syncResult, folder, client)
		return nil, err
	}

	log.Info("client protocol created")

	// Request the plugin
	raw, err := rpcClient.Dispense("kv_grpc")
	if err != nil {
		log.Errorf("Plugin dispense Error: %v", err.Error())
		stopAPI(syncResult, folder, client)
		return nil, err
	}

	log.Info("client plugin dispensed")

	// We should have a KV store now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	kv := raw.(pythonapi)

	log.Info("client plugin interface taken")

	return &NetAppAPI{
		pythonapi:  kv,
		client:     client,
		status:     "RUNNING",
		statusLock: &sync.Mutex{},
		apiFiles:   syncResult,
		root:       folder,
		Version:    "TBA"}, nil
}

// GRPCClient is an implementation of KV that talks over RPC.
type GRPCClient struct{ client keystore.KeyStoreClient }

func (m *GRPCClient) Put(key string, value string) error {
	_, err := m.client.Put(context.Background(), &keystore.PutRequest{
		Key:   key,
		Value: value,
	})
	return err
}

func (m *GRPCClient) Get(key string) (string, error) {
	resp, err := m.client.Get(context.Background(), &keystore.GetRequest{
		Key: key,
	})
	if err != nil {
		return "", err
	}

	return resp.Value, nil
}

// Here is the gRPC server that GRPCClient talks to.
type GRPCServer struct {
	// This is the real implementation
	Impl pythonapi
}

func (m *GRPCServer) Put(
	ctx context.Context,
	req *keystore.PutRequest) (*keystore.Empty, error) {
	return &keystore.Empty{}, m.Impl.Put(req.Key, req.Value)
}

func (m *GRPCServer) Get(
	ctx context.Context,
	req *keystore.GetRequest) (*keystore.GetResponse, error) {
	v, err := m.Impl.Get(req.Key)
	return &keystore.GetResponse{Value: v}, err
}

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	"kv_grpc": &KVGRPCPlugin{},
}

// This is the implementation of plugin.GRPCPlugin so we can serve/consume this.
type KVGRPCPlugin struct {
	// GRPCPlugin must still implement the Plugin interface
	plugin.Plugin
	// Concrete implementation, written in Go. This is only used for plugins
	// that are written in Go.
	Impl pythonapi
}

func (p *KVGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	keystore.RegisterKeyStoreServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

func (p *KVGRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: keystore.NewKeyStoreClient(c)}, nil
}
