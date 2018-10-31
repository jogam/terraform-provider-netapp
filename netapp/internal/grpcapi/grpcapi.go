package grpcapi

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// APIScripts the scripts required for this API
var APIScripts = []string{
	"__init__.py",
	"grpcapi_pb2_grpc.py",
	"grpcapi_pb2.py",
	"grpcapi.py",
	"registry.py",
	"apicmd/__init__.py",
	"apicmd/system.py",
	"apicmd/testing.py",
	"apicmd/network.py",
}

// APIMain is the script to be called from python
var APIMain = "grpcapi.py"

// PythonAPI is the interface exposed
type PythonAPI interface {
	GRPCNetAppAPI
}

// GRPCNetAppApi is the interface that is implemented by GRPC/Python
type GRPCNetAppAPI interface {
	Call(cmd string, data []byte) (bool, string, []byte, error)
	Shutdown(clientID string) (bool, error)
}

// GRPCClient is an implementation of KV that talks over RPC.
type gRPCClient struct{ client GRPCNetAppApiClient }

func (m *gRPCClient) Call(cmd string, data []byte) (bool, string, []byte, error) {
	// TODO: expose context created here to upstream for timeout config
	resp, err := m.client.Call(context.Background(), &CallRequest{
		Cmd:  cmd,
		Data: data,
	})
	if err != nil {
		return false, "", nil, err
	}

	return resp.Success, resp.Errmsg, resp.Data, err
}

func (m *gRPCClient) Shutdown(clientID string) (bool, error) {
	// TODO: expose context created here to upstream for timeout config
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
	Impl GRPCNetAppAPI
}

func (m *gRPCServer) Call(
	ctx context.Context,
	req *CallRequest) (*CallResponse, error) {
	succ, errmsg, data, err := m.Impl.Call(req.Cmd, req.Data)
	return &CallResponse{
		Success: succ, Errmsg: errmsg, Data: data}, err
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
	MagicCookieKey:   "NETAPP_API_PLUGIN",
	MagicCookieValue: "N3tA99c00k13s3CReTk8",
}

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	"grpcapi": &gRPCApiPlugin{},
}

// This is the implementation of plugin.GRPCPlugin so we can serve/consume this.
type gRPCApiPlugin struct {
	// GRPCPlugin must still implement the Plugin interface
	plugin.Plugin
	// Concrete implementation, written in Go. This is only used for plugins
	// that are written in Go.
	Impl GRPCNetAppAPI
}

func (p *gRPCApiPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	RegisterGRPCNetAppApiServer(s, &gRPCServer{Impl: p.Impl})
	return nil
}

func (p *gRPCApiPlugin) GRPCClient(
	ctx context.Context, broker *plugin.GRPCBroker,
	c *grpc.ClientConn) (interface{}, error) {
	return &gRPCClient{client: NewGRPCNetAppApiClient(c)}, nil
}
