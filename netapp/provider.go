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

			"host": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NETAPP_HOST", nil),
				Description: "The NetApp host FQDN/IP for NetApp ONTAP API.",
			},

			"nmsdk_root_path": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NETAPP_MSDK_ROOT_PATH", nil),
				Description: "The path to the NetApp Manageability SDK root folder",
			},

			"api_folder": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NETAPP_API_FOLDER", nil),
				Description: "Path to folder where the NetApp api should be unpacked.",
			},

			// "allow_unverified_ssl": &schema.Schema {
			// 	Type:        schema.TypeBool,
			// 	Optional:    true,
			// 	DefaultFunc: schema.EnvDefaultFunc("VSPHERE_ALLOW_UNVERIFIED_SSL", false),
			// 	Description: "If set, VMware vSphere client will permit unverifiable SSL certificates.",
			// },
		},

		ResourcesMap: map[string]*schema.Resource{
			"netapp_key_value": resourceNetappKeyValue(),
		},

		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	c, err := NewConfig(d)
	if err != nil {
		return nil, err
	}

	return c, nil
}
