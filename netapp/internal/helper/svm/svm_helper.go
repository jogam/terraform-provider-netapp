package svm

import (
	"fmt"

	"github.com/jogam/terraform-provider-netapp/netapp/internal/helper/pythonapi"
)

type ProtocolInfo struct {
	Type string `json:"type"`
}

type Request struct {
	Name          string `json:"name,omitempty"`
	NewName       string `json:"new_name,omitempty"`
	Force         string `json:"force,omitempty"`
	UUID          string `json:"uuid,omitempty"`
	IPSpace       string `json:"ipspace"`
	RootAggr      string `json:"root_aggr,omitempty"`
	RootSecStyle  string `json:"root_sec_style,omitempty"`
	RootName      string `json:"root_name,omitempty"`
	RootRetention string `json:"root_retent,omitempty"`
	//Protocols     []ProtocolInfo `json:"protocols,omitempty"` // also from SVM
}

type Info struct {
	pythonapi.ResourceInfo
	Request

	ConfigLocked bool   `json:"locked"`
	OperState    string `json:"oper_state"`
	SvmState     string `json:"svm_state"`

	ProtoEnabled  []string `json:"proto_enabled"`
	ProtoInactive []string `json:"proto_inactive"`
}

type JobResult struct {
	Info

	Status string `json:"status"` // <result-status>
	JobID  int    `json:"jobid"`  // <result-jobid>
	ErrNo  int    `json:"errno"`  // <result-error-code>
	ErrMsg string `json:"errmsg"` // <result-error-message>
}

//******************************************************************
// status section
//******************************************************************

const svmGetCmd = "SVM.GET"

func GetByUUID(client *pythonapi.NetAppAPI, uuid string) (*Info, error) {
	request := &Request{UUID: uuid}
	return svmGet(client, request)
}

func GetByName(client *pythonapi.NetAppAPI, name string) (*Info, error) {
	request := &Request{Name: name}
	return svmGet(client, request)
}

func svmGet(client *pythonapi.NetAppAPI, request *Request) (*Info, error) {
	resp := Info{}
	err := pythonapi.MakeAPICall(client, svmGetCmd, request, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

const svmCreateCmd = "SVM.CREATE"

func Create(client *pythonapi.NetAppAPI, request *Request) (*JobResult, error) {
	response := &JobResult{}
	err := pythonapi.MakeAPICall(client, svmCreateCmd, request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

const svmDeleteCmd = "SVM.DELETE"

func DeleteByName(client *pythonapi.NetAppAPI, name string) (*JobResult, error) {
	request := Request{Name: name}
	response := &JobResult{}
	err := pythonapi.MakeAPICall(client, svmDeleteCmd, &request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

type simpleSvmCmd string

const (
	// StartCmd start SVM
	StartCmd simpleSvmCmd = "SVM.START"
	// StopCmd stop svm
	StopCmd simpleSvmCmd = "SVM.STOP"
	// UnlockCmd unlock svm configuration
	UnlockCmd simpleSvmCmd = "SVM.UNLOCK"
)

// ExecuteSimpleCommand execute start/stop/unlock of SVM
func ExecuteSimpleCommand(
	client *pythonapi.NetAppAPI, name string,
	simpleCmd simpleSvmCmd, force bool) error {

	request := Request{Name: name}
	if force {
		request.Force = fmt.Sprintf("%v", true)
	}
	response := pythonapi.EmptyResponse{}
	return pythonapi.MakeAPICall(
		client, fmt.Sprintf("%s", simpleCmd), &request, &response)
}

const svmRenameCmd = "SVM.RENAME"

// Rename rename SVM
func Rename(client *pythonapi.NetAppAPI, name, newName string) error {
	request := Request{Name: name, NewName: newName}
	response := pythonapi.EmptyResponse{}
	return pythonapi.MakeAPICall(
		client, svmRenameCmd, &request, &response)
}

type InstanceRequest struct {
	SvmInstanceName string `json:"svm_name"` // the name of the SVM instance to execture the command at
}

type simpleVolCmd string

const (
	// VolumeOnlineCommand bring the SVM volume online
	VolumeOnlineCommand simpleVolCmd = "SVM.VOL.ONLINE"
	// VolumeOfflineCommand take the SVM volume offline
	VolumeOfflineCommand simpleVolCmd = "SVM.VOL.OFFLINE"
	// VolumeRestrictCommand make SVM volume inaccessible by users
	VolumeRestrictCommand simpleVolCmd = "SVM.VOL.RESTRICT"
	// VolumeDeleteCommand delete SVM colume
	VolumeDeleteCommand simpleVolCmd = "SVM.VOL.DELETE"
)

type VolumeRequest struct {
	InstanceRequest
	VolumeName string `json:"name"`           // the name of the SVM volume to work on
	Size       string `json:"size,omitempty"` // size of the volume 1m, 1g, 1t ...
}

type VolumeInfo struct {
	VolumeRequest
}

func VolumeSimpleCommand(
	client *pythonapi.NetAppAPI,
	svmName, volName string,
	simpleCmd simpleVolCmd) error {
	request := VolumeRequest{VolumeName: volName}
	request.SvmInstanceName = svmName
	response := pythonapi.EmptyResponse{}
	return pythonapi.MakeAPICall(
		client, fmt.Sprintf("%s", simpleCmd), &request, &response)
}

const svmVolumeSizeCmd = "SVM.VOL.SIZE"

func VolumeSizeCommand(
	client *pythonapi.NetAppAPI, request *VolumeRequest) (*VolumeInfo, error) {
	response := &VolumeInfo{}
	err := pythonapi.MakeAPICall(client, svmVolumeSizeCmd, request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
