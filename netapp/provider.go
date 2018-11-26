package netapp

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider returns a terraform.ResourceProvider
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

			"api_port": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NETAPP_API_PORT", "12343"),
				Description: "Port on which the NetApp api should be started (Default: 12343).",
			},

			"api_client_registry_port": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NETAPP_API_CR_PORT", "12342"),
				Description: "Port on which the NetApp api client registry should be started (Default: 12342).",
			},

			// "ontap_version": &schema.Schema{
			// 	Type:        schema.TypeString,
			// 	Computed:    true,
			// 	Description: "The ONTAP version installed on the NetApp host",
			// },

			// "os_version": &schema.Schema{
			// 	Type:        schema.TypeString,
			// 	Computed:    true,
			// 	Description: "The system OS version installed on the NetApp host",
			// },

			// "allow_unverified_ssl": &schema.Schema {
			// 	Type:        schema.TypeBool,
			// 	Optional:    true,
			// 	DefaultFunc: schema.EnvDefaultFunc("VSPHERE_ALLOW_UNVERIFIED_SSL", false),
			// 	Description: "If set, VMware vSphere client will permit unverifiable SSL certificates.",
			// },
		},

		ResourcesMap: map[string]*schema.Resource{
			"netapp_port":            resourceNetAppPort(),
			"netapp_vlan":            resourceNetAppVlan(),
			"netapp_ipspace":         resourceNetAppIPSpace(),
			"netapp_broadcastdomain": resourceNetAppBroadcastDomain(),
			"netapp_subnet":          resourceNetAppSubnet(),
			"netapp_svm":             resourceNetAppSVM(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"netapp_node": dataSourceNetAppNode(),
			"netapp_aggr": dataSourceNetAppAggr(),
		},

		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	c, err := NewConfig(d)
	if err != nil {
		return nil, err
	}

	// need to pass resource data to client for computed ONTAP/OS
	return c.Client(d)
}
