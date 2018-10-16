package pythonapi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Python_Api_Create(t *testing.T) {
	r := require.New(t)
	api, err := CreateAPI(tmpDir, "")

	r.NoError(err)

	err = api.Put("testy", "testing")
	r.NoError(err)

	val, err := api.Get("testy")
	r.NoError(err)

	r.Equalf("testing", val, "read value should match")

	err = api.Stop()
	r.NoError(err)
}
