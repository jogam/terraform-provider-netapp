package network

import (
	"fmt"
	"strconv"

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
	pythonapi.ResourceInfo
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
	pythonapi.ResourceInfo
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

func IPSpaceGetByName(client *pythonapi.NetAppAPI, name string) (*IPSpaceInfo, error) {
	req := &IPSpaceRequest{Name: name}
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

type BcDomainRequest struct {
	Name       string   `json:"name,omitempty"`
	NewName    string   `json:"new_name,omitempty"`
	Mtu        string   `json:"mtu,omitempty"`
	IPSpace    string   `json:"ipspace,omitempty"`
	Ports      []string `json:"ports,omitempty"`
	StatusOnly string   `json:"statusonly,omitempty"`
}

type BcDomainPortInfo struct {
	Name         string `json:"name"`          // <port>
	UpdateStatus string `json:"update_status"` // <port-update-status>
	StatusDetail string `json:"status_detail"` // <port-update-status-detail>
}

type BcDomainInfo struct {
	pythonapi.ResourceInfo
	Name             string             `json:"name"`          // <broadcast-domain>
	FailoverGroups   []string           `json:"failovergrps"`  // <failover-groups>
	IPSpace          string             `json:"ipspace"`       // <ipspace>
	Mtu              string             `json:"mtu"`           // <mtu>
	PortUpdateStatus string             `json:"update_status"` // <port-update-status-combined>
	Ports            []BcDomainPortInfo `json:"ports"`         // <ports>
	SubnetNames      []string           `json:"subnets"`       // <subnet-names>
}

const bcDomainGetCmd = "NW.BRCDOM.GET"

func BcDomainGet(client *pythonapi.NetAppAPI, name string) (*BcDomainInfo, error) {
	req := &BcDomainRequest{Name: name}
	resp := &BcDomainInfo{}
	err := pythonapi.MakeAPICall(client, bcDomainGetCmd, req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

const bcDomainStatusCmd = "NW.BRCDOM.STATUS"

func BcDomainStatus(client *pythonapi.NetAppAPI, name string) (*BcDomainInfo, error) {
	req := &BcDomainRequest{Name: name, StatusOnly: "set"}
	resp := &BcDomainInfo{}
	err := pythonapi.MakeAPICall(client, bcDomainStatusCmd, req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func BcDomainWaitForInProgressDone(client *pythonapi.NetAppAPI, name string) (string, error) {
	var bcInfo *BcDomainInfo
	var err error
	for true {
		bcInfo, err = BcDomainStatus(client, name)
		if err != nil {
			return "", err
		}

		if bcInfo.PortUpdateStatus != "in_progress" {
			break
		}
	}

	return bcInfo.PortUpdateStatus, nil
}

const bcDomainCreateCmd = "NW.BRCDOM.CREATE"

func BcDomainCreate(client *pythonapi.NetAppAPI, request *BcDomainRequest) (*BcDomainInfo, error) {
	resp := &BcDomainInfo{}
	err := pythonapi.MakeAPICall(client, bcDomainCreateCmd, request, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

const bcDomainRenameCmd = "NW.BRCDOM.RENAME"

func BcDomainRename(
	client *pythonapi.NetAppAPI,
	name string, ipspace string,
	newName string) error {
	req := &BcDomainRequest{Name: name, IPSpace: ipspace, NewName: newName}
	resp := &pythonapi.EmptyResponse{}
	return pythonapi.MakeAPICall(client, bcDomainRenameCmd, req, resp)
}

const bcDomainPortAddCmd = "NW.BRCDOM.PORT.ADD"
const bcDomainPortRemoveCmd = "NW.BRCDOM.PORT.REMOVE"

func BcDomainPortsModify(
	client *pythonapi.NetAppAPI,
	name string, ipspace string,
	portNames []string,
	add bool, remove bool) (*BcDomainInfo, error) {

	if (add && remove) || (!add && !remove) {
		return nil, fmt.Errorf(
			"modify broadcast domain [%s] ports must either add or remove"+
				" got [add,remove]: [%v,%v]",
			name, add, remove)
	}

	req := &BcDomainRequest{Name: name, Ports: portNames, IPSpace: ipspace}
	resp := &BcDomainInfo{}
	var err error
	if add {
		err = pythonapi.MakeAPICall(client, bcDomainPortAddCmd, req, resp)
	} else {
		err = pythonapi.MakeAPICall(client, bcDomainPortRemoveCmd, req, resp)
	}

	if err != nil {
		return nil, err
	}

	return resp, nil
}

const bcDomainUpdateCmd = "NW.BRCDOM.UPDATE"

func BcDomainUpdate(
	client *pythonapi.NetAppAPI,
	name string, ipspace string, mtu int) (*BcDomainInfo, error) {
	req := &BcDomainRequest{Name: name, IPSpace: ipspace, Mtu: strconv.Itoa(mtu)}
	resp := &BcDomainInfo{}
	err := pythonapi.MakeAPICall(client, bcDomainUpdateCmd, req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

const bcDomainDeleteCmd = "NW.BRCDOM.DELETE"

func BcDomainDelete(
	client *pythonapi.NetAppAPI,
	name string, ipspace string) (*BcDomainInfo, error) {

	req := &BcDomainRequest{Name: name, IPSpace: ipspace}
	resp := &BcDomainInfo{}
	err := pythonapi.MakeAPICall(client, bcDomainDeleteCmd, req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type SubnetRequest struct {
	Name            string   `json:"name,omitempty"` // <subnet-name>
	NewName         string   `json:"new_name,omitempty"`
	BroadCastDomain string   `json:"bc_domain,omitempty"` // <broadcast-domain>
	IPSpace         string   `json:"ipspace,omitempty"`   // <ipspace>
	Subnet          string   `json:"subnet,omitempty"`    // <subnet>
	Gateway         string   `json:"gateway,omitempty"`   // <gateway>
	IPRanges        []string `json:"ip_ranges,omitempty"` // <ip-ranges>
}

type SubnetInfo struct {
	pythonapi.ResourceInfo
	SubnetRequest

	IPCount     int `json:"ip_count"` // <used-count>
	IPUsed      int `json:"ip_used"`  // <total-count>
	IPAvailable int `json:"ip_avail"` // <available-count>
}

const subnetGetCmd = "NW.SUBNET.GET"

func SubnetGet(client *pythonapi.NetAppAPI, request *SubnetRequest) (*SubnetInfo, error) {
	response := &SubnetInfo{}
	err := pythonapi.MakeAPICall(client, subnetGetCmd, request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

const subnetCreateCmd = "NW.SUBNET.CREATE"

func SubnetCreate(client *pythonapi.NetAppAPI, request *SubnetRequest) (*SubnetInfo, error) {
	response := &SubnetInfo{}
	err := pythonapi.MakeAPICall(client, subnetCreateCmd, request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

const subnetDeleteCmd = "NW.SUBNET.DELETE"

func SubnetDelete(client *pythonapi.NetAppAPI, request *SubnetRequest) error {
	response := &pythonapi.EmptyResponse{}
	return pythonapi.MakeAPICall(client, subnetDeleteCmd, request, response)
}

const subnetRenameCmd = "NW.SUBNET.RENAME"

func SubnetRename(client *pythonapi.NetAppAPI, request *SubnetRequest) error {
	response := &pythonapi.EmptyResponse{}
	return pythonapi.MakeAPICall(client, subnetRenameCmd, request, response)
}

const subnetIPRangeAddCmd = "NW.SUBNET.IPR.ADD"
const subnetIPRangeRemoveCmd = "NW.SUBNET.IPR.REMOVE"

func SubnetIpRangeModify(
	client *pythonapi.NetAppAPI,
	name string, ipspace string,
	ipRanges []string,
	add bool, remove bool) error {

	if (add && remove) || (!add && !remove) {
		return fmt.Errorf(
			"modify subnet [%s] IP ranges must either add or remove"+
				" got [add,remove]: [%v,%v]",
			name, add, remove)
	}

	req := &SubnetRequest{Name: name, IPRanges: ipRanges, IPSpace: ipspace}
	resp := &pythonapi.EmptyResponse{}
	if add {
		return pythonapi.MakeAPICall(client, subnetIPRangeAddCmd, req, resp)
	}

	return pythonapi.MakeAPICall(client, subnetIPRangeRemoveCmd, req, resp)
}

const subnetModifyCmd = "NW.SUBNET.MODIFY"

func SubnetModify(client *pythonapi.NetAppAPI, request *SubnetRequest) error {
	response := &pythonapi.EmptyResponse{}
	return pythonapi.MakeAPICall(client, subnetModifyCmd, request, response)
}
