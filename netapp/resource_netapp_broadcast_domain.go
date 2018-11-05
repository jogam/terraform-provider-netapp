package netapp

import (
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	netappnw "github.com/jogam/terraform-provider-netapp/netapp/internal/helper/network"
)

func resourceNetAppBroadcastDomain() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The name of the Broadcast Domain.",
				Required:    true,
			},

			"ipspace": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The managed object ID of the IPSpace this Broadcast Domain belongs to (netapp default: Default).",
				Required:    true,
				ForceNew:    true,
			},

			"mtu": &schema.Schema{
				Type: schema.TypeInt,
				Description: "MTU to be applied to all ports belonging to this Broadcast Domain." +
					" (NOTE: underlying physical interfaces must be configured to at least" +
					" same MTU value via admin_mtu! - default: 1500)",
				// DOCU: during UP ensure that PHYS-INT MTU was increased in previous apply
				//		 during DOWN on PHYS-INT, ensure that BroadCast domain is first below new value
				Optional: true,
				Default:  1500,
			},

			"ports": &schema.Schema{
				Type:        schema.TypeList,
				Description: "The list of managed object ID's for the ports/vlans belonging to this Broadcast Domain.",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			//******************************************************************
			// status section
			//******************************************************************

			"status_port_update": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The broadcast domains overall port update status.",
				Computed:    true,
			},

			"status_ipspace": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The IPSpace name current used by the broadcast domain.",
				Computed:    true,
			},

			"failover_groups": &schema.Schema{
				Type:        schema.TypeList,
				Description: "The failover port/vlan groups belonging to this broadcast domain.",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"subnet_names": &schema.Schema{
				Type:        schema.TypeList,
				Description: "The subnet names using this broadcast domain.",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},

		Create: resourceNetAppBroadcastDomainCreate,
		Read:   resourceNetAppBroadcastDomainRead,
		Update: resourceNetAppBroadcastDomainUpdate,
		Delete: resourceNetAppBroadcastDomainDelete,

		// as per: https://www.terraform.io/docs/extend/resources.html#importers
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceNetAppBroadcastDomainCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api

	name := d.Get("name").(string)
	req := &netappnw.BcDomainRequest{Name: name}

	mtu, isSet := d.GetOk("mtu")
	if isSet {
		req.Mtu = strconv.Itoa(mtu.(int))
	}

	ipspace, isSet := d.GetOk("ipspace")
	if isSet {
		ipSpaceInfo, err := netappnw.IPSpaceGetByUUID(client, ipspace.(string))
		if err != nil {
			return fmt.Errorf("no IPSpace found for [%s], got: %s", ipspace, err)
		}

		req.IPSpace = ipSpaceInfo.Name
	}

	portIds, isSet := d.GetOk("ports")
	if isSet {
		var qpNames []string
		portIDStrings := interfaceArrayToStringArray(portIds.([]interface{}))
		sort.Strings(portIDStrings)
		for _, pID := range portIDStrings {

			pName, err := getNetQualifiedNameFromID(pID)
			if err != nil {
				return fmt.Errorf("could not process port definition [%s], got: %s", pID, err)
			}

			qpNames = append(qpNames, pName)
		}

		req.Ports = qpNames
	}

	bcInfo, err := netappnw.BcDomainCreate(client, req)

	portStatus := bcInfo.PortUpdateStatus
	if portStatus == "in_progress" {
		portStatus, err = netappnw.BcDomainWaitForInProgressDone(client, name)
		if err != nil {
			return fmt.Errorf("create wait finished caused: %s", err)
		}
	}

	if portStatus == "error" {
		return fmt.Errorf("create failed, use refresh to get details")
	}

	d.SetId(name)

	return resourceNetAppBroadcastDomainRead(d, meta)
}

func resourceNetAppBroadcastDomainRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api

	name := d.Id()
	bcInfo, err := netappnw.BcDomainGet(client, name)
	if err != nil {
		return err
	}

	if bcInfo.NonExist {
		// resource not found, flag as gone
		d.SetId("")
		return nil
	}

	d.Set("name", bcInfo.Name)
	d.Set("status_port_update", bcInfo.PortUpdateStatus)

	if len(bcInfo.Mtu) > 0 {
		// empty string will fail on atoi!
		mtu, err := strconv.Atoi(bcInfo.Mtu)
		if err != nil {
			return fmt.Errorf("could not convert MTU [%s] to int, got: %s", bcInfo.Mtu, err)
		}
		d.Set("mtu", mtu)
	}

	d.Set("status_ipspace", bcInfo.IPSpace)
	ipSpaceInfo, err := netappnw.IPSpaceGetByName(client, bcInfo.IPSpace)
	if err != nil {
		return fmt.Errorf("could not get IPSpace [%s] data, got: %s", bcInfo.IPSpace, err)
	}
	d.Set("ipspace", ipSpaceInfo.UUID)

	if len(bcInfo.Ports) > 0 {
		var pqNames []string
		for _, pInfo := range bcInfo.Ports {
			if pInfo.UpdateStatus != "complete" {
				pqNames = append(pqNames,
					pInfo.Name+" --> ["+pInfo.UpdateStatus+"]:"+pInfo.StatusDetail)
			} else {
				resID, err := getResourceIDfromNetQualifiedName(client, pInfo.Name)
				if err != nil {
					log.Printf("[WARN] could not get resource ID, got: %s", err)
					pqNames = append(pqNames, pInfo.Name)
				} else {
					pqNames = append(pqNames, resID)
				}
			}
		}

		sort.Strings(pqNames)
		err := d.Set("ports", stringArrayToInterfaceArray(pqNames))
		if err != nil {
			return fmt.Errorf("set new port data failed: %s", err)
		}
	} else {
		emptyPorts := make([]interface{}, 0)
		d.Set("ports", emptyPorts)
	}

	sort.Strings(bcInfo.FailoverGroups)
	err = d.Set("failover_groups", stringArrayToInterfaceArray(bcInfo.FailoverGroups))
	if err != nil {
		return fmt.Errorf("set new failover group data failed: %s", err)
	}

	sort.Strings(bcInfo.SubnetNames)
	err = d.Set("subnet_names", stringArrayToInterfaceArray(bcInfo.SubnetNames))
	if err != nil {
		return fmt.Errorf("set new subnet name data failed: %s", err)
	}

	// set the ID of the resource to the name
	d.SetId(bcInfo.Name)

	return nil
}

func resourceNetAppBroadcastDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api

	// Enable partial state mode
	d.Partial(true)

	// need ipspace for part updates, get here
	ipspaceUUIDnew, ipspaceUUIDold := d.GetChange("ipspace")
	ipspaceUUID := ipspaceUUIDold.(string)
	ipSpaceInfo, err := netappnw.IPSpaceGetByUUID(client, ipspaceUUID)
	if err != nil {
		return fmt.Errorf("no IPSpace [%s] found, got: %s", ipspaceUUID, err)
	}

	// update broadcast domain name if changed
	if d.HasChange("name") {
		name, newName := d.GetChange("name")
		err := netappnw.BcDomainRename(
			client, name.(string), ipSpaceInfo.Name, newName.(string))
		if err != nil {
			return err
		}

		// done name change, indicate in partial
		d.SetPartial("name")
	}

	name := d.Get("name").(string)
	d.SetId(name)

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
				// get net qualified name for port/vlan ID
				pName, err := getNetQualifiedNameFromID(opID)
				if err != nil {
					return fmt.Errorf("failed collect remove ports with: %s", err)
				}

				rmPortNames = append(rmPortNames, pName)
			}
		}

		// remove ports from broadcast domain
		if len(rmPortNames) > 0 {
			bcInfo, err := netappnw.BcDomainPortsModify(
				client, name, ipSpaceInfo.Name,
				rmPortNames, false, true)
			if err != nil {
				return fmt.Errorf("port remove error: %s", err)
			}

			portStatus := bcInfo.PortUpdateStatus
			if portStatus == "in_progress" {
				portStatus, err = netappnw.BcDomainWaitForInProgressDone(client, name)
				if err != nil {
					return fmt.Errorf("remove port wait finished caused: %s", err)
				}
			}

			if portStatus == "error" {
				return fmt.Errorf("remove port failed, use refresh to get details")
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
				// get net qualified name for port/vlan ID
				pName, err := getNetQualifiedNameFromID(npID)
				if err != nil {
					return fmt.Errorf("failed collect add ports with: %s", err)
				}

				addPortNames = append(addPortNames, pName)
			}
		}

		if len(addPortNames) > 0 {
			// add ports to broadcast domain
			bcInfo, err := netappnw.BcDomainPortsModify(
				client, name, ipSpaceInfo.Name,
				addPortNames, true, false)
			if err != nil {
				return fmt.Errorf("port add error: %s", err)
			}

			portStatus := bcInfo.PortUpdateStatus
			if portStatus == "in_progress" {
				portStatus, err = netappnw.BcDomainWaitForInProgressDone(client, name)
				if err != nil {
					return fmt.Errorf("add port wait finished caused: %s", err)
				}
			}

			if portStatus == "error" {
				return fmt.Errorf("add port failed, use refresh to get details")
			}
		}

		// done ports update, indicate in partial state
		d.SetPartial("ports")
	}

	// update rest of parameters if changes present
	if d.HasChange("mtu") || d.HasChange("ipspace") {
		// get the new IPSpace UUID IpSpaceInfo for the update
		ipspaceUUID := ipspaceUUIDnew.(string)
		ipSpaceInfo, err := netappnw.IPSpaceGetByUUID(client, ipspaceUUID)
		if err != nil {
			return fmt.Errorf("no IPSpace [%s] found, got: %s", ipspaceUUID, err)
		}

		bcInfo, err := netappnw.BcDomainUpdate(
			client, name,
			ipSpaceInfo.Name, d.Get("mtu").(int))
		if err != nil {
			return fmt.Errorf("general parameter update fail: %s", err)
		}

		portStatus := bcInfo.PortUpdateStatus
		if portStatus == "in_progress" {
			portStatus, err = netappnw.BcDomainWaitForInProgressDone(client, name)
			if err != nil {
				return fmt.Errorf("general param update wait for caused: %s", err)
			}
		}

		if portStatus == "error" {
			return fmt.Errorf("general param update failed, use refresh to get details")
		}

		d.SetPartial("mtu")
		d.SetPartial("ipspace")
	}

	// We succeeded, disable partial mode. This causes Terraform to save
	// all fields again.
	d.Partial(false)

	return nil
}

func resourceNetAppBroadcastDomainDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api
	name := d.Get("name").(string)

	ipSpace := d.Get("ipspace").(string)
	ipSpaceInfo, err := netappnw.IPSpaceGetByUUID(client, ipSpace)
	if err != nil {
		return fmt.Errorf("no IPSpace found for [%s], got: %s", ipSpace, err)
	}

	bcInfo, err := netappnw.BcDomainDelete(client, name, ipSpaceInfo.Name)
	if err != nil {
		return fmt.Errorf("delete failed: %s", err)
	}

	portStatus := bcInfo.PortUpdateStatus
	if portStatus == "in_progress" {
		portStatus, err = netappnw.BcDomainWaitForInProgressDone(client, name)
		if err != nil {
			return fmt.Errorf("delete wait for caused: %s", err)
		}
	}

	if portStatus == "error" {
		return fmt.Errorf("delete failed, use refresh to get details")
	}

	return nil
}
