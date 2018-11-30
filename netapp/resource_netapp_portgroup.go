package netapp

import (
	"fmt"
	"reflect"
	"sort"

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
			ForceNew: true,
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
			ForceNew: true,
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
	nodeInfo, err := netappsys.NodeGetByUUID(client, nodeID)
	if err != nil {
		return err
	}

	// determine group name following pattern: a[0..999][a-z], e.g. a0a ... a999z
	// use a0t .. a999t and search via port pattern a*t in net-port-get-iter
	pfRes, err := netappsys.PortFindByNamePattern(client, nodeInfo.Name, "a*t")
	// create mapping of existing port names
	portNameExists := map[string]bool{}
	for _, pName := range pfRes.Names {
		portNameExists[pName] = true
	}
	// start at a0t and increase integer part until new port name is found
	// NOTE: if for some reason there are already 1000 groups it will fail!
	var newPortName string
	for index := 0; index < 1001; index++ {
		newPortName = fmt.Sprintf("a%vt", index)
		if !portNameExists[newPortName] {
			break
		}
	}

	// create and populate modify / create request
	request := &netappsys.PortGroupModifyRequest{}
	request.NodeName = nodeInfo.Name
	request.GroupName = newPortName
	request.Mode = d.Get("mode").(string)
	request.LoadDistribution = d.Get("load_distribution").(string)

	// create port group
	err = netappsys.PortGroupCreate(client, request)
	if err != nil {
		return err
	}

	// get port info for group as input for resource id generation
	pInfo, err := netappsys.PortGetByNames(client, nodeInfo.Name, newPortName)
	if err != nil {
		return fmt.Errorf(
			"could not find newly created port group in ports, got: %s", err)
	}

	// set ID for upgrade
	d.SetId(createPortID(pInfo))

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

	// get ID for node and set in resource data
	nInfo, err := netappsys.NodeGetByName(client, nodeName)
	if err != nil {
		return fmt.Errorf("could not get ID for node [%s] got: %s", nodeName, err)
	}
	d.Set("node_id", nInfo.UUID)

	pgInfo, err := netappsys.PortGroupGetByNames(client, nodeName, portName)
	if err != nil {
		return fmt.Errorf("no PortGroup for Node/Name [%s/%s], got: %s",
			nodeName, portName, err)
	}

	if pgInfo.NonExist {
		d.SetId("")
		return nil
	}

	// get port ID's for all ports configured in group
	var groupPortIds []string
	portNameDataMap := map[string]string{}
	for _, pN := range pgInfo.Ports {
		pInfo, err := netappsys.PortGetByNames(client, nodeName, pN)
		if err != nil {
			return fmt.Errorf(
				"get port ID's --> no port for "+
					"Node/Name [%s/%s], got: %s",
				nodeName, pN, err)
		}

		pID := createPortID(pInfo)
		portNameDataMap[pN] = pID
		groupPortIds = append(groupPortIds, pID)
	}

	// create port id arrays for up/down ports
	// NOTE: this might not be necessary, nice to have
	var upPortIds []string
	for _, pN := range pgInfo.PortsUp {
		upPortIds = append(upPortIds, portNameDataMap[pN])
	}

	var downPortIds []string
	for _, pN := range pgInfo.PortsDown {
		downPortIds = append(downPortIds, portNameDataMap[pN])
	}

	d.Set("mode", pgInfo.Mode)
	d.Set("load_distribution", pgInfo.LoadDistribution)

	sort.Strings(groupPortIds)
	d.Set("ports", stringArrayToInterfaceArray(groupPortIds))

	d.Set("name", pgInfo.GroupName)
	d.Set("status_group_link", pgInfo.GroupLinkStatus)

	sort.Strings(upPortIds)
	d.Set("status_ports_active",
		stringArrayToInterfaceArray(upPortIds))

	sort.Strings(downPortIds)
	d.Set("status_ports_inactive",
		stringArrayToInterfaceArray(downPortIds))

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

	// Enable partial state mode
	d.Partial(true)

	// check port changes and update accordingly
	if d.HasChange("ports") {
		oldPorts, newPorts := d.GetChange("ports")
		oldPortIds := interfaceArrayToStringArray(oldPorts.([]interface{}))
		newPortIds := interfaceArrayToStringArray(newPorts.([]interface{}))

		// find removed port names
		var rmPortNames []string
		for _, opID := range oldPortIds {
			found := false
			for _, npID := range newPortIds {
				if opID == npID {
					found = true
					break
				}
			}

			if !found {
				// get port name for port ID
				_, pName, err := getNodePortNameFromPortID(opID)
				if err != nil {
					return fmt.Errorf("failed collect remove ports with: %s", err)
				}

				rmPortNames = append(rmPortNames, pName)
			}
		}

		// remove ports from port group
		if len(rmPortNames) > 0 {
			err := netappsys.PortGroupPortsModify(
				client, nodeName, portName,
				rmPortNames, false, true)
			if err != nil {
				return fmt.Errorf("port remove error: %s", err)
			}
		}

		// find added port names
		var addPortNames []string
		for _, npID := range newPortIds {
			found := false
			for _, opID := range oldPortIds {
				if opID == npID {
					found = true
					break
				}
			}

			if !found {
				// get port name for port ID
				_, pName, err := getNodePortNameFromPortID(npID)
				if err != nil {
					return fmt.Errorf("failed collect add ports with: %s", err)
				}

				addPortNames = append(addPortNames, pName)
			}
		}

		if len(addPortNames) > 0 {
			// add ports to broadcast domain
			err := netappsys.PortGroupPortsModify(
				client, nodeName, portName,
				addPortNames, true, false)
			if err != nil {
				return fmt.Errorf("port add error: %s", err)
			}
		}

		// done ports update, indicate in partial state
		d.SetPartial("ports")
	}

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

		// done mtu / up change, indicate in partial
		d.SetPartial("admin_up")
		d.SetPartial("admin_mtu")
	}

	// We succeeded, disable partial mode. This causes Terraform to save
	// all fields again.
	d.Partial(false)

	// do a final read before exiting update
	return resourceNetAppPortGroupRead(d, meta)
}

func resourceNetAppPortGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api

	pgID := d.Id()
	nodeName, portName, err := getNodePortNameFromPortID(pgID)
	if err != nil {
		return fmt.Errorf("could not get node / name from ID [%s], got: %s", pgID, err)
	}

	return netappsys.PortGroupDelete(client, nodeName, portName)
}
