package pythonapi

import (
	"os"
	"testing"
)

var tmpDir = "/tmp/python-api"

func setup() {
	// nothing here
}

func tearDown() {
	//DirDeleteRecursive(tmpDir)
}

func TestMain(m *testing.M) {
	setup()
	res := m.Run()
	tearDown()
	os.Exit(res)
}
