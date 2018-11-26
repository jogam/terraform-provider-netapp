package netapp

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	netappsys "github.com/jogam/terraform-provider-netapp/netapp/internal/helper/system"
)

func dataSourceNetAppAggr() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNetAppAggrRead,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The name of the aggregation.",
				Required:    true,
				ForceNew:    true,
			},

			"node_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The managed object ID of the node the aggregate belongs to.",
				Required:    true,
				ForceNew:    true,
			},

			"flexvol_count": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Number of flex volume on this aggregate.",
				Computed:    true,
			},

			"percent_used_capacity": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Used capacity in %.",
				Computed:    true,
			},

			"percent_used_physical": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Physical space used in %.",
				Computed:    true,
			},

			"size_total": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Total size of aggregate in Byte.",
				Computed:    true,
			},

			"size_used": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Used size of aggregate in Byte.",
				Computed:    true,
			},

			"size_available": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Available size of aggregate in Byte.",
				Computed:    true,
			},

			"size_reserved": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Reserved size of aggregate in Byte.",
				Computed:    true,
			},
		},
	}
}

func dataSourceNetAppAggrRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api
	aggrName := d.Get("name").(string)
	nodeID := d.Get("node_id").(string)

	// get the NetApp node, e.g. confirm exists
	nodeInfo, err := netappsys.NodeGetByUUID(client, nodeID)
	if err != nil {
		return fmt.Errorf("could not find node [%s], got: %s", nodeID, err)
	}

	aggrInfo, err := netappsys.AggrGetByName(client, aggrName)
	if err != nil {
		return fmt.Errorf(
			"unable for find aggregate [%s] on node [%s], got: %s",
			aggrName, nodeInfo.Name, err)
	}

	if aggrInfo.NonExist {
		d.SetId("")
		return nil
	}

	d.Set("name", aggrInfo.Name)
	d.Set("flexvol_count", aggrInfo.FlexVolCount)
	d.Set("percent_used_capacity", aggrInfo.PctUsedCapacity)
	d.Set("percent_used_physical", aggrInfo.PctUsedPhysical)
	d.Set("size_total", aggrInfo.SizeTotal)
	d.Set("size_used", aggrInfo.SizeUsed)
	d.Set("size_available", aggrInfo.SizeAvailable)
	d.Set("size_reserved", aggrInfo.SizeReserved)

	d.SetId(aggrInfo.UUID)

	return nil
}
