package netapp

import (
	netappsys "github.com/jogam/terraform-provider-netapp/netapp/internal/helper/system"
)

func createPortID(portInfo *netappsys.PortInfo) string {
	return portInfo.NodeName + "|" + portInfo.PortName + "|" + portInfo.Mac
}
