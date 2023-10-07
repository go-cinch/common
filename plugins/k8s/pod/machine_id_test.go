package pod

import (
	"fmt"
	"os"
	"testing"
)

func TestMachineId(t *testing.T) {
	ips := []string{
		"123456",
		"127.0.1.1",
		"127.0.1.2",
		"127.0.0.1",
		"127.0.0.2",
		"192.168.1.1",
		"192.168.1.2",
		"192.168.0.1",
		"192.168.0.2",
		"172.17.1.1",
		"172.17.1.2",
		"172.17.0.1",
		"172.17.0.2",
		"10.20.1.1",
		"10.20.1.2",
		"10.20.0.1",
		"10.20.0.2",
	}
	for _, item := range ips {
		_ = os.Setenv("POD_IP", item)
		id, err := MachineId()
		fmt.Printf("get machine id %d from %s, err: %v\n", id, item, err)
	}
}
