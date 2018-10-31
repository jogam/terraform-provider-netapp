package netapp

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	netappnw "github.com/jogam/terraform-provider-netapp/netapp/internal/helper/network"
	netappsys "github.com/jogam/terraform-provider-netapp/netapp/internal/helper/system"
)

func resourceNetAppVlan() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"parent_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The managed object ID of the port/interface this VLAN belongs to.",
				Required:    true,
				ForceNew:    true,
			},

			"vlan_id": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Identified for this VLAN [1..4094]",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(int)
					if v < 1 || v > 4094 {
						errs = append(errs, fmt.Errorf(
							"%q must be between 1..4094, was: %v", key, v))
					}
					return
				},
			},

			//******************************************************************
			// status section
			//******************************************************************

			"net_qualified_name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The qualified port/interface/vlan name for NetApp broadcast domain,etc.",
				Computed:    true,
			},

			"node_name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The node name this VLAN is defined for.",
				Computed:    true,
			},

			"parent_name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The parent interface/port name for this VLAN.",
				Computed:    true,
			},

			"interface_name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The VLAN interface name.",
				Computed:    true,
			},
		},

		Create: resourceNetAppVlanCreate,
		Read:   resourceNetAppVlanRead,
		Delete: resourceNetAppVlanDelete,

		// as per: https://www.terraform.io/docs/extend/resources.html#importers
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func createVlanRequestFromParentVlanID(
	meta interface{},
	parentID string, vlanID int) (*netappnw.VlanRequest, error) {

	nodeName, portName, err := getNodePortNameFromPortID(parentID)
	if err != nil {
		return nil, fmt.Errorf(
			"currently only node ports supported for VLAN creation, got ID: %s",
			parentID)
	}

	client := meta.(*NetAppClient).api
	pgResp, err := netappsys.PortGetByNames(client, nodeName, portName)
	if err != nil {
		return nil, fmt.Errorf("could not read VLAN parent port node: %s", err)
	}

	req := netappnw.VlanRequest{
		NodeName:   pgResp.NodeName,
		ParentName: pgResp.PortName,
		VlanID:     strconv.Itoa(vlanID),
	}
	return &req, nil
}

func resourceNetAppVlanCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api

	parentID := d.Get("parent_id").(string)
	vlanID := d.Get("vlan_id").(int)

	req, err := createVlanRequestFromParentVlanID(meta, parentID, vlanID)
	if err != nil {
		return err
	}
	err = netappnw.VlanCreate(client, req)
	if err != nil {
		if strings.Contains(err.Error(), "reason=\"duplicate entry\"") {
			// duplicate vlan, hint at import feature
			resID := createVlanID(req)
			return fmt.Errorf(
				"vlan [%v] on port [%s] already exists, import via cmd: "+
					"terraform import $RESNAME$ '%s'",
				vlanID, parentID, resID)
		}

		return fmt.Errorf(
			"vlan ID [%v] on port [%s] create, got: %s",
			vlanID, parentID, err)
	}

	return resourceNetAppVlanRead(d, meta)
}

func resourceNetAppVlanRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api

	parentID := d.Get("parent_id").(string)
	vlanID := d.Get("vlan_id").(int)

	var req *netappnw.VlanRequest
	var err error
	if parentID != "" && vlanID != 0 {
		// normal resource read
		req, err = createVlanRequestFromParentVlanID(meta, parentID, vlanID)
		if err != nil {
			return err
		}
	} else {
		// resource import, attempt get required info from ID
		nodeName, portName, vlanID, err := getNodePortNameVlanFromVlanID(d.Id())
		if err != nil {
			return err
		}

		req = &netappnw.VlanRequest{
			NodeName:   nodeName,
			ParentName: portName,
			VlanID:     strconv.Itoa(vlanID),
		}

		// write back vlan ID to resource
		d.Set("vlan_id", vlanID)
	}

	vgResp, err := netappnw.VlanGet(client, req)
	if err != nil {
		log.Printf(
			"vlan ID [%v] on port [%s] read, got: %s",
			vlanID, parentID, err)
		d.SetId("")
		return nil
	}

	if parentID == "" {
		// we are importing, get parent ID
		pInfo, err := netappsys.PortGetByNames(
			client, vgResp.NodeName, vgResp.ParentName)
		if err != nil {
			return fmt.Errorf(
				"could not get port info for parent port during vlan [%s] import, got: %s",
				d.Id(), err)
		}

		parentID = createPortID(pInfo)

		// write back parent port ID to resource
		d.Set("parent_id", parentID)
	}

	d.Set("parent_name", vgResp.ParentName)
	d.Set("node_name", vgResp.NodeName)
	d.Set("interface_name", vgResp.Name)
	d.Set("net_qualified_name", createNetQualifiedName(vgResp))

	d.SetId(createVlanID(vgResp))

	return nil
}

func resourceNetAppVlanDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api

	parentID := d.Get("parent_id").(string)
	vlanID := d.Get("vlan_id").(int)

	req, err := createVlanRequestFromParentVlanID(meta, parentID, vlanID)
	if err != nil {
		return err
	}
	err = netappnw.VlanDelete(client, req)
	if err != nil {
		return fmt.Errorf(
			"vlan ID [%v] on port [%s] delete, got: %s",
			vlanID, parentID, err)
	}

	return nil
}
