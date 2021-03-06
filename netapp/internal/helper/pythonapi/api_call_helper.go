package pythonapi

import (
	"encoding/json"
	"fmt"
	"log"
)

// EmptyResponse for API calls without return value
type EmptyResponse struct {
	Dummy int `json:"dummy"`
}

// ResourceInfo is base struct for NetApp resource Info's returned
type ResourceInfo struct {
	NonExist bool `json:"non_exist,omitempty"` // flag to indicate that resource does not exist
}

// MakeAPICall realizes the Marshall/Unmarshall and actual API call
func MakeAPICall(
	client *NetAppAPI, cmdName string,
	request, response interface{}) error {

	byteReq, err := json.Marshal(request)
	if err != nil {
		log.Printf(
			"[ERROR] could not marshal api call [%s] request [%v], got: %s",
			cmdName, request, err)
		return fmt.Errorf("api call [%s] request marshal error: %s", cmdName, err)
	}
	succ, errmsg, data, err := client.Call(cmdName, byteReq)
	if err != nil {
		log.Printf("[ERROR] could not execute API call [%s], got: %s", cmdName, err)
		return err
	}

	if !succ || errmsg != "" {
		log.Printf("[WARN] api call [%s] not successful got: %v", cmdName, errmsg)
		return fmt.Errorf("api call [%s] failed with msg: %v", cmdName, errmsg)
	}

	err = json.Unmarshal(data, response)
	if err != nil {
		log.Printf(
			"[ERROR] could not unmarshal api call [%s] response [%v], got: %s",
			cmdName, data, err)
		return fmt.Errorf("api call [%s] request unmarshal error: %s", cmdName, err)
	}

	return nil
}
