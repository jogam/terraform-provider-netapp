package netapp

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"user": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NETAPP_USER", nil),
				Description: "The user name for NetApp ONTAP API.",
			},

			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NETAPP_PASSWORD", nil),
				Description: "The user password for NetApp ONTAP API.",
			},

			"netapp_host": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NETAPP_HOST", nil),
				Description: "The NetApp host FQDN/IP for NetApp ONTAP API.",
			},
			// "allow_unverified_ssl": &schema.Schema {
			// 	Type:        schema.TypeBool,
			// 	Optional:    true,
			// 	DefaultFunc: schema.EnvDefaultFunc("VSPHERE_ALLOW_UNVERIFIED_SSL", false),
			// 	Description: "If set, VMware vSphere client will permit unverifiable SSL certificates.",
			// },
		},

		ResourcesMap: map[string]*schema.Resource{},

		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	return nil, nil
}
