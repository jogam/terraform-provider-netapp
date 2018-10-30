package network

import (
	"github.com/jogam/terraform-provider-netapp/netapp/internal/helper/pythonapi"
)

const vlanGetCmd = "NW.VLAN.GET"

type VlanRequest struct {
	NodeName   string `json:"node_name,omitempty"`   // <node>
	ParentName string `json:"parent_name,omitempty"` // <parent-interface>
	VlanID     string `json:"vlan_id,omitempty"`     // <vlanid>
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
	err := pythonapi.MakeAPICall(client, vlanCreateCmd, request, &resp)
	return err
}

const vlanDeleteCmd = "NW.VLAN.DELETE"

func VlanDelete(client *pythonapi.NetAppAPI, request *VlanRequest) error {
	resp := pythonapi.EmptyResponse{}
	err := pythonapi.MakeAPICall(client, vlanDeleteCmd, request, &resp)
	return err
}
