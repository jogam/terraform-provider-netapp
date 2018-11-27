package netapp

import (
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/jogam/terraform-provider-netapp/netapp/internal/helper/pythonapi"
	netappsys "github.com/jogam/terraform-provider-netapp/netapp/internal/helper/system"
)

func schemaNodePortGroup() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"node_id": &schema.Schema{
			Type:        schema.TypeString,
			Description: "The managed object ID of the node to configure the port (group) for.",
			Required:    true,
			ForceNew:    true,
		},

		"admin_up": &schema.Schema{
			Type:        schema.TypeBool,
			Description: "Change the port (group) status to up if true.",
			Optional:    true,
		},

		"admin_mtu": &schema.Schema{
			Type:        schema.TypeInt,
			Description: "Configure MTU for this port (group).",
			Optional:    true,
			Default:     1500,
		},

		//******************************************************************
		// status section
		//******************************************************************

		"status": &schema.Schema{
			Type:        schema.TypeString,
			Description: "The current port (group) status.",
			Computed:    true,
		},

		"health": &schema.Schema{
			Type:        schema.TypeString,
			Description: "Current health status for this port (group).",
			Computed:    true,
		},

		"mac_address": &schema.Schema{
			Type:        schema.TypeString,
			Description: "The port (group)s MAC address.",
			Computed:    true,
		},

		"broadcast_domain": &schema.Schema{
			Type:        schema.TypeString,
			Description: "port (group)s associated broadcast domain.",
			Computed:    true,
		},

		"status_ipspace": &schema.Schema{
			Type:        schema.TypeString,
			Description: "Current port (group)s associated ipspace.",
			Computed:    true,
		},

		"status_mtu": &schema.Schema{
			Type:        schema.TypeInt,
			Description: "Current MTU for this port (group).",
			Computed:    true,
		},

		"status_auto": &schema.Schema{
			Type:        schema.TypeBool,
			Description: "Current autonegotiate status on this port (group).",
			Computed:    true,
		},

		"status_speed": &schema.Schema{
			Type:        schema.TypeString,
			Description: "Current speed for this port (group).",
			Computed:    true,
		},

		"status_duplex": &schema.Schema{
			Type:        schema.TypeString,
			Description: "Current duplex status for this port (group).",
			Computed:    true,
		},

		"status_flow": &schema.Schema{
			Type:        schema.TypeString,
			Description: "Current flow status for this port (group).",
			Computed:    true,
		},

		"status_autorevert_delay": &schema.Schema{
			Type:        schema.TypeInt,
			Description: "Current delay in seconds before auto-reverting a LIF to this cluster port (group).",
			Computed:    true,
		},

		"status_ignore_health": &schema.Schema{
			Type:        schema.TypeBool,
			Description: "Current ignore port health status on hosted logical interface.",
			Computed:    true,
		},

		"status_role": &schema.Schema{
			Type:        schema.TypeString,
			Description: "Current port (group) role (deprecated in ONTAP 8.3)",
			Computed:    true,
		},
	}
}

func schemaNodePortGroupRead(d *schema.ResourceData, portInfo *netappsys.PortInfo) error {

	// write back admin parameters if they are set in resource data
	for key, param := range map[string]ParamDefinition{
		"admin_up":  ParamDefinition{&portInfo.AdminUp, reflect.Bool},
		"admin_mtu": ParamDefinition{&portInfo.AdminMtu, reflect.Int}} {
		if err := writeToSchemaIfInCfg(d, key, param); err != nil {
			return err
		}
	}

	// status write back - computed/info params, e.g. possibly not specified by user
	for key, param := range map[string]ParamDefinition{
		"status":                  ParamDefinition{&portInfo.Status, reflect.String},
		"health":                  ParamDefinition{&portInfo.Health, reflect.String},
		"mac_address":             ParamDefinition{&portInfo.Mac, reflect.String},
		"broadcast_domain":        ParamDefinition{&portInfo.BroadCastDomain, reflect.String},
		"status_ipspace":          ParamDefinition{&portInfo.IPSpace, reflect.String},
		"status_mtu":              ParamDefinition{&portInfo.Mtu, reflect.Int},
		"status_auto":             ParamDefinition{&portInfo.Auto, reflect.Bool},
		"status_speed":            ParamDefinition{&portInfo.Speed, reflect.String},
		"status_duplex":           ParamDefinition{&portInfo.Duplex, reflect.String},
		"status_flow":             ParamDefinition{&portInfo.Flow, reflect.String},
		"status_autorevert_delay": ParamDefinition{&portInfo.AutoRevertDelay, reflect.Int},
		"status_ignore_health":    ParamDefinition{&portInfo.IgnoreHealth, reflect.Bool},
		"status_role":             ParamDefinition{&portInfo.Role, reflect.String}} {
		if err := writeToSchema(d, key, param); err != nil {
			return err
		}
	}

	return nil
}

func createPortID(portInfo *netappsys.PortInfo) string {
	var builder strings.Builder
	fmt.Fprintf(
		&builder, "%s|%s|%s",
		portInfo.NodeName, portInfo.PortName, portInfo.Mac)
	return builder.String()
}

func getNodePortNameFromPortID(portID string) (string, string, error) {
	pIDParts := strings.Split(portID, "|")
	if len(pIDParts) != 3 {
		return "", "", fmt.Errorf("invalid port id: %s", portID)
	}

	return pIDParts[0], pIDParts[1], nil
}

func getNetQualifiedNameFromID(pvid string) (string, error) {
	parts := strings.Split(pvid, "|")
	if len(parts) != 3 {
		return "", fmt.Errorf(
			"net qualified item ID must have 3 parts"+
				" separated by '|', got: %s", pvid)
	}

	var builder strings.Builder
	// always assume that ID starts with NODE|PORT-NAME
	fmt.Fprintf(&builder, "%s:%s", parts[0], parts[1])

	_, err := net.ParseMAC(parts[2])
	if err == nil {
		// its a port: NODE|PORT-NAME|MAC
		return builder.String(), nil
	}

	vlanID, err := strconv.Atoi(parts[2])
	if err == nil && vlanID > 0 {
		// should be a vlan: NODE|PORT-NAME|VLAN-ID
		fmt.Fprintf(&builder, "-%s", parts[2])
		return builder.String(), nil
	}

	return "", fmt.Errorf(
		"could not get net-qualified-name for [%s], got: %s with err: %s",
		pvid, builder.String(), err)
}

func getResourceIDfromNetQualifiedName(client *pythonapi.NetAppAPI, nqName string) (string, error) {
	parts := strings.Split(nqName, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf(
			"net-qualified name format shoud be "+
				"[NodeName:PortName], got: %s", nqName)
	}

	pInfo, err := netappsys.PortGetByNames(client, parts[0], parts[1])
	if err != nil {
		return "", fmt.Errorf("could not get port [%s] info, got: %s", nqName, err)
	}

	var builder strings.Builder
	switch pInfo.Type {
	case "physical":
		builder.WriteString(createPortID(pInfo))
	case "vlan":
		fmt.Fprintf(
			&builder, "%s|%s|%s",
			pInfo.VlanNode,
			pInfo.VlanPort,
			pInfo.VlanID)
	default:
		return "", fmt.Errorf(
			"unsupported port type [%s] for net-qualified name [%s]",
			pInfo.Type, nqName)
	}

	return builder.String(), nil
}
