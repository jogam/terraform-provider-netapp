package netapp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jogam/terraform-provider-netapp/netapp/internal/helper/pythonapi"
	netappsys "github.com/jogam/terraform-provider-netapp/netapp/internal/helper/system"
)

func createPortID(portInfo *netappsys.PortInfo) string {
	var builder strings.Builder
	fmt.Fprintf(
		&builder, "%s|%s|%s",
		portInfo.NodeName, portInfo.PortName, portInfo.Mac)
	return builder.String()
}

func getNodePortNameFromPortID(portID string) (string, string, error) {
	pIDParts := strings.Split(portID, "|")
	if len(pIDParts) != 3 {
		return "", "", fmt.Errorf("invalid port id: %s", portID)
	}

	return pIDParts[0], pIDParts[1], nil
}

func getNetQualifiedNameFromID(pvid string) (string, error) {
	parts := strings.Split(pvid, "|")
	if len(parts) != 3 {
		return "", fmt.Errorf(
			"net qualified item ID must have 3 parts"+
				" separated by '|', got: %s", pvid)
	}

	var builder strings.Builder
	// always assume that ID starts with NODE|PORT-NAME
	fmt.Fprintf(&builder, "%s:%s", parts[0], parts[1])

	if len(strings.Split(parts[2], ":")) == 6 {
		// its a port: NODE|PORT-NAME|MAC
		return builder.String(), nil
	}

	vlanID, err := strconv.Atoi(parts[2])
	if err == nil && vlanID > 0 {
		// should be a vlan: NODE|PORT-NAME|VLAN-ID
		fmt.Fprintf(&builder, "-%s", parts[2])
		return builder.String(), nil
	}

	return "", fmt.Errorf(
		"could not get net-qualified-name for [%s], got: %s with err: %s",
		pvid, builder.String(), err)
}

func getResourceIDfromNetQualifiedName(client *pythonapi.NetAppAPI, nqName string) (string, error) {
	parts := strings.Split(nqName, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf(
			"net-qualified name format shoud be "+
				"[NodeName:PortName], got: %s", nqName)
	}

	pInfo, err := netappsys.PortGetByNames(client, parts[0], parts[1])
	if err != nil {
		return "", fmt.Errorf("could not get port [%s] info, got: %s", nqName, err)
	}

	var builder strings.Builder
	switch pInfo.Type {
	case "physical":
		builder.WriteString(createPortID(pInfo))
	case "vlan":
		fmt.Fprintf(
			&builder, "%s|%s|%s",
			pInfo.VlanNode,
			pInfo.VlanPort,
			pInfo.VlanID)
	default:
		return "", fmt.Errorf(
			"unsupported port type [%s] for net-qualified name [%s]",
			pInfo.Type, nqName)
	}

	return builder.String(), nil
}
