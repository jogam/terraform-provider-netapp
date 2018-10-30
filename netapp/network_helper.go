package netapp

import (
	"fmt"
	"strconv"
	"strings"

	netappnw "github.com/jogam/terraform-provider-netapp/netapp/internal/helper/network"
)

func createNetQualifiedName(vlanInfo *netappnw.VlanInfo) string {
	var builder strings.Builder
	fmt.Fprintf(
		&builder, "%s:%s", vlanInfo.NodeName, vlanInfo.Name)
	return builder.String()
}

func createVlanID(vlanInfo *netappnw.VlanInfo) string {
	var builder strings.Builder
	fmt.Fprintf(
		&builder, "%s|%s|%s",
		vlanInfo.NodeName, vlanInfo.ParentName, vlanInfo.VlanID)
	return builder.String()
}

func getNodePortNameVlanFromVlanID(vlanID string) (string, string, int, error) {
	viParts := strings.Split(vlanID, "|")
	if len(viParts) != 3 {
		return "", "", -1, fmt.Errorf(
			"could not extract information from vlan ID: %s", vlanID)
	}

	vID, err := strconv.Atoi(viParts[2])
	if err != nil {
		return "", "", -1, fmt.Errorf(
			"could not convert vlan ID to int from: %s", vlanID)
	}
	return viParts[0], viParts[1], vID, nil
}
