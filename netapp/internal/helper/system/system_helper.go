package system

import (
	"encoding/json"
	"fmt"
	"log"

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
	byteReq, err := json.Marshal(request)
	if err != nil {
		log.Printf(
			"[ERROR] could not marshal connect request [%v], got: %s",
			request, err)
		return nil, fmt.Errorf("connect request marshal error: %s", err)
	}
	succ, errmsg, data, err := client.Call(connectCmd, byteReq)
	if err != nil {
		log.Printf("[ERROR] could not execute API connect call, got: %s", err)
		return nil, err
	}

	if !succ || errmsg != "" {
		log.Printf("[WARN] api connect call not successful got: %v", errmsg)
		return nil, fmt.Errorf("api connect call failed with msg: %v", errmsg)
	}

	resp := ConnectResponse{}
	err = json.Unmarshal(data, &resp)
	if err != nil {
		log.Printf(
			"[ERROR] could not unmarshal api connect response [%v], got: %s",
			data, err)
		return nil, fmt.Errorf("connect request unmarshal error: %s", err)
	}

	return &resp, nil
}
