package pythonapi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func rwTest(
	api *NetAppAPI, r *require.Assertions,
	key string, value string) {

	err := api.Put(key, value)
	r.NoError(err)

	val, err := api.Get(key)
	r.NoError(err)

	r.Equalf(value, val, "read value should match")
}

func Test_Python_Api_Create(t *testing.T) {
	r := require.New(t)
	api, err := CreateAPI(tmpDir, "", "1234")

	r.NoError(err)

	rwTest(api, r, "testy", "testing")

	err = api.Stop()
	r.NoError(err)
}

func Test_Python_Api_MultiCreate(t *testing.T) {
	r := require.New(t)
	api, err := CreateAPI(tmpDir, "", "1234")
	r.NoError(err)

	rwTest(api, r, "testy-multi", "testing")

	// create more
	for i := 0; i < 5; i++ {
		addapi, err := CreateAPI(tmpDir, "", "1234")
		r.NoError(err)

		rwTest(addapi, r, "testy-multi-"+string(i), "testing-"+string(i))

		err = addapi.Stop()
		r.NoError(err)
	}

	err = api.Stop()
	r.NoError(err)
}
