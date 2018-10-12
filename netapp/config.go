package netapp

import (
	"github.com/hashicorp/terraform/helper/schema"
)

type Config struct {
	User     string
	Password string
	Host     string
	SdkRoot  string
}

// NewConfig returns a new Config from the supplied ResourceData
func NewConfig(d *schema.ResourceData) (*Config, error) {
	c := &Config{
		User:     d.Get("user").(string),
		Password: d.Get("password").(string),
		Host:     d.Get("host").(string),
		SdkRoot:  d.Get("nmsdk_root_path").(string),
	}

	return c, nil
}
