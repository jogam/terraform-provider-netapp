package pythonapi

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

const testKeyValueCmd = "TEST.KEYVALUE"

// KeyValueRequest is the required input for the Testing Call
type KeyValueRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Write bool   `json:"write"`
}

// KeyValueResponse is the returned result for a test call
type KeyValueResponse struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	Modified bool   `json:"modified"`
}

// TestKeyValue executes a KeyValue API test call
func testKeyValue(client *NetAppAPI, request *KeyValueRequest) (*KeyValueResponse, error) {
	resp := KeyValueResponse{}
	err := MakeAPICall(client, testKeyValueCmd, request, &resp)
	return &resp, err
}

func rwTest(
	api *NetAppAPI, r *require.Assertions,
	key string, value string) {
	resp, err := testKeyValue(api, &KeyValueRequest{
		Key: key, Value: value, Write: true,
	})
	r.NoError(err)
	r.Equal(key, resp.Key)
	r.Equal(value, resp.Value)
	r.True(resp.Modified)

	resp, err = testKeyValue(api, &KeyValueRequest{
		Key: key, Value: "", Write: false,
	})
	r.NoError(err)
	r.Equal(key, resp.Key)
	r.Equal(value, resp.Value)
	r.False(resp.Modified)
}

func Test_Python_Api_Create(t *testing.T) {
	r := require.New(t)
	api, err := CreateAPI(
		tmpDir, "/home/gmueller/software/netapp/netapp-manageability-sdk-9.4",
		"56789", "1234")

	r.NoError(err)

	rwTest(api, r, "testy", "testing")

	err = api.Stop()
	r.NoError(err)
}

func Test_Python_Api_MultiCreate(t *testing.T) {
	r := require.New(t)
	api, err := CreateAPI(
		tmpDir, "/home/gmueller/software/netapp/netapp-manageability-sdk-9.4",
		"56789", "1234")
	r.NoError(err)

	rwTest(api, r, "testy-multi", "testing")

	// create more
	for i := 1; i < 6; i++ {
		addapi, err := CreateAPI(
			tmpDir, "/home/gmueller/software/netapp/netapp-manageability-sdk-9.4",
			"56789", "1234")
		r.NoError(err)

		rwTest(addapi, r, fmt.Sprintf("testy-multi-%d", i), fmt.Sprintf("testing-%d", i))

		// err = addapi.Stop()
		// r.NoError(err)
	}

	err = api.Stop()
	r.NoError(err)
}
