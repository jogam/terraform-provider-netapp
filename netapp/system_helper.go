package netapp

import (
	"fmt"
	"strings"

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
