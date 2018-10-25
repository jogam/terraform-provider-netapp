package system

import (
	"github.com/jogam/terraform-provider-netapp/netapp/internal/helper/pythonapi"
)

const connectCmd = "SYS.CONNECT"

// ConnectRequest is the required input for the NetApp API connection
type ConnectRequest struct {
	// TODO: here we would also have HTTP/HTTPS, vserver vs filer etc...
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"pwd"`
}

// ConnectResponse is the returned result for a Connect call
type ConnectResponse struct {
	OntapVersion string `json:"version_ontap"`
	OsVersion    string `json:"version_os"`
}

// Connect returns ONTAP and OS version from NetApp host
func Connect(client *pythonapi.NetAppAPI, request *ConnectRequest) (*ConnectResponse, error) {
	resp := ConnectResponse{}
	err := pythonapi.MakeAPICall(client, connectCmd, request, &resp)

	return &resp, err
}

const getNodeCmd = "SYS.GET.NODE"

// GetNodeRequest to get node information from NetApp
type GetNodeRequest struct {
	Name string `json:"name,omitempty"`
	UUID string `json:"uuid,omitempty"`
}

// GetNodeResponse contains node information for Terraform
type GetNodeResponse struct {
	Name    string `json:"name"`
	Serial  string `json:"serial"`
	ID      string `json:"id"`
	UUID    string `json:"uuid"`
	Version string `json:"version"`
	Healty  bool   `json:"healthy"`
	Uptime  int    `json:"uptime"`
}

// GetNodeByName to find node for a given name
func GetNodeByName(client *pythonapi.NetAppAPI, name string) (*GetNodeResponse, error) {
	request := &GetNodeRequest{
		Name: name, UUID: "",
	}
	resp := GetNodeResponse{}
	err := pythonapi.MakeAPICall(client, getNodeCmd, request, &resp)

	return &resp, err
}

// GetNodeByUUID to retrieve NetApp node data for UUID / Terraform resource ID
func GetNodeByUUID(client *pythonapi.NetAppAPI, uuid string) (*GetNodeResponse, error) {
	request := &GetNodeRequest{
		Name: "", UUID: uuid,
	}
	resp := GetNodeResponse{}
	err := pythonapi.MakeAPICall(client, getNodeCmd, request, &resp)

	return &resp, err
}
