package pod

import (
	"fmt"
	"net"
	"os"
)

// MachineId gen machineId from pod ip
// need set status.podIP first
// https://kubernetes.io/docs/tasks/inject-data-application/environment-variable-expose-pod-information/
func MachineId() (uint16, error) {
	ipStr := os.Getenv("POD_IP")
	if len(ipStr) == 0 {
		return 0, fmt.Errorf("cannot get POD_IP")
	}
	ip := net.ParseIP(ipStr).To4()
	if len(ip) < 4 {
		return 0, fmt.Errorf("invalid POD_IP: %s", ipStr)
	}
	// pod ip is Class A, consistent high 8 digits in the same k8s cluster
	// 00000000.00000000.00000000.00000000 => 11111111.11111111.11111111.11111111
	// so we use low 8 digits as machine id
	// 00000000.00000000 => 11111111.11111111
	return uint16(ip[2])<<8 + uint16(ip[3]), nil
}
