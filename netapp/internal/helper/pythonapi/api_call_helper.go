package pythonapi

import (
	"encoding/json"
	"fmt"
	"log"
)

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
