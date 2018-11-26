package netapp

import (
	"github.com/hashicorp/terraform/helper/schema"
	netappsvm "github.com/jogam/terraform-provider-netapp/netapp/internal/helper/svm"
)

func svmProtocolSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The protocol type: 'nfs', 'cifs', 'fcp', 'iscsi'",
			},
		},
	}
}

func createProtocolSet(protInfos []netappsvm.ProtocolInfo) *schema.Set {
	s := make([]interface{}, 0)
	for _, protInfo := range protInfos {
		p := make(map[string]interface{})
		p["type"] = protInfo.Type
		s = append(s, p)
	}

	return schema.NewSet(schema.HashResource(svmProtocolSchema()), s)
}

func createProtocolInfoArray(p interface{}) (*[]netappsvm.ProtocolInfo, error) {
	piList := p.(*schema.Set).List()
	pInfoArray := make([]netappsvm.ProtocolInfo, len(piList))
	for _, pi := range piList {
		// protocol info is a schema.Resource
		protInfoMap := pi.(map[string]interface{})
		// create proto info object with direct read of simple type required elements
		pInfo := netappsvm.ProtocolInfo{Type: protInfoMap["type"].(string)}
		// now do checking: if val, ok := dict["foo"]; ok { // add foo value to info}
		pInfoArray = append(pInfoArray, pInfo)
	}

	return &pInfoArray, nil
}
