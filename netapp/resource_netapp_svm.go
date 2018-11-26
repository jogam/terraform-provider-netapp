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

func resourceNetAppSVMCreate(d *schema.ResourceData, meta interface{}) error {
	//client := meta.(*NetAppClient).api
	return nil
}

func resourceNetAppSVMRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*NetAppClient).api

	var svmInfo *netappsvm.SvmInfo
	_, err := uuid.Parse(d.Id())
	if err != nil {
		// no valid UUID as SVM ID, assume it is an import
		nInt, isSet := d.GetOk("name")
		if !isSet {
			return fmt.Errorf("SVM import requires 'name' variable be set" +
				" via -var 'SVMNAME' option!")
		}

		svmInfo, err = netappsvm.SvmGetByName(client, nInt.(string))
	} else {
		// valid UUID for SVM, use that for retrieval
		svmInfo, err = netappsvm.SvmGetByUUID(client, d.Id())
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

	aggInfo, err := netappsys.AggrGetByName(client, svmInfo.RootAggr)
	if err != nil {
		return fmt.Errorf(
			"root aggregate [%s] not found by name, got %s",
			svmInfo.RootAggr, err)
	}
	d.Set("rootvol_aggregate", aggInfo.UUID)

	for key, param := range map[string]ParamDefinition{
		"rootvol_security_style":  ParamDefinition{&svmInfo.RootSecStyle, reflect.String},
		"rootvol_name":            ParamDefinition{&svmInfo.RootName, reflect.String},
		"rootvol_size":            ParamDefinition{&svmInfo.RootSize, reflect.String},
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

	if len(svmInfo.Protocols) > 0 {
		err := d.Set("protocols", createProtocolSet(svmInfo.Protocols))
		if err != nil {
			return fmt.Errorf("failed to set protocols with: %s", err)
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

	// TODO: read root volume size and protocol configurations from SVM connection

	// set the ID to UUID
	d.SetId(svmInfo.UUID)

	return nil
}

func resourceNetAppSVMUpdate(d *schema.ResourceData, meta interface{}) error {
	//client := meta.(*NetAppClient).api
	return nil
}

func resourceNetAppSVMDelete(d *schema.ResourceData, meta interface{}) error {
	//client := meta.(*NetAppClient).api
	return nil
}
