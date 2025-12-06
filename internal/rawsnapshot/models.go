// File: jms-core-x/internal/rawsnapshot/models.go
package rawsnapshot

import (
	"net"
	"time"

	"jms-core-x/internal/discovery"
)

type RawSnapshot struct {
	Time time.Time

	LLDP []discovery.LLDPDevice
	UBNT []discovery.UBNTDevice

	IPScan []IPScanEntry
	ARP    []ARPEntry

	Duration time.Duration
	Errors   []error
}

type IPScanEntry struct {
	IP        net.IP
	Responded bool
	RTT       time.Duration

	PortOpen bool
	Method   string
}

type ARPEntry struct {
	IP        net.IP
	MAC       net.HardwareAddr
	Interface string
	Source    string
}

type SnapshotError struct {
	Source string
	Err    error
}

func (e SnapshotError) Error() string {
	if e.Err == nil {
		return e.Source
	}
	return e.Source + ": " + e.Err.Error()
}
