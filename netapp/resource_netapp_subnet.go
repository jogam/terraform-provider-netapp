package netapp

import (
	"fmt"
	"net"
	"sort"

	"github.com/hashicorp/terraform/helper/schema"
	netappnw "github.com/jogam/terraform-provider-netapp/netapp/internal/helper/network"
)

func resourceNetAppSubnet() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The name of the Subnet.",
				Required:    true,
			},

			"broadcast_domain": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The managed object ID of the Broadcast Domain this subnet is created for.",
				Required:    true,
				ForceNew:    true,
			},

			"subnet": &schema.Schema{
				Type:        schema.TypeString,
				Description: "IP subnet, e.g. 192.168.1.0/24",
				Required:    true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					cidr := val.(string)
					_, _, err := net.ParseCIDR(cidr)
					if err != nil {
						errs = append(errs, fmt.Errorf(
							"[%s] not a valid subnet: %s", cidr, err))
					}

					return
				},
			},

			"gateway": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Gateway for subnet, e.g. 192.168.1.1",
				Optional:    true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					ip := val.(string)
					if net.ParseIP(ip) == nil {
						errs = append(errs, fmt.Errorf(
							"[%s] not a valid gateway IP", ip))
					}

					return
				},
			},

			"ip_ranges": &schema.Schema{
				Type: schema.TypeList,
				Description: "IP address/range that this subnet provides. For example: " +
					"IPv4: 192.168.1.2 or 192.168.1.5-192.168.1.9 etc. or " +
					"IPv6: f6::c3 or f6::c5-f6::c9 etc..",
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			//******************************************************************
			// status section
			//******************************************************************

			"stat_ip_total": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Total # of IP addresses in this subnet.",
				Computed:    true,
			},

			"stat_ip_used": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Count of used IP addresses in this subnet.",
				Computed:    true,
			},

			"stat_ip_available": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Count of availabe IP addresses in this subnet.",
				Computed:    true,
			},
		},

		Create: resourceNetAppSubnetCreate,
		Read:   resourceNetAppSubnetRead,
		Update: resourceNetAppSubnetUpdate,
		Delete: resourceNetAppSubnetDelete,

		// as per: https://www.terraform.io/docs/extend/resources.html#importers
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func writeSubnetInfoToMeta(
	snInfo *netappnw.SubnetInfo, d *schema.ResourceData) error {
	d.Set("name", snInfo.Name)
	d.Set("broadcast_domain", snInfo.BroadCastDomain)
	d.Set("subnet", snInfo.Subnet)
	d.Set("gateway", snInfo.Gateway)

	if len(snInfo.IPRanges) > 0 {
		// write back ip ranges
		sort.Strings(snInfo.IPRanges)
		err := d.Set("ip_ranges", stringArrayToInterfaceArray(snInfo.IPRanges))
		if err != nil {
			return fmt.Errorf("set ip ranges failed: %s", err)
		}
	} else {
		emptyRanges := make([]interface{}, 0)
		d.Set("ip_ranges", emptyRanges)
	}

	d.Set("stat_ip_total", snInfo.IPCount)
	d.Set("stat_ip_used", snInfo.IPUsed)
	d.Set("stat_ip_available", snInfo.IPAvailable)

	d.SetId(createSubnetID(snInfo))

	return nil
}

func resourceNetAppSubnetCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api

	request := &netappnw.SubnetRequest{
		Name: d.Get("name").(string),
	}

	bcd := d.Get("broadcast_domain").(string)
	// get broadcast domain info, name & ipspace is required for create
	// NOTE: broadcast domain ID equals name
	bcInfo, err := netappnw.BcDomainGet(client, bcd)
	if err != nil {
		return fmt.Errorf("no valid broadcast domain exists, got: %s", err)
	}
	request.BroadCastDomain = bcInfo.Name
	request.IPSpace = bcInfo.IPSpace

	subNet := d.Get("subnet").(string)
	// get the subnet, already checked for error via validate func
	// here we want to correct user notation issues!!
	_, snNET, _ := net.ParseCIDR(subNet)
	request.Subnet = snNET.String()

	gw, isSet := d.GetOk("gateway")
	if isSet {
		request.Gateway = gw.(string)
	}

	ipRangeInts, isSet := d.GetOk("ip_ranges")
	if isSet {
		ipRanges := interfaceArrayToStringArray(ipRangeInts.([]interface{}))
		sort.Strings(ipRanges)
		request.IPRanges = ipRanges
	}

	sNInfo, err := netappnw.SubnetCreate(client, request)
	if err != nil {
		return err
	}

	return writeSubnetInfoToMeta(sNInfo, d)
}

func resourceNetAppSubnetRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api

	request, err := subnetRequestFromID(d.Id())
	if err != nil {
		return err
	}

	sNInfo, err := netappnw.SubnetGet(client, request)
	if err != nil {
		return err
	}

	if sNInfo.NonExist {
		d.SetId("")
		return nil
	}

	return writeSubnetInfoToMeta(sNInfo, d)
}
func resourceNetAppSubnetUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api

	// Enable partial state mode
	d.Partial(true)

	// need ipspace for part updates, get here
	bcd := d.Get("broadcast_domain").(string)
	bcdInfo, err := netappnw.BcDomainGet(client, bcd)
	if err != nil {
		return fmt.Errorf("no valid broadcast domain, got: %s", err)
	}

	// update subnet name if changed
	if d.HasChange("name") {
		name, newName := d.GetChange("name")
		request := &netappnw.SubnetRequest{
			Name:    name.(string),
			NewName: newName.(string),
			IPSpace: bcdInfo.IPSpace}
		err := netappnw.SubnetRename(client, request)
		if err != nil {
			return err
		}

		// done name change, indicate in partial
		d.SetPartial("name")
	}

	// udpate resource ID with new subnet name
	name := d.Get("name").(string)
	newSnID, err := updateSubnetID(d.Id(), name)
	if err != nil {
		return fmt.Errorf(
			"resource ID update failed during rename, got: %s", err)
	}
	d.SetId(newSnID)

	// check remove old ip ranges
	if d.HasChange("ip_ranges") {
		oldRangeInts, newRangeInts := d.GetChange("ip_ranges")
		oldRanges := interfaceArrayToStringArray(oldRangeInts.([]interface{}))
		newRanges := interfaceArrayToStringArray(newRangeInts.([]interface{}))

		// find removed IP ranges
		var rmRanges []string
		for _, orName := range oldRanges {
			found := false
			for _, nrName := range newRanges {
				if orName == nrName {
					found = true
					break
				}
			}

			if !found {
				rmRanges = append(rmRanges, orName)
			}
		}

		// remove ip ranges from subnet
		if len(rmRanges) > 0 {
			err := netappnw.SubnetIpRangeModify(
				client, name, bcdInfo.IPSpace,
				rmRanges, false, true)
			if err != nil {
				return fmt.Errorf("ip range remove error: %s", err)
			}
		}
	}

	// update ip subnet and gateway before adding new ranges
	if d.HasChange("gateway") || d.HasChange("subnet") {
		request := &netappnw.SubnetRequest{
			Name: name, IPSpace: bcdInfo.IPSpace,
			Gateway: d.Get("gateway").(string),
		}

		subNet := d.Get("subnet").(string)
		// get the subnet, already checked for error via validate func
		// here we want to correct user notation issues!!
		_, snNET, _ := net.ParseCIDR(subNet)
		request.Subnet = snNET.String()

		err := netappnw.SubnetModify(client, request)
		if err != nil {
			return fmt.Errorf("subnet GW/NET update failed: %s", err)
		}

		d.SetPartial("gateway")
		d.SetPartial("subnet")
	}

	// add new ip ranges after subnet/gateway change
	if d.HasChange("ip_ranges") {
		oldRangeInts, newRangeInts := d.GetChange("ip_ranges")
		oldRanges := interfaceArrayToStringArray(oldRangeInts.([]interface{}))
		newRanges := interfaceArrayToStringArray(newRangeInts.([]interface{}))

		// find added ip ranges
		var addRanges []string
		for _, nrName := range newRanges {
			found := false
			for _, orName := range oldRanges {
				if orName == nrName {
					found = true
					break
				}
			}

			if !found {
				addRanges = append(addRanges, nrName)
			}
		}

		if len(addRanges) > 0 {
			// add ip ranges to subnet
			err := netappnw.SubnetIpRangeModify(
				client, name, bcdInfo.IPSpace,
				addRanges, true, false)
			if err != nil {
				return fmt.Errorf("ip range add error: %s", err)
			}
		}

		// done ports update, indicate in partial state
		d.SetPartial("ip_ranges")
	}

	// We succeeded, disable partial mode. This causes Terraform to save
	// all fields again.
	d.Partial(false)

	return resourceNetAppSubnetRead(d, meta)
}

func resourceNetAppSubnetDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api

	request, err := subnetRequestFromID(d.Id())
	if err != nil {
		return err
	}

	return netappnw.SubnetDelete(client, request)
}
