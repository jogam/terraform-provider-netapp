package netapp

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jogam/terraform-provider-netapp/netapp/internal/helper/pythonapi"

	netappsys "github.com/jogam/terraform-provider-netapp/netapp/internal/helper/system"
)

type NetAppClient struct {
	api *pythonapi.NetAppAPI
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

func (c *Config) savedOrNewApiSession() (*pythonapi.NetAppAPI, bool, error) {
	// TODO: look into saving conn info via ReattachConfig from go-plugin client
	api, err := pythonapi.CreateAPI(
		c.ApiPath, c.SdkRoot,
		c.RegPort, c.ApiPort)
	if err != nil {
		return nil, false, fmt.Errorf("Error creating python NetApp API: %s", err)
	}

	return api, false, nil
}

func (c *Config) connectToAPI(client *NetAppClient, d *schema.ResourceData) error {
	// connect and get the ONTAP/OS version
	resp, err := netappsys.Connect(
		client.api, &netappsys.ConnectRequest{
			Host: c.Host, User: c.User, Password: c.Password,
		})

	if err != nil {
		return err
	}

	d.Set("ontap_version", resp.OntapVersion)
	d.Set("os_version", resp.OsVersion)

	return nil
}

func (c *Config) Client(d *schema.ResourceData) (*NetAppClient, error) {
	client := new(NetAppClient)

	var saved bool
	var err error

	client.api, saved, err = c.savedOrNewApiSession()
	if err != nil {
		return nil, err
	}

	if !saved {
		err := c.connectToAPI(client, d)
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}
