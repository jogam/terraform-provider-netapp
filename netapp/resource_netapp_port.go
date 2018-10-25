package netapp

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	netappsys "github.com/jogam/terraform-provider-netapp/netapp/internal/helper/system"
)

func resourceNetAppPort() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"node_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The managed object ID of the node to set the port up on.",
				Required:    true,
				ForceNew:    true,
			},

			"nic_name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The name of the nic on the device (usually static).",
				Required:    true,
				ForceNew:    true,
			},

			"up": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Change the port status to up if true",
				Optional:    true,
			},

			"mtu": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Configure MTU for this port.",
				Optional:    true,
			},

			"auto": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Set duplex, speed, flow to autonegotiate on this port (true).",
				Optional:    true,
				Default:     true,
			},

			"speed": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The speed settings for this port, must be on of [undef, auto, 10, 100, (10,25,40,100)000].",
				Optional:    true,
				Default:     "auto",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					switch v {
					case
						"undef", "auto",
						"10", "100", "1000", "10000",
						"25000", "40000", "100000":
						return
					}

					errs = append(errs, fmt.Errorf("%q must be one of [undef, auto, half, full]", key))
					return
				},
			},

			"duplex": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The duplex settings for this port, must be on of [undef, auto, half, full].",
				Optional:    true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					switch v {
					case "undef", "auto", "half", "full":
						return
					}

					errs = append(errs, fmt.Errorf("%q must be one of [undef, auto, half, full]", key))
					return
				},
			},

			"flow": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The flow settings for this port, unchecked but seem to be on of [none, full].",
				Optional:    true,
			},

			"autorevert_delay": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "For a cluster port, configure the delay in seconds before auto-reverting a LIF to this port.",
				Optional:    true,
			},

			"ignore_health": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Ignore port health status on hosted logical interface (true).",
				Optional:    true,
			},
			"ipspace": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Specify the ports associated ipspace <-- should probably not be set here...",
				Optional:    true,
			},

			"role": &schema.Schema{
				Type:        schema.TypeString,
				Description: "specify the port role (deprecated in ONTAP 8.3 --> use ipspace)",
				Optional:    true,
			},

			//******************************************************************
			// status section
			//******************************************************************

			"status": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The current port status.",
				Computed:    true,
			},

			"health": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Current health status for this port.",
				Computed:    true,
			},

			"mac_address": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The ports MAC address.",
				Computed:    true,
			},

			"broadcast_domain": &schema.Schema{
				Type:        schema.TypeString,
				Description: "ports associated broadcast domain.",
				Computed:    true,
			},

			"status_mtu": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Current MTU for this port.",
				Computed:    true,
			},

			"status_auto": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Current autonegotiate status on this port.",
				Computed:    true,
			},

			"status_speed": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Current speed for this port.",
				Computed:    true,
			},

			"status_duplex": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Current duplex status for this port.",
				Computed:    true,
			},

			"status_flow": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Current flow status for this port.",
				Computed:    true,
			},
		},

		Create: resourceNetAppPortCreate,
		Read:   resourceNetAppPortRead,
		Update: resourceNetAppPortUpdate,
		Delete: resourceNetAppPortDelete,
	}
}

func resourceNetAppPortCreate(d *schema.ResourceData, meta interface{}) error {
	// NetApp ports are not neccesarily created, we just confirm they exist
	// in reality this an update call
	return resourceNetAppPortUpdate(d, meta)
}

func resourceNetAppPortRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api
	nodeID := d.Get("node_id").(string)
	portName := d.Get("nic_name").(string)

	// get the NetApp node, e.g. confirm exists
	nodeInfo, err := netappsys.GetNodeByUUID(client, nodeID)
	if err != nil {
		return err
	}

	resp, err := netappsys.GetPortByNames(client, nodeInfo.Name, portName)
	if err != nil {
		return err
	}
	//
	if err = d.Set("up", resp.AdminUp); err != nil {
		return err
	}
	if err = d.Set("mtu", resp.AdminMtu); err != nil {
		return err
	}
	if err = d.Set("auto", resp.AdminAuto); err != nil {
		return err
	}
	if err = d.Set("speed", resp.AdminSpeed); err != nil {
		return err
	}
	if err = d.Set("duplex", resp.AdminSpeed); err != nil {
		return err
	}
	if err = d.Set("flow", resp.AdminFlow); err != nil {
		return err
	}
	//
	if err = d.Set("autorevert_delay", resp.AutoRevertDelay); err != nil {
		return err
	}
	if err = d.Set("ignore_health", resp.IgnoreHealth); err != nil {
		return err
	}
	if err = d.Set("ipspace", resp.IPSpace); err != nil {
		return err
	}
	if err = d.Set("role", resp.Role); err != nil {
		return err
	}

	// status write back
	if err = d.Set("status", resp.Status); err != nil {
		return err
	}
	if err = d.Set("status_mtu", resp.Mtu); err != nil {
		return err
	}
	if err = d.Set("status_auto", resp.Auto); err != nil {
		return err
	}
	if err = d.Set("status_speed", resp.Speed); err != nil {
		return err
	}
	if err = d.Set("status_duplex", resp.Duplex); err != nil {
		return err
	}
	if err = d.Set("status_flow", resp.Flow); err != nil {
		return err
	}

	d.SetId(createPortID(resp))

	return nil
}

func resourceNetAppPortUpdate(d *schema.ResourceData, meta interface{}) error {

	return resourceNetAppPortRead(d, meta)
}

func resourceNetAppPortDelete(d *schema.ResourceData, meta interface{}) error {
	// can not delete hardware, e.g. do nothing
	return nil
}
