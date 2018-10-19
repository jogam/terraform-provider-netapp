package keystore

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// KV is the interface that we're exposing as a plugin.
type PythonAPI interface {
	Put(key string, value string) error
	Get(key string) (string, error)
	Shutdown(clientID string) (bool, error)
}

// GRPCClient is an implementation of KV that talks over RPC.
type gRPCClient struct{ client KeyStoreClient }

func (m *gRPCClient) Put(key string, value string) error {
	_, err := m.client.Put(context.Background(), &PutRequest{
		Key:   key,
		Value: value,
	})
	return err
}

func (m *gRPCClient) Get(key string) (string, error) {
	resp, err := m.client.Get(context.Background(), &GetRequest{
		Key: key,
	})
	if err != nil {
		return "", err
	}

	return resp.Value, nil
}

func (m *gRPCClient) Shutdown(clientID string) (bool, error) {
	resp, err := m.client.Shutdown(context.Background(), &ShutdownRequest{
		Clientid: clientID,
	})
	if err != nil {
		return false, err
	}

	return resp.Result, nil
}

// Here is the gRPC server that GRPCClient talks to.
type gRPCServer struct {
	// This is the real implementation
	Impl PythonAPI
}

func (m *gRPCServer) Put(
	ctx context.Context,
	req *PutRequest) (*Empty, error) {
	return &Empty{}, m.Impl.Put(req.Key, req.Value)
}

func (m *gRPCServer) Get(
	ctx context.Context,
	req *GetRequest) (*GetResponse, error) {
	v, err := m.Impl.Get(req.Key)
	return &GetResponse{Value: v}, err
}

func (m *gRPCServer) Shutdown(
	ctx context.Context,
	req *ShutdownRequest) (*ShutdownResponse, error) {
	v, err := m.Impl.Shutdown(req.Clientid)
	return &ShutdownResponse{Result: v}, err
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
	Impl PythonAPI
}

func (p *KVGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	RegisterKeyStoreServer(s, &gRPCServer{Impl: p.Impl})
	return nil
}

func (p *KVGRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &gRPCClient{client: NewKeyStoreClient(c)}, nil
}
