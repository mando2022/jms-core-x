// File: jms-core-x/internal/rawsnapshot/rawsnapshot.go
package rawsnapshot

import (
	"context"
	"time"

	"jms-core-x/internal/discovery"
)

type IPScanner interface {
	Scan(ctx context.Context, cidr string) ([]IPScanEntry, error)
}

type ARPReader interface {
	Read(ctx context.Context) ([]ARPEntry, error)
}

type Engine struct {
	LLDPSource discovery.LLDPSource
	UBNTSource discovery.UBNTSource

	ifaceRealName string
	cidr          string

	ipScanner IPScanner
	arpReader ARPReader
}

func NewEngine(
	lldp discovery.LLDPSource,
	ubnt discovery.UBNTSource,
	ifaceRealName string,
	cidr string,
	ipScanner IPScanner,
	arpReader ARPReader,
) *Engine {
	e := &Engine{
		LLDPSource:    lldp,
		UBNTSource:    ubnt,
		ifaceRealName: ifaceRealName,
		cidr:          cidr,
	}

	if ipScanner != nil {
		e.ipScanner = ipScanner
	} else {
		e.ipScanner = &TCPIPScanner{
			Ports:       []int{80, 443},
			Timeout:     500 * time.Millisecond,
			MaxParallel: 64,
		}
	}

	if arpReader != nil {
		e.arpReader = arpReader
	} else {
		e.arpReader = &SystemARPReader{}
	}

	return e
}

func (e *Engine) BuildSnapshot(ctx context.Context) RawSnapshot {
	start := time.Now()

	snap := RawSnapshot{
		Time:   start,
		Errors: make([]error, 0),
	}

	snap.LLDP = e.LLDPSource.Snapshot()
	snap.UBNT = e.UBNTSource.Snapshot()

	if e.ipScanner != nil && e.cidr != "" {
		ipEntries, err := e.ipScanner.Scan(ctx, e.cidr)
		if err != nil {
			snap.Errors = append(snap.Errors, SnapshotError{"ipscan", err})
		} else {
			snap.IPScan = ipEntries
		}
	}

	if e.arpReader != nil {
		arpEntries, err := e.arpReader.Read(ctx)
		if err != nil {
			snap.Errors = append(snap.Errors, SnapshotError{"arp", err})
		} else {
			snap.ARP = arpEntries
		}
	}

	snap.Duration = time.Since(start)
	return snap
}
