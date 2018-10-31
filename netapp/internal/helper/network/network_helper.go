package network

import (
	"github.com/jogam/terraform-provider-netapp/netapp/internal/helper/pythonapi"
)

const vlanGetCmd = "NW.VLAN.GET"

type VlanConfig interface {
	GetNodeName() string
	GetParentName() string
	GetVlanID() string
}

type VlanRequest struct {
	NodeName   string `json:"node_name,omitempty"`   // <node>
	ParentName string `json:"parent_name,omitempty"` // <parent-interface>
	VlanID     string `json:"vlan_id,omitempty"`     // <vlanid>
}

func (req VlanRequest) GetNodeName() string {
	return req.NodeName
}

func (req VlanRequest) GetParentName() string {
	return req.ParentName
}

func (req VlanRequest) GetVlanID() string {
	return req.VlanID
}

type VlanInfo struct {
	VlanRequest
	Name string `json:"name,omitempty"` // <interface-name>
}

func VlanGet(client *pythonapi.NetAppAPI, request *VlanRequest) (*VlanInfo, error) {
	resp := VlanInfo{}
	err := pythonapi.MakeAPICall(client, vlanGetCmd, request, &resp)

	return &resp, err
}

const vlanCreateCmd = "NW.VLAN.CREATE"

func VlanCreate(client *pythonapi.NetAppAPI, request *VlanRequest) error {
	resp := pythonapi.EmptyResponse{}
	return pythonapi.MakeAPICall(client, vlanCreateCmd, request, &resp)
}

const vlanDeleteCmd = "NW.VLAN.DELETE"

func VlanDelete(client *pythonapi.NetAppAPI, request *VlanRequest) error {
	resp := pythonapi.EmptyResponse{}
	return pythonapi.MakeAPICall(client, vlanDeleteCmd, request, &resp)
}

const ipspaceGetCmd = "NW.IPSPACE.GET"

type IPSpaceRequest struct {
	Name    string `json:"name,omitempty"`
	UUID    string `json:"uuid,omitempty"`
	NewName string `json:"new_name,omitempty"`
}

type IPSpaceInfo struct {
	Name             string   `json:"name"`       // <ipspace>
	UUID             string   `json:"uuid"`       // <uuid>
	BroadCastDomains []string `json:"bc_domains"` // <broadcast-domains>
	Ports            []string `json:"ports"`      // <ports>
	VServers         []string `json:"vservers"`   // <vservers>
}

func IPSpaceGetByUUID(client *pythonapi.NetAppAPI, uuid string) (*IPSpaceInfo, error) {
	req := &IPSpaceRequest{UUID: uuid}
	resp := &IPSpaceInfo{}
	err := pythonapi.MakeAPICall(client, ipspaceGetCmd, req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

const ipspaceCreateCmd = "NW.IPSPACE.CREATE"

func IPSpaceCreate(client *pythonapi.NetAppAPI, name string) (string, error) {
	req := &IPSpaceRequest{Name: name}
	resp := &IPSpaceInfo{}
	err := pythonapi.MakeAPICall(client, ipspaceCreateCmd, req, resp)
	if err != nil {
		return "", err
	}

	return resp.UUID, nil
}

const ipspaceUpdateCmd = "NW.IPSPACE.UPDATE"

func IPSpaceUpdate(client *pythonapi.NetAppAPI, name string, newName string) error {
	req := &IPSpaceRequest{Name: name, NewName: newName}
	resp := &pythonapi.EmptyResponse{}
	return pythonapi.MakeAPICall(client, ipspaceUpdateCmd, req, resp)
}

const ipspaceDeleteCmd = "NW.IPSPACE.DELETE"

func IPSpaceDelete(client *pythonapi.NetAppAPI, name string) error {
	req := &IPSpaceRequest{Name: name}
	resp := &pythonapi.EmptyResponse{}
	return pythonapi.MakeAPICall(client, ipspaceDeleteCmd, req, resp)
}
