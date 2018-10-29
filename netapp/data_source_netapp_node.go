package netapp

import (
	"github.com/hashicorp/terraform/helper/schema"
	netappsys "github.com/jogam/terraform-provider-netapp/netapp/internal/helper/system"
)

func dataSourceNetAppNode() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNetAppNodeRead,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The name of the node. If not provided assumed to be single node / not cluster.",
				Required:    true,
			},

			"serial_number": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The serial number of the node.",
				Computed:    true,
			},

			"system_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The system ID of the node.",
				Computed:    true,
			},

			"version": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The product version of the node.",
				Computed:    true,
			},

			"healthy": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "The health status of node.",
				Computed:    true,
			},

			"uptime": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "The uptime of the node [s].",
				Computed:    true,
			},
		},
	}
}

func dataSourceNetAppNodeRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api
	nodeName := d.Get("name").(string)
	resp, err := netappsys.NodeGetByName(client, nodeName)
	if err != nil {
		return err
	}

	if err = d.Set("name", resp.Name); err != nil {
		return err
	}
	if err = d.Set("serial_number", resp.Serial); err != nil {
		return err
	}
	if err = d.Set("system_id", resp.ID); err != nil {
		return err
	}
	if err = d.Set("version", resp.Version); err != nil {
		return err
	}
	if err = d.Set("healthy", resp.Healty); err != nil {
		return err
	}
	if err = d.Set("uptime", resp.Uptime); err != nil {
		return err
	}

	d.SetId(resp.UUID)

	return nil
}
