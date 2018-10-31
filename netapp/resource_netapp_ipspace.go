package netapp

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	netappnw "github.com/jogam/terraform-provider-netapp/netapp/internal/helper/network"
)

func resourceNetAppIPSpace() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The name of the IPSpace.",
				Required:    true,
			},

			//******************************************************************
			// status section
			//******************************************************************

			"uuid": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The unique identifier for the IPSpace.",
				Computed:    true,
				ForceNew:    true,
			},

			"broadcast_domains": &schema.Schema{
				Type:        schema.TypeList,
				Description: "The broadcast domains using this IPSpace.",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"ports": &schema.Schema{
				Type:        schema.TypeList,
				Description: "The net-qualified ports belonging to this IPspace.",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"vservers": &schema.Schema{
				Type:        schema.TypeList,
				Description: "The vservers using this IPspace.",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},

		Create: resourceNetAppIPSpaceCreate,
		Read:   resourceNetAppIPSpaceRead,
		Update: resourceNetAppIPSpaceUpdate,
		Delete: resourceNetAppIPSpaceDelete,

		// as per: https://www.terraform.io/docs/extend/resources.html#importers
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceNetAppIPSpaceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api

	name := d.Get("name").(string)
	uuid, err := netappnw.IPSpaceCreate(client, name)
	if err != nil {
		return err
	}

	// write back the new ID
	d.SetId(uuid)

	// fill in the rest of the data
	return resourceNetAppIPSpaceRead(d, meta)
}

func resourceNetAppIPSpaceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api

	ipSpaceInfo, err := netappnw.IPSpaceGetByUUID(client, d.Id())
	if err != nil {
		log.Printf("[WARN] no IPSpace found for [%s], got: %s", d.Id(), err)
		d.SetId("")
		return nil
	}

	d.Set("name", ipSpaceInfo.Name)
	d.Set("uuid", ipSpaceInfo.UUID)
	if err = d.Set("broadcast_domains", stringArrayToInterfaceArray(ipSpaceInfo.BroadCastDomains)); err != nil {
		return fmt.Errorf("broadcast domain loading failed: %s", err)
	}
	if err = d.Set("ports", stringArrayToInterfaceArray(ipSpaceInfo.Ports)); err != nil {
		return fmt.Errorf("ports loading failed: %s", err)
	}
	if err = d.Set("vservers", stringArrayToInterfaceArray(ipSpaceInfo.VServers)); err != nil {
		return fmt.Errorf("vservers loading failed: %s", err)
	}

	d.SetId(ipSpaceInfo.UUID)

	return nil
}

func resourceNetAppIPSpaceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api

	ipSpaceInfo, err := netappnw.IPSpaceGetByUUID(client, d.Id())
	if err != nil {
		return fmt.Errorf("could not get IPSpace [%s] during update, got %s", d.Id(), err)
	}

	name := d.Get("name").(string)
	err = netappnw.IPSpaceUpdate(client, ipSpaceInfo.Name, name)
	if err != nil {
		return err
	}

	return resourceNetAppIPSpaceRead(d, meta)
}

func resourceNetAppIPSpaceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api

	ipSpaceInfo, err := netappnw.IPSpaceGetByUUID(client, d.Id())
	if err != nil {
		return fmt.Errorf("could not get IPSpace [%s] during delete, got %s", d.Id(), err)
	}
	return netappnw.IPSpaceDelete(client, ipSpaceInfo.Name)
}
