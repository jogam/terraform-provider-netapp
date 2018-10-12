package netapp

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceNetappKeyValue() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetappKeyValueCreate,
		Read:   resourceNetappKeyValueRead,
		Update: resourceNetappKeyValueUpdate,
		Delete: resourceNetappKeyValueDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNetappKeyValueImport,
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"key": {
				Type:        schema.TypeString,
				Description: "NetApp test resource key",
				Required:    true,
				ForceNew:    true,
			},
			"value": {
				Type:        schema.TypeString,
				Description: "NetApp test resource value",
				Required:    true,
			},
		},
	}
}

func resourceNetappKeyValueCreate(d *schema.ResourceData, meta interface{}) error {
	key := d.Get("key").(string)
	value := d.Get("value").(string)

	d.SetId(strings.Join([]string{key, value}, ":"))
	return resourceNetappKeyValueRead(d, meta)
}

func resourceNetappKeyValueRead(d *schema.ResourceData, meta interface{}) error {
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return fmt.Errorf("Resource ID misformed: %s", d.Id())
	}

	key, value := parts[0], parts[1]

	d.Set("key", key)
	d.Set("value", value)

	return nil
}

func resourceNetappKeyValueUpdate(d *schema.ResourceData, meta interface{}) error {
	if d.HasChange("value") {
		// key value has changed, get to it!
		return resourceNetappKeyValueCreate(d, meta)
	}

	return fmt.Errorf("cannot update key! only value @ %s, new values [%s:%s]",
		d.Id(), d.Get("key").(string), d.Get("value").(string))
}

func resourceNetappKeyValueDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceNetappKeyValueImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceNetappKeyValueCreate(d, meta)
	return []*schema.ResourceData{d}, nil
}
