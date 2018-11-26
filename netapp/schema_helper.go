package netapp

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func stringArrayToTypeSet(array []string) *schema.Set {
	arr := stringArrayToInterfaceArray(array)
	return schema.NewSet(schema.HashString, arr)
}

func stringArrayToInterfaceArray(array []string) []interface{} {
	arr := make([]interface{}, 0)
	for _, str := range array {
		arr = append(arr, str)
	}

	return arr
}

func interfaceArrayToStringArray(array []interface{}) []string {
	var arr []string
	for _, v := range array {
		arr = append(arr, v.(string))
	}

	return arr
}

// ParamDefinition is a simple representation of a terraform
// data/resource parameter
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
