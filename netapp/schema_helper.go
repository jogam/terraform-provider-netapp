package netapp

import (
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
