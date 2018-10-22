package netapp

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jogam/terraform-provider-netapp/netapp/internal/helper/pythonapi"
)

type NetAppClient struct {
	Api *pythonapi.NetAppAPI
}

type Config struct {
	User     string
	Password string
	Host     string
	SdkRoot  string
	ApiPath  string
	ApiPort  string
	RegPort  string
}

// NewConfig returns a new Config from the supplied ResourceData
func NewConfig(d *schema.ResourceData) (*Config, error) {
	c := &Config{
		User:     d.Get("user").(string),
		Password: d.Get("password").(string),
		Host:     d.Get("host").(string),
		SdkRoot:  d.Get("nmsdk_root_path").(string),
		ApiPath:  d.Get("api_folder").(string),
		ApiPort:  d.Get("api_port").(string),
		RegPort:  d.Get("api_client_registry_port").(string),
	}

	return c, nil
}

func (c *Config) Client() (*NetAppClient, error) {
	api, err := pythonapi.CreateAPI(
		c.ApiPath, c.SdkRoot,
		c.RegPort, c.ApiPort)
	if err != nil {
		return nil, fmt.Errorf("Error creating python NetApp API: %s", err)
	}

	client := &NetAppClient{
		Api: api,
	}

	return client, nil
}
