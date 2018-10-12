package main

import (
	"fmt"
	"os"
	"os/exec"

	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/go-plugin/examples/grpc/shared"

	//"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/jogam/terraform-provider-netapp/netapp/keystore"
)

// GRPCClient is an implementation of KV that talks over RPC.
type GRPCClient struct{ client keystore.KeyStoreClient }

func (m *GRPCClient) Put(key string, value []byte) error {
	_, err := m.client.Put(context.Background(), &keystore.PutRequest{
		Key:   key,
		Value: value,
	})
	return err
}

func (m *GRPCClient) Get(key string) ([]byte, error) {
	resp, err := m.client.Get(context.Background(), &keystore.GetRequest{
		Key: key,
	})
	if err != nil {
		return nil, err
	}

	return resp.Value, nil
}

// Here is the gRPC server that GRPCClient talks to.
type GRPCServer struct {
	// This is the real implementation
	Impl KV
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

// KV is the interface that we're exposing as a plugin.
type KV interface {
	Put(key string, value []byte) error
	Get(key string) ([]byte, error)
}

// This is the implementation of plugin.GRPCPlugin so we can serve/consume this.
type KVGRPCPlugin struct {
	// GRPCPlugin must still implement the Plugin interface
	plugin.Plugin
	// Concrete implementation, written in Go. This is only used for plugins
	// that are written in Go.
	Impl KV
}

func (p *KVGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	keystore.RegisterKeyStoreServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

func (p *KVGRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: keystore.NewKeyStoreClient(c)}, nil
}

func exit(client *plugin.Client, value int) {
	client.Kill()
	os.Exit(value)
}

func main() {
	// We don't want to see the plugin logs.
	//log.SetOutput(ioutil.Discard)

	// We're a host. Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  Handshake,
		Plugins:          PluginMap,
		Cmd:              exec.Command("sh", "-c", os.Getenv("KV_PLUGIN")),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})
	defer client.Kill()

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		fmt.Println("Error:", err.Error())
		exit(client, 1)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("kv_grpc")
	if err != nil {
		fmt.Println("Error:", err.Error())
		exit(client, 1)
	}

	// We should have a KV store now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	kv := raw.(shared.KV)
	os.Args = os.Args[1:]
	switch os.Args[0] {
	case "get":
		result, err := kv.Get(os.Args[1])
		if err != nil {
			fmt.Println("Error:", err.Error())
			exit(client, 1)
		}

		fmt.Println(string(result))

	case "put":
		err := kv.Put(os.Args[1], []byte(os.Args[2]))
		if err != nil {
			fmt.Println("Error:", err.Error())
			exit(client, 1)
		}

	default:
		fmt.Printf("Please only use 'get' or 'put', given: %q", os.Args[0])
		exit(client, 1)
	}
	exit(client, 0)
}
