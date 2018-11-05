package netapp

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

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
				Description: "Change the port status to up if true.",
				Optional:    true,
			},

			"admin_mtu": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Configure MTU for this port (default: 1500).",
				Optional:    true,
				Default:     1500,
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

type ParamDefinition struct {
	Value *string
	Kind  reflect.Kind
}

func writeToSchema(
	d *schema.ResourceData,
	key string, param ParamDefinition) error {
	value := *param.Value
	if len(strings.TrimSpace(value)) == 0 {
		// no value present, do nothing
		return nil
	}

	var newVal interface{}
	var err error

	switch param.Kind {
	case reflect.Int:
		newVal, err = strconv.Atoi(value)
	case reflect.String:
		newVal = value
	case reflect.Bool:
		newVal, err = strconv.ParseBool(value)
	default:
		return fmt.Errorf("unsupported datatype: %s", param.Kind)
	}

	if err != nil {
		return fmt.Errorf(
			"could not convert [%s = %s], error: %s",
			key, value, err)
	}
	return d.Set(key, newVal)
}

func writeToSchemaIfInCfg(
	d *schema.ResourceData,
	key string, param ParamDefinition) error {
	_, isCfg := d.GetOk(key)
	if isCfg {
		return writeToSchema(d, key, param)
	}

	return nil
}

func writeToValueIfInCfg(
	d *schema.ResourceData,
	key string, param ParamDefinition) (bool, error) {
	cfgValue, isSet := d.GetOkExists(key)
	// changed check removed, does not seem to work...
	if isSet {
		dest := param.Value
		switch param.Kind {
		case reflect.Int:
			intVal := cfgValue.(int)
			*dest = strconv.Itoa(intVal)
		case reflect.String:
			*dest = cfgValue.(string)
		case reflect.Bool:
			boolVal := cfgValue.(bool)
			*dest = strconv.FormatBool(boolVal)
		default:
			return false, fmt.Errorf("unsupported datatype: %s", param.Kind)
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

	pInfo, err := netappsys.PortGetByNames(client, nodeInfo.Name, portName)
	if err != nil {
		return err
	}

	if pInfo.NonExist {
		d.SetId("")
		return nil
	}

	// write back admin parameters if they are set in resource data
	for key, param := range map[string]ParamDefinition{
		"admin_up":         ParamDefinition{&pInfo.AdminUp, reflect.Bool},
		"admin_mtu":        ParamDefinition{&pInfo.AdminMtu, reflect.Int},
		"autorevert_delay": ParamDefinition{&pInfo.AutoRevertDelay, reflect.Int},
		"ignore_health":    ParamDefinition{&pInfo.IgnoreHealth, reflect.Bool},
		"ipspace":          ParamDefinition{&pInfo.IPSpace, reflect.String},
		"role":             ParamDefinition{&pInfo.Role, reflect.String}} {
		if err = writeToSchemaIfInCfg(d, key, param); err != nil {
			return err
		}
	}

	cfgAdmin, isCfgAdmin := d.GetOk("admin_auto")
	if isCfgAdmin {
		// user cares about auto-negotiation
		cfgAdminAuto := cfgAdmin.(bool)

		runAdminAuto, err := strconv.ParseBool(pInfo.AdminAuto)
		if err != nil {
			return fmt.Errorf("auto-negotiation configured but not reported")
		}

		if err = d.Set("admin_auto", runAdminAuto); err != nil {
			return err
		}

		if !cfgAdminAuto && !runAdminAuto {
			// autonegotiation true in neither requested nor status
			// read back the admin settings from response
			for key, param := range map[string]ParamDefinition{
				"admin_speed":  ParamDefinition{&pInfo.AdminSpeed, reflect.String},
				"admin_duplex": ParamDefinition{&pInfo.AdminDuplex, reflect.String},
				"admin_flow":   ParamDefinition{&pInfo.AdminFlow, reflect.String}} {
				if err = writeToSchema(d, key, param); err != nil {
					return err
				}
			}
		}
	}

	// status write back - computed/info params, e.g. possibly not specified by user
	for key, param := range map[string]ParamDefinition{
		"status":                  ParamDefinition{&pInfo.Status, reflect.String},
		"health":                  ParamDefinition{&pInfo.Health, reflect.String},
		"mac_address":             ParamDefinition{&pInfo.Mac, reflect.String},
		"broadcast_domain":        ParamDefinition{&pInfo.BroadCastDomain, reflect.String},
		"status_ipspace":          ParamDefinition{&pInfo.IPSpace, reflect.String},
		"status_mtu":              ParamDefinition{&pInfo.Mtu, reflect.Int},
		"status_auto":             ParamDefinition{&pInfo.Auto, reflect.Bool},
		"status_speed":            ParamDefinition{&pInfo.Speed, reflect.String},
		"status_duplex":           ParamDefinition{&pInfo.Duplex, reflect.String},
		"status_flow":             ParamDefinition{&pInfo.Flow, reflect.String},
		"status_autorevert_delay": ParamDefinition{&pInfo.AutoRevertDelay, reflect.Int},
		"status_ignore_health":    ParamDefinition{&pInfo.IgnoreHealth, reflect.Bool},
		"status_role":             ParamDefinition{&pInfo.Role, reflect.String}} {
		if err = writeToSchema(d, key, param); err != nil {
			return err
		}
	}

	d.SetId(createPortID(pInfo))

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

	for key, param := range map[string]ParamDefinition{
		"admin_up":         ParamDefinition{&req.Up, reflect.Bool},
		"admin_mtu":        ParamDefinition{&req.Mtu, reflect.Int},
		"autorevert_delay": ParamDefinition{&req.AutoRevertDelay, reflect.Int},
		"ignore_health":    ParamDefinition{&req.IgnoreHealth, reflect.Bool},
		"ipspace":          ParamDefinition{&req.IPSpace, reflect.String},
		"role":             ParamDefinition{&req.Role, reflect.String}} {
		written, err := writeToValueIfInCfg(d, key, param)
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
			req.Auto = strconv.FormatBool(isAuto)
			changeCnt++
		}
	}

	if !isAuto {
		for key, param := range map[string]ParamDefinition{
			"admin_speed":  ParamDefinition{&req.Speed, reflect.String},
			"admin_duplex": ParamDefinition{&req.Duplex, reflect.String},
			"admin_flow":   ParamDefinition{&req.Flow, reflect.String}} {
			written, err := writeToValueIfInCfg(d, key, param)
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
