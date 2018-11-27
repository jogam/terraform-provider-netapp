package netapp

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform/helper/schema"
	netappsys "github.com/jogam/terraform-provider-netapp/netapp/internal/helper/system"
)

func resourceNetAppPortGroup() *schema.Resource {
	// NetAPP VIF/port group blurp: https://kb.netapp.com/app/answers/answer_view/a_id/1001531

	s := map[string]*schema.Schema{
		"mode": &schema.Schema{
			Type: schema.TypeString,
			Description: "Trunking/Group mode for this port group, valid values are:" +
				"'singlemode': single interface of group is active, failover only" +
				"'multimode': all interfaces active, usually requires connection to single switch!" +
				"'multimode_lacp': like multimode, but with 802.3ad fault detection",
			Required: true,
			ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
				v := val.(string)
				switch v {
				case
					"singlemode", "multimode", "multimode_lacp":
					return
				}

				errs = append(errs, fmt.Errorf("%q must be one of [singlemode, multimode, multimode_lacp]", key))
				return
			},
		},

		"load_distribution": &schema.Schema{
			Type: schema.TypeString,
			Description: "Traffic/Load distribution for this port group, valid values are:" +
				"'mac': based on the MAC address of the source" +
				"'ip': based on the IP address of the source (NetAPP default)" +
				"'sequential': round-robin to each interface, can cause issues with TCP and delays (care)" +
				"'port': based on source/destination IP + transport layer port number (recommended)",
			Required: true,
			ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
				v := val.(string)
				switch v {
				case
					"mac", "ip", "sequential", "port":
					return
				}

				errs = append(errs, fmt.Errorf("%q must be one of [mac, ip, sequential, port]", key))
				return
			},
		},

		"ports": &schema.Schema{
			Type:        schema.TypeList,
			Description: "The list of managed object ID's for the ports belonging to this group.",
			Required:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},

		//******************************************************************
		// status section
		//******************************************************************

		"name": &schema.Schema{
			Type:        schema.TypeString,
			Description: "Automagically assigned port group name.",
			Computed:    true,
		},

		"status_group_link": &schema.Schema{
			Type:        schema.TypeString,
			Description: "Overall port group link status, out of [full, partial, none]",
			Computed:    true,
		},

		"status_ports_active": &schema.Schema{
			Type:        schema.TypeList,
			Description: "List of active ports in the port group.",
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},

		"status_ports_inactive": &schema.Schema{
			Type:        schema.TypeList,
			Description: "List of inactive ports in the port group.",
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}

	// merge default node port/group schema with port list
	mergeSchema(s, schemaNodePortGroup())

	return &schema.Resource{
		Schema: s,
		Create: resourceNetAppPortGroupCreate,
		Read:   resourceNetAppPortGroupRead,
		Update: resourceNetAppPortGroupUpdate,
		Delete: resourceNetAppPortGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceNetAppPortGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api
	nodeID := d.Get("node_id").(string)

	// get the NetApp node, e.g. confirm exists
	//nodeInfo, err := netappsys.NodeGetByUUID(client, nodeID)
	_, err := netappsys.NodeGetByUUID(client, nodeID)
	if err != nil {
		return err
	}

	// TODO: determine group name following pattern: a[0..999][a-z], e.g. a0a ... a999z
	// use a0t .. a999t and search via port pattern a*t in net-port-get-iter

	// TODO: create port group via net-port-ifgrp-create

	// configure mtu/up and ports belonging to group via update
	return resourceNetAppPortGroupUpdate(d, meta)
}

func resourceNetAppPortGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api

	// get ID and extract node and port group name
	pgID := d.Id()
	nodeName, portName, err := getNodePortNameFromPortID(pgID)
	if err != nil {
		return fmt.Errorf("could not get node / name from ID [%s], got: %s", pgID, err)
	}

	pgInfo, err := netappsys.PortGroupGetByNames(client, nodeName, portName)
	if err != nil {
		return fmt.Errorf("no PortGroup for Node/Name [%s/%s], got: %s",
			nodeName, portName, err)
	}

	if pgInfo.NonExist {
		d.SetId("")
		return nil
	}

	d.Set("name", pgInfo.GroupName)
	d.Set("status_group_link", pgInfo.GroupLinkStatus)
	d.Set("status_ports_active",
		stringArrayToInterfaceArray(pgInfo.PortsUp))
	d.Set("status_ports_inactive",
		stringArrayToInterfaceArray(pgInfo.PortsDown))

	pInfo, err := netappsys.PortGetByNames(client, nodeName, portName)
	if err != nil {
		return fmt.Errorf("no port for Node/Name [%s/%s], got: %s",
			nodeName, portName, err)
	}

	// read default status and admin_up / admin_mtu
	schemaNodePortGroupRead(d, pInfo)

	// set ID as for port, that ensures that
	// VLAN / ipspace / broadcast domain work identical
	d.SetId(createPortID(pInfo))

	return nil
}

func resourceNetAppPortGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api

	pgID := d.Id()
	nodeName, portName, err := getNodePortNameFromPortID(pgID)
	if err != nil {
		return fmt.Errorf("could not get node / name from ID [%s], got: %s", pgID, err)
	}

	// TODO: implement port group updates, e.g. add/remove ports

	// process the port based changes (up / mtu)
	req := netappsys.PortModifyRequest{}
	req.NodeName = nodeName
	req.PortName = portName

	changeCnt := 0

	for key, param := range map[string]ParamDefinition{
		"admin_up":  ParamDefinition{&req.Up, reflect.Bool},
		"admin_mtu": ParamDefinition{&req.Mtu, reflect.Int}} {
		written, err := writeToValueIfInCfg(d, key, param)
		if err != nil {
			return err
		}

		if written {
			changeCnt++
		}
	}

	if changeCnt > 0 {
		err = netappsys.PortModify(client, &req)
		if err != nil {
			return err
		}
	}

	// do a final read before exiting update
	return resourceNetAppPortGroupRead(d, meta)
}

func resourceNetAppPortGroupDelete(d *schema.ResourceData, meta interface{}) error {
	// can not delete hardware, e.g. do nothing
	return nil
}
