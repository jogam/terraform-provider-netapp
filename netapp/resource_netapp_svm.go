package netapp

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"

	"github.com/hashicorp/terraform/helper/schema"
	netappnw "github.com/jogam/terraform-provider-netapp/netapp/internal/helper/network"
	netappsvm "github.com/jogam/terraform-provider-netapp/netapp/internal/helper/svm"
	netappsys "github.com/jogam/terraform-provider-netapp/netapp/internal/helper/system"
)

func resourceNetAppSVM() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The name of the SVM (storage virtual machine) / vserver.",
				Required:    true,
			},

			"ipspace": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The managed object ID of the IP space this SVM/vserver uses.",
				Required:    true,
				ForceNew:    true,
			},

			"rootvol_aggregate": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The managed object ID for the Aggregate the SVM root volume should be created in!",
				Required:    true,
				ForceNew:    true,
			},

			"rootvol_security_style": &schema.Schema{
				Type: schema.TypeString,
				Description: "The SVM root volume security style: " +
					"'unix' for NFS, 'ntfs' for CIFS, 'mixed' for both" +
					" NetAPP default: 'unix'",
				Optional: true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					style := val.(string)
					switch style {
					case "unix", "ntfs", "mixed":
						return
					}

					errs = append(errs, fmt.Errorf("%q must be one of [unix, ntfs, mixed]", key))
					return
				},
			},

			"rootvol_name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Name of the SVM root volume, defaults to $SVM-NAME$_root.",
				Optional:    true,
			},

			"rootvol_size": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: "Size of the SVM root volume in bytes with extensions: (" +
					"k [kB], m [mB], g [GB], t [TB]), e.g. 50m would be 50 MB, default 1GB",
			},

			"rootvol_retention_hours": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Retention of root volume [h] after SVM delete, as string!.",
			},

			"protocol": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Protocol definition(s) for this SVM.",
				Elem:        svmProtocolSchema(),
			},

			//******************************************************************
			// status section
			//******************************************************************

			"status_locked": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "The (config) locked state of the SVM.",
				Computed:    true,
			},

			"status_state_oper": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The operational state of the SVM.",
				Computed:    true,
			},

			"status_state_svm": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The state of the SVM itself.",
				Computed:    true,
			},

			"status_proto_enabled": &schema.Schema{
				Type:        schema.TypeList,
				Description: "List of enabled protocols on the SVM.",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"status_proto_inactive": &schema.Schema{
				Type:        schema.TypeList,
				Description: "List of inactive protocols on the SVM.",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"status_rootvol_security_style": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The SVM root volume security style as defined in NetAPP",
				Computed:    true,
			},

			"status_rootvol_name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Name of the SVM root volume as set in NetAPP.",
				Computed:    true,
			},

			"status_rootvol_size": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Size of the SVM root volume as set in NetAPP",
			},

			"status_rootvol_retention_hours": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Retention of root volume [h] after SVM delete, as string!.",
			},
		},

		Create: resourceNetAppSVMCreate,
		Read:   resourceNetAppSVMRead,
		Update: resourceNetAppSVMUpdate,
		Delete: resourceNetAppSVMDelete,

		// as per: https://www.terraform.io/docs/extend/resources.html#importers
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resolveSvmJob(client *NetAppClient, jobRes *netappsvm.JobResult, cmdType string) error {
	// transfer error / status since job and svm create/delete status are different
	errCode := jobRes.ErrNo
	errMsg := jobRes.ErrMsg
	success := (jobRes.Status == "succeeded")
	if jobRes.Status == "in_progress" {
		// status in progress wait for job to 'end'
		jInfo, err := netappsys.JobWaitDone(client.api, jobRes.JobID)
		if err != nil {
			return fmt.Errorf("SVM %s job wait error: %s", cmdType, err)
		}

		// job ended, transfer error / status
		errCode = jInfo.ErrNo
		errMsg = jInfo.Message
		success = (jInfo.Status == "success")
	}

	if !success {
		return fmt.Errorf(
			"%s SVM failed with [err#] MSG: [%v] %s",
			cmdType, errCode, errMsg)
	}

	return nil
}

func resourceNetAppSVMCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api

	request := &netappsvm.Request{}
	request.Name = d.Get("name").(string)

	ipsInfo, err := netappnw.IPSpaceGetByUUID(client, d.Get("ipspace").(string))
	if err != nil {
		return fmt.Errorf("could not get ipspace info, got: %s", err)
	}
	request.IPSpace = ipsInfo.Name

	aggInfo, err := netappsys.AggrGetByUUID(client, d.Get("rootvol_aggregate").(string))
	if err != nil {
		return fmt.Errorf("could not get root aggregate data, got: %s", err)
	}
	request.RootAggr = aggInfo.Name

	rtSecStyle, isSet := d.GetOk("rootvol_security_style")
	if isSet {
		request.RootSecStyle = rtSecStyle.(string)
	}

	rtVolName, isSet := d.GetOk("rootvol_name")
	if isSet {
		request.RootName = rtVolName.(string)
	} else {
		// if root volume name not provided set to
		request.RootName = request.Name + "_root"
	}

	// create the SVM
	svmJobRes, err := netappsvm.Create(client, request)
	if err != nil {
		return fmt.Errorf("SVM create error: %s", err)
	}

	// wait for job to complete and process data
	err = resolveSvmJob(meta.(*NetAppClient), svmJobRes, "create")
	if err != nil {
		return err
	}

	// must get UUID for SVM and set resource ID for update
	svmInfo, err := netappsvm.GetByName(client, svmJobRes.Name)
	if err != nil {
		return fmt.Errorf(
			"failed to read newly created SVM, got: %s", err)
	}
	d.SetId(svmInfo.UUID)

	// update the newly created resource to configure volume/proto's
	return resourceNetAppSVMUpdate(d, meta)
}

func resourceNetAppSVMRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api

	var svmInfo *netappsvm.Info
	_, err := uuid.Parse(d.Id())
	if err != nil {
		// no valid UUID as SVM ID, assume it is an import
		nInt, isSet := d.GetOk("name")
		if !isSet {
			return fmt.Errorf("SVM import requires 'name' variable be set" +
				" via -var 'SVMNAME' option!")
		}

		svmInfo, err = netappsvm.GetByName(client, nInt.(string))
	} else {
		// valid UUID for SVM, use that for retrieval
		svmInfo, err = netappsvm.GetByUUID(client, d.Id())
	}

	if err != nil {
		// could not get SVM
		return fmt.Errorf("could not retrieve SVM info, got: %s", err)
	}

	// check if SVM exists and reset/return if not
	if svmInfo.NonExist {
		d.SetId("")
		return nil
	}

	d.Set("name", svmInfo.Name)
	ipsInfo, err := netappnw.IPSpaceGetByName(client, svmInfo.IPSpace)
	if err != nil {
		// that is very unlikely, but why not...
		return fmt.Errorf(
			"SVM ipspace [%s] not found by name, got: %s",
			svmInfo.IPSpace, err)
	}
	d.Set("ipspace", ipsInfo.UUID)

	if len(svmInfo.RootAggr) > 0 {
		aggInfo, err := netappsys.AggrGetByName(client, svmInfo.RootAggr)
		if err != nil {
			return fmt.Errorf(
				"root aggregate [%s] not found by name, got %s",
				svmInfo.RootAggr, err)
		}
		d.Set("rootvol_aggregate", aggInfo.UUID)
	}

	for key, param := range map[string]ParamDefinition{
		"rootvol_security_style":  ParamDefinition{&svmInfo.RootSecStyle, reflect.String},
		"rootvol_name":            ParamDefinition{&svmInfo.RootName, reflect.String},
		"rootvol_retention_hours": ParamDefinition{&svmInfo.RootRetention, reflect.String}} {
		if err = writeToSchemaIfInCfg(d, key, param); err != nil {
			return err
		}

		// write rootvol info's to status flags
		key = "status_" + key
		if err = writeToSchema(d, key, param); err != nil {
			return err
		}
	}

	// write status values back
	d.Set("status_locked", svmInfo.ConfigLocked)
	d.Set("status_state_oper", svmInfo.OperState)
	d.Set("status_state_svm", svmInfo.SvmState)
	d.Set("status_proto_enabled",
		stringArrayToInterfaceArray(svmInfo.ProtoEnabled))
	d.Set("status_proto_inactive",
		stringArrayToInterfaceArray(svmInfo.ProtoInactive))

	// query for root volume size
	volInfReq := netappsvm.VolumeRequest{}
	volInfReq.SvmInstanceName = svmInfo.Name
	volInfReq.VolumeName = svmInfo.RootName
	volInfo, err := netappsvm.VolumeSizeCommand(client, &volInfReq)
	if err != nil {
		return fmt.Errorf("failed to get root vol size, got: %s", err)
	}
	if err = writeToSchemaIfInCfg(
		d, "rootvol_size",
		ParamDefinition{&volInfo.Size, reflect.String}); err != nil {
		return err
	}
	d.Set("status_rootvol_size", volInfo.Size)

	// TODO: read protocol configurations from SVM connection
	// if len(svmInfo.Protocols) > 0 {
	// 	err := d.Set("protocols", createProtocolSet(svmInfo.Protocols))
	// 	if err != nil {
	// 		return fmt.Errorf("failed to set protocols with: %s", err)
	// 	}
	// }

	// set the ID to UUID
	d.SetId(svmInfo.UUID)

	return nil
}

func resourceNetAppSVMUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api

	svmInfo, err := netappsvm.GetByUUID(client, d.Id())
	if err != nil {
		// no SVM exists for ID!!
		return err
	}

	// Enable partial state mode
	d.Partial(true)

	// name changed, if so act on it
	if d.HasChange("name") {
		name, newName := d.GetChange("name")

		// only do rename if both names have a valid value
		if len(name.(string)) > 0 && len(newName.(string)) > 0 {
			err = netappsvm.Rename(client, name.(string), newName.(string))
			if err != nil {
				return err
			}

			// make sure we update the new name in the SVM info object!
			svmInfo.Name = newName.(string)

			// done name update, indicate in partial state
			d.SetPartial("name")
		}
	}

	// size changed, resize it
	if d.HasChange("rootvol_size") {
		// issue root volume re-size
		volSizeReq := netappsvm.VolumeRequest{}
		volSizeReq.SvmInstanceName = svmInfo.Name
		volSizeReq.VolumeName = svmInfo.RootName
		volSizeReq.Size = d.Get("rootvol_size").(string)
		volInfo, err := netappsvm.VolumeSizeCommand(client, &volSizeReq)
		if err != nil {
			return fmt.Errorf("failed to resize root vol, got: %s", err)
		}

		// write back returned value and indicate partial state
		d.Set("rootvol_size", volInfo.Size)
		d.Set("status_rootvol_size", volInfo.Size)
		d.SetPartial("rootvol_size")
	}

	// following later...
	// TODO: retention is probably on SVM level volume API!
	// NOTE: retention hours <-- should they be int vs string?

	// "rootvol_retention_hours": &schema.Schema{
	// 	Type:        schema.TypeString,
	// 	Optional:    true,
	// 	Description: "Retention of root volume [h] after SVM delete, as string!.",
	// },

	// TODO: create the required protocols etc...
	// "protocol": &schema.Schema{
	// 	Type:        schema.TypeSet,
	// 	Optional:    true,
	// 	Description: "Protocol definition(s) for this SVM.",
	// 	Elem:        svmProtocolSchema(),
	// },

	// We succeeded, disable partial mode. This causes Terraform to save
	// all fields again.
	d.Partial(false)

	return resourceNetAppSVMRead(d, meta)
}

func resourceNetAppSVMDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api

	svmInfo, err := netappsvm.GetByUUID(client, d.Id())
	if err != nil {
		return fmt.Errorf("SVM get during delete error: %s", err)
	}

	// stop SVM
	err = netappsvm.ExecuteSimpleCommand(
		client, svmInfo.Name,
		netappsvm.StopCmd, false)
	if err != nil {
		return fmt.Errorf(
			"SVM delete failed during SVM stop with: %s", err)
	}

	// take SVM root volume offline
	err = netappsvm.VolumeSimpleCommand(
		client, svmInfo.Name,
		svmInfo.RootName, netappsvm.VolumeOfflineCommand)
	if err != nil {
		return fmt.Errorf(
			"SVM delete failed during rootVol offline with: %s", err)
	}

	// delete SVM root volume
	err = netappsvm.VolumeSimpleCommand(
		client, svmInfo.Name,
		svmInfo.RootName, netappsvm.VolumeDeleteCommand)
	if err != nil {
		return fmt.Errorf(
			"SVM delete failed during rootVol delete with: %s", err)
	}

	// delete the SVM
	svmJobRes, err := netappsvm.DeleteByName(client, svmInfo.Name)
	if err != nil {
		return fmt.Errorf("SVM delete error: %s", err)
	}

	// wait for job to complete and process data
	return resolveSvmJob(meta.(*NetAppClient), svmJobRes, "create")
}
