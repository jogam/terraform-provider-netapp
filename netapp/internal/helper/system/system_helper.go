package system

import (
	"fmt"

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

const nodeGetCmd = "SYS.NODE.GET"

// NodeGetRequest to get node information from NetApp
type NodeGetRequest struct {
	Name string `json:"name,omitempty"`
	UUID string `json:"uuid,omitempty"`
}

// NodeInfo contains node information for Terraform
type NodeInfo struct {
	pythonapi.ResourceInfo
	Name    string `json:"name"`
	Serial  string `json:"serial"`
	ID      string `json:"id"`
	UUID    string `json:"uuid"`
	Version string `json:"version"`
	Healty  bool   `json:"healthy"`
	Uptime  int    `json:"uptime"`
}

// NodeGetByName to find node for a given name
func NodeGetByName(client *pythonapi.NetAppAPI, name string) (*NodeInfo, error) {
	request := &NodeGetRequest{
		Name: name, UUID: "",
	}
	resp := NodeInfo{}
	err := pythonapi.MakeAPICall(client, nodeGetCmd, request, &resp)

	return &resp, err
}

// NodeGetByUUID to retrieve NetApp node data for UUID / Terraform resource ID
func NodeGetByUUID(client *pythonapi.NetAppAPI, uuid string) (*NodeInfo, error) {
	request := &NodeGetRequest{
		Name: "", UUID: uuid,
	}
	resp := NodeInfo{}
	err := pythonapi.MakeAPICall(client, nodeGetCmd, request, &resp)

	return &resp, err
}

const portGetInfoCmd = "SYS.PORT.GET"

type PortGetRequest struct {
	NodeName string `json:"node"` // <node>
	PortName string `json:"port"` // <port>
}

type PortInfo struct {
	// all parameters as strings to overcome bool/int omitempty behaviour
	PortGetRequest
	pythonapi.ResourceInfo

	AutoRevertDelay string `json:"auto_rev_delay"` // <autorevert-delay>
	IgnoreHealth    string `json:"ignr_health"`    // <ignore-health-status>
	IPSpace         string `json:"ipspace"`        // <ipspace>
	Role            string `json:"role"`           // <role></role>

	AdminUp     string `json:"admin_up"`     // <is-administrative-up>
	AdminMtu    string `json:"admin_mtu"`    // <mtu-admin>
	AdminAuto   string `json:"admin_auto"`   // <is-administrative-auto-negotiate>
	AdminSpeed  string `json:"admin_speed"`  // <administrative-speed>
	AdminDuplex string `json:"admin_duplex"` // <administrative-duplex>
	AdminFlow   string `json:"admin_flow"`   // <administrative-flowcontrol>

	Status          string `json:"status"`           // <link-status>
	Health          string `json:"health"`           // <health-status>
	Mac             string `json:"mac"`              // <mac-address>
	BroadCastDomain string `json:"broadcast_domain"` // <broadcast-domain>
	Mtu             string `json:"mtu"`              // <mtu>
	Auto            string `json:"auto"`             // <is-operational-auto-negotiate>
	Speed           string `json:"speed"`            // <operational-speed>
	Duplex          string `json:"duplex"`           // <operational-duplex>
	Flow            string `json:"flow"`             // <operational-flowcontrol>

	Type string `json:"type"` // <port-type></port-type>

	VlanID   string `json:"vlan_id,omitempty"`   // <vlan-id></vlan-id>
	VlanNode string `json:"vlan_node,omitempty"` // <vlan-node></vlan-node>
	VlanPort string `json:"vlan_port,omitempty"` // <vlan-port></vlan-port>

	// <remote-device-id></remote-device-id>
}

func PortGetByNames(
	client *pythonapi.NetAppAPI,
	nodeName string, portName string) (*PortInfo, error) {

	request := &PortGetRequest{
		NodeName: nodeName, PortName: portName,
	}
	resp := PortInfo{}
	err := pythonapi.MakeAPICall(client, portGetInfoCmd, request, &resp)
	return &resp, err
}

const portModifyCmd = "SYS.PORT.MODIFY"

type PortModifyRequest struct {
	// all parameters as strings to overcome bool/int omitempty behaviour
	PortGetRequest
	Up     string `json:"up,omitempty"`     // <is-administrative-up>
	Mtu    string `json:"mtu,omitempty"`    // <mtu>
	Auto   string `json:"auto,omitempty"`   // <is-administrative-auto-negotiate>
	Duplex string `json:"duplex,omitempty"` // <administrative-duplex>
	Flow   string `json:"flow,omitempty"`   // <administrative-flowcontrol>
	Speed  string `json:"speed,omitempty"`  // <administrative-speed>

	AutoRevertDelay string `json:"auto_rev_delay,omitempty"` // <autorevert-delay>
	IgnoreHealth    string `json:"ignr_health,omitempty"`    // <ignore-health-status>
	IPSpace         string `json:"ipspace,omitempty"`        // <ipspace>
	Role            string `json:"role,omitempty"`           // <role>
}

func PortModify(
	client *pythonapi.NetAppAPI,
	request *PortModifyRequest) error {

	resp := pythonapi.EmptyResponse{}
	return pythonapi.MakeAPICall(client, portModifyCmd, request, &resp)
}

type PortFindResult struct {
	Names []string `json:"ports"`
}

const portFindByPatternCmd = "SYS.PORT.FIND.PATTERN"

func PortFindByNamePattern(
	client *pythonapi.NetAppAPI,
	nodeName, pattern string) (*PortFindResult, error) {
	request := &PortGetRequest{NodeName: nodeName, PortName: pattern}
	response := &PortFindResult{}
	err := pythonapi.MakeAPICall(client, portFindByPatternCmd, request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

type PortGroupGetRequest struct {
	NodeName  string `json:"node"` // <node>
	GroupName string `json:"name"` // <ifgrp-name>
}

type PortGroupModifyRequest struct {
	PortGroupGetRequest

	Mode             string   `json:"mode"`  //<mode>
	LoadDistribution string   `json:"dist"`  //<distribution-function>
	Ports            []string `json:"ports"` //<ports>
}

type PortGroupInfo struct {
	PortGroupModifyRequest
	pythonapi.ResourceInfo

	GroupLinkStatus string   `json:"part"`       // <port-participation>
	PortsDown       []string `json:"ports_down"` // <down-ports>
	PortsUp         []string `json:"ports_up"`   // <up-ports>
}

const portGroupGetCmd = "SYS.PORTGROUP.GET"

func PortGroupGetByNames(
	client *pythonapi.NetAppAPI,
	nodeName, groupName string) (*PortGroupInfo, error) {

	request := PortGroupGetRequest{NodeName: nodeName, GroupName: groupName}
	response := PortGroupInfo{}
	err := pythonapi.MakeAPICall(client, portGroupGetCmd, &request, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

const portGroupCreateCmd = "SYS.PORTGROUP.CREATE"

func PortGroupCreate(
	client *pythonapi.NetAppAPI,
	request *PortGroupModifyRequest) error {
	resp := pythonapi.EmptyResponse{}
	return pythonapi.MakeAPICall(client, portGroupCreateCmd, request, &resp)
}

const portGroupPortAddCmd = "SYS.PORTGROUP.PORT.ADD"
const portGroupPortRemoveCmd = "SYS.PORTGROUP.PORT.REMOVE"

func PortGroupPortsModify(
	client *pythonapi.NetAppAPI,
	node, name string,
	portNames []string,
	add bool, remove bool) error {

	if (add && remove) || (!add && !remove) {
		return fmt.Errorf(
			"modify port group [%s] ports must either add or remove"+
				" got [add,remove]: [%v,%v]",
			name, add, remove)
	}

	req := &PortGroupModifyRequest{Ports: portNames}
	req.NodeName = node
	req.GroupName = name

	resp := &pythonapi.EmptyResponse{}
	var err error
	if add {
		err = pythonapi.MakeAPICall(client, portGroupPortAddCmd, req, resp)
	} else {
		err = pythonapi.MakeAPICall(client, portGroupPortRemoveCmd, req, resp)
	}

	return err
}

const portGroupDeleteCmd = "SYS.PORTGROUP.DELETE"

func PortGroupDelete(
	client *pythonapi.NetAppAPI,
	nodeName, groupName string) error {
	req := &PortGroupGetRequest{NodeName: nodeName, GroupName: groupName}
	resp := pythonapi.EmptyResponse{}
	return pythonapi.MakeAPICall(client, portGroupDeleteCmd, req, &resp)
}

const aggrGetCmd = "SYS.AGGR.GET"

// AggrGetRequest to get node information from NetApp
type AggrGetRequest struct {
	Name  string   `json:"name,omitempty"`
	UUID  string   `json:"uuid,omitempty"`
	Nodes []string `json:"nodes,omitempty"`
}

// AggrInfo contains aggregate information for Terraform
type AggrInfo struct {
	pythonapi.ResourceInfo
	AggrGetRequest

	FlexVolCount    int `json:"flexvol_cnt"`
	PctUsedCapacity int `json:"pct_used_cap"`
	PctUsedPhysical int `json:"pct_used_phys"`
	SizeTotal       int `json:"size_total"`
	SizeUsed        int `json:"size_used"`
	SizeAvailable   int `json:"size_avail"`
	SizeReserved    int `json:"size_reserve"`
}

// AggrGetByName to find aggregate for given name
func AggrGetByName(client *pythonapi.NetAppAPI, name string) (*AggrInfo, error) {
	request := &AggrGetRequest{Name: name}
	resp := AggrInfo{}
	err := pythonapi.MakeAPICall(client, aggrGetCmd, request, &resp)

	return &resp, err
}

// AggrGetByUUID to find aggregate for given UUID
func AggrGetByUUID(client *pythonapi.NetAppAPI, uuid string) (*AggrInfo, error) {
	request := &AggrGetRequest{UUID: uuid}
	resp := AggrInfo{}
	err := pythonapi.MakeAPICall(client, aggrGetCmd, request, &resp)

	return &resp, err
}
