package pythonapi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SynchBoxToOS(t *testing.T) {
	r := require.New(t)
	res, err := SynchBoxToOS(tmpDir, &requiredAPIScripts)
	r.NoError(err)
	r.Equal(9, res.FileCount())

	for _, fRes := range res.files {
		r.True(fRes.Updated && fRes.Available)
	}
}
