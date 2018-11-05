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

func createVlanID(vlanCfg netappnw.VlanConfig) string {
	var builder strings.Builder
	fmt.Fprintf(
		&builder, "%s|%s|%s",
		vlanCfg.GetNodeName(),
		vlanCfg.GetParentName(),
		vlanCfg.GetVlanID())
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

func createSubnetID(subnetInfo *netappnw.SubnetInfo) string {
	var builder strings.Builder
	fmt.Fprintf(
		&builder, "%s|%s|%s",
		subnetInfo.BroadCastDomain,
		subnetInfo.IPSpace,
		subnetInfo.Name)
	return builder.String()
}

func updateSubnetID(subnetID string, newName string) (string, error) {
	parts := strings.Split(subnetID, "|")
	if len(parts) != 3 {
		return "", fmt.Errorf(
			"could not update subnet ID with new name from ID [%s]", subnetID)
	}

	var builder strings.Builder
	fmt.Fprintf(
		&builder, "%s|%s|%s",
		parts[0], parts[1], newName)
	return builder.String(), nil
}

func subnetRequestFromID(subnetID string) (*netappnw.SubnetRequest, error) {
	parts := strings.Split(subnetID, "|")
	if len(parts) != 3 {
		return nil, fmt.Errorf(
			"could not create subnet request from ID [%s]", subnetID)
	}

	return &netappnw.SubnetRequest{
		Name:            parts[2],
		BroadCastDomain: parts[0],
		IPSpace:         parts[1]}, nil
}
