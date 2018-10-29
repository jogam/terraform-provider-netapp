package netapp

import (
	"fmt"
	"reflect"

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

			"admin_up": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Change the port status to up if true (default: true)",
				Optional:    true,
			},

			"admin_mtu": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Configure MTU for this port (default: 1500).",
				Optional:    true,
			},

			"admin_auto": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Set duplex, speed, flow to autonegotiate on this port, e.g. ignore those parameter values (default: true).",
				Optional:    true,
			},

			"admin_speed": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The speed settings for this port, must be on of [undef, auto, 10, 100, (10,25,40,100)000].",
				Optional:    true,
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

			"admin_duplex": &schema.Schema{
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

			"admin_flow": &schema.Schema{
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

			"status_ipspace": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Current ports associated ipspace.",
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

			"status_autorevert_delay": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Current delay in seconds before auto-reverting a LIF to this cluster port.",
				Computed:    true,
			},

			"status_ignore_health": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Current ignore port health status on hosted logical interface.",
				Computed:    true,
			},

			"status_role": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Current  port role (deprecated in ONTAP 8.3)",
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

func writeToSchemaIfInCfg(d *schema.ResourceData, key string, value interface{}) error {
	_, isCfg := d.GetOk(key)
	if isCfg {
		if err := d.Set(key, value); err != nil {
			return err
		}
	}

	return nil
}

func writeToValueIfInCfg(d *schema.ResourceData, key string, value *reflect.Value) (bool, error) {
	cfgValue, isSet := d.GetOkExists(key)
	// changed check removed, does not seem to work...
	if isSet {
		v := value.Elem()
		switch v.Kind() {
		case reflect.Int:
			v.SetInt(int64(cfgValue.(int)))
		case reflect.String:
			v.SetString(cfgValue.(string))
		case reflect.Bool:
			v.SetBool(cfgValue.(bool)) // TODO: attempt to write to address NIL!!
		default:
			return false, fmt.Errorf("unsupported datatype: %s", v.Kind())
		}

		return true, nil
	}

	return false, nil
}

func resourceNetAppPortRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api
	nodeID := d.Get("node_id").(string)
	portName := d.Get("nic_name").(string)

	// get the NetApp node, e.g. confirm exists
	nodeInfo, err := netappsys.NodeGetByUUID(client, nodeID)
	if err != nil {
		return err
	}

	resp, err := netappsys.PortGetByNames(client, nodeInfo.Name, portName)
	if err != nil {
		return err
	}

	// write back admin parameters if they are set in resource data
	for key, value := range map[string]interface{}{
		"admin_up":         resp.AdminUp,
		"admin_mtu":        resp.AdminMtu,
		"autorevert_delay": resp.AutoRevertDelay,
		"ignore_health":    resp.IgnoreHealth,
		"ipspace":          resp.IPSpace,
		"role":             resp.Role} {
		if err = writeToSchemaIfInCfg(d, key, value); err != nil {
			return err
		}
	}

	cfgAdmin, isCfgAdmin := d.GetOk("admin_auto")
	if isCfgAdmin {
		// user cares about auto-negotiation
		cfgAdminAuto := cfgAdmin.(bool)
		if err = d.Set("admin_auto", resp.AdminAuto); err != nil {
			return err
		}

		if !cfgAdminAuto && !resp.AdminAuto {
			// autonegotiation true in neither requested nor status
			// read back the admin settings from response
			for key, value := range map[string]interface{}{
				"admin_speed":  resp.AdminSpeed,
				"admin_duplex": resp.AdminDuplex,
				"admin_flow":   resp.AdminFlow} {
				if err = d.Set(key, value); err != nil {
					return err
				}
			}
		}
	}

	// status write back - computed/info params, e.g. possibly not specified by user
	for key, value := range map[string]interface{}{
		"status":                  resp.Status,
		"health":                  resp.Health,
		"mac_address":             resp.Mac,
		"broadcast_domain":        resp.BroadCastDomain,
		"status_ipspace":          resp.IPSpace,
		"status_mtu":              resp.Mtu,
		"status_auto":             resp.Auto,
		"status_speed":            resp.Speed,
		"status_duplex":           resp.Duplex,
		"status_flow":             resp.Flow,
		"status_autorevert_delay": resp.AutoRevertDelay,
		"status_ignore_health":    resp.IgnoreHealth,
		"status_role":             resp.Role} {
		if err = d.Set(key, value); err != nil {
			return err
		}
	}

	d.SetId(createPortID(resp))

	return nil
}

func resourceNetAppPortUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api
	nodeID := d.Get("node_id").(string)
	portName := d.Get("nic_name").(string)

	// get the NetApp node, e.g. confirm exists
	nodeInfo, err := netappsys.NodeGetByUUID(client, nodeID)
	if err != nil {
		return err
	}

	req := netappsys.PortModifyRequest{}
	req.NodeName = nodeInfo.Name
	req.PortName = portName

	changeCnt := 0

	for key, value := range map[string]reflect.Value{
		"admin_up":         reflect.ValueOf(req.Up),
		"admin_mtu":        reflect.ValueOf(&req.Mtu),
		"autorevert_delay": reflect.ValueOf(&req.AutoRevertDelay),
		"ignore_health":    reflect.ValueOf(req.IgnoreHealth),
		"ipspace":          reflect.ValueOf(&req.IPSpace),
		"role":             reflect.ValueOf(&req.Role)} {
		written, err := writeToValueIfInCfg(d, key, &value)
		if err != nil {
			return err
		}

		if written {
			changeCnt++
		}
	}

	cfgValue, isSet := d.GetOk("admin_auto")
	isAuto := false
	if isSet {
		isAuto = cfgValue.(bool)
		if d.HasChange("admin_auto") { // probably does not work either
			req.Auto = &isAuto
			changeCnt++
		}
	}

	if !isAuto {
		for key, value := range map[string]reflect.Value{
			"admin_speed":  reflect.ValueOf(&req.Speed),
			"admin_duplex": reflect.ValueOf(&req.Duplex),
			"admin_flow":   reflect.ValueOf(&req.Flow)} {
			written, err := writeToValueIfInCfg(d, key, &value)
			if err != nil {
				return err
			}

			if written {
				changeCnt++
			}
		}
	}

	if changeCnt > 0 {
		err = netappsys.PortModify(client, &req)
		if err != nil {
			return err
		}
	}

	return resourceNetAppPortRead(d, meta)
}

func resourceNetAppPortDelete(d *schema.ResourceData, meta interface{}) error {
	// can not delete hardware, e.g. do nothing
	return nil
}
