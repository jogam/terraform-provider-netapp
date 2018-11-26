package svm

import (
	"github.com/jogam/terraform-provider-netapp/netapp/internal/helper/pythonapi"
)

type ProtocolInfo struct {
	Type string `json:"type"`
}

type SvmRequest struct {
	Name          string         `json:"name,omitempty"`
	UUID          string         `json:"uuid,omitempty"`
	IPSpace       string         `json:"ipspace"`
	RootAggr      string         `json:"root_aggr,omitempty"`
	RootSecStyle  string         `json:"root_sec_style,omitempty"`
	RootName      string         `json:"root_name,omitempty"`
	RootSize      string         `json:"root_size,omitempty"` // must come from SVM direct, e.g. connect via second call...
	RootRetention string         `json:"root_retent,omitempty"`
	Protocols     []ProtocolInfo `json:"protocols,omitempty"` // also from SVM
}

type SvmInfo struct {
	pythonapi.ResourceInfo
	SvmRequest

	ConfigLocked bool   `json:"locked"`
	OperState    string `json:"oper_state"`
	SvmState     string `json:"svm_state"`

	ProtoEnabled  []string `json:"proto_enabled"`
	ProtoInactive []string `json:"proto_inactive"`
}

//******************************************************************
// status section
//******************************************************************

const svmGetCmd = "SVM.GET"

func SvmGetByUUID(client *pythonapi.NetAppAPI, uuid string) (*SvmInfo, error) {
	request := &SvmRequest{UUID: uuid}
	return svmGet(client, request)
}

func SvmGetByName(client *pythonapi.NetAppAPI, name string) (*SvmInfo, error) {
	request := &SvmRequest{Name: name}
	return svmGet(client, request)
}

func svmGet(client *pythonapi.NetAppAPI, request *SvmRequest) (*SvmInfo, error) {
	resp := SvmInfo{}
	err := pythonapi.MakeAPICall(client, svmGetCmd, request, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
