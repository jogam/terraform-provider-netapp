package netapp

import (
	"os"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
)

func testAccClientGenerateConfig(t *testing.T) *Config {

	return &Config{
		User:     os.Getenv("NETAPP_USER"),
		Password: os.Getenv("NETAPP_PASSWORD"),
		Host:     os.Getenv("NETAPP_HOST"),
		SdkRoot:  os.Getenv("NETAPP_MSDK_ROOT_PATH"),
		ApiPath:  os.Getenv("NETAPP_API_FOLDER"),
	}
}
func TestNewConfig(t *testing.T) {
	expected := &Config{
		User:     "foo",
		Password: "bar",
		Host:     "cookie",
		SdkRoot:  "rootPath",
		ApiPath:  "apiFldr",
	}

	r := &schema.Resource{Schema: Provider().(*schema.Provider).Schema}
	d := r.Data(nil)
	d.Set("user", expected.User)
	d.Set("password", expected.Password)
	d.Set("host", expected.Host)
	d.Set("nmsdk_root_path", expected.SdkRoot)
	d.Set("api_folder", expected.ApiPath)

	actual, err := NewConfig(d)
	if err != nil {
		t.Fatalf("error creating new configuration: %s", err)
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected %#v, got %#v", expected, actual)
	}
}
