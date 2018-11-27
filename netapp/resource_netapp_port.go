package netapp

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	netappsys "github.com/jogam/terraform-provider-netapp/netapp/internal/helper/system"
)

func resourceNetAppPort() *schema.Resource {
	s := map[string]*schema.Schema{
		"nic_name": &schema.Schema{
			Type:        schema.TypeString,
			Description: "The name of the nic on the device (usually static).",
			Required:    true,
			ForceNew:    true,
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

				errs = append(errs, fmt.Errorf("%q must be one of [undef, auto, 10 .. 100000]", key))
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
	}

	// add standard port(group) node id, admin_up, admin_mtu and status
	mergeSchema(s, schemaNodePortGroup())

	return &schema.Resource{
		Schema: s,
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

	// read default status and admin_up / admin_mtu
	schemaNodePortGroupRead(d, pInfo)

	// write back admin parameters if they are set in resource data
	for key, param := range map[string]ParamDefinition{
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
