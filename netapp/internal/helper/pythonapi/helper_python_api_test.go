package pythonapi

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func rwTest(
	api *NetAppAPI, r *require.Assertions,
	key string, value string) {

	succ, errmsg, data, err := api.Call(key+":"+value, nil)
	r.NoError(err)
	r.True(succ)
	r.Zero(errmsg)
	r.NotZero(data)
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
