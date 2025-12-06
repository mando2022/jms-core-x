// File: jms-core-x/internal/rawsnapshot/ip_scan.go
package rawsnapshot

import (
	"context"
	"net"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type TCPIPScanner struct {
	Ports       []int
	Timeout     time.Duration
	MaxParallel int
}

func (s *TCPIPScanner) Scan(ctx context.Context, cidr string) ([]IPScanEntry, error) {
	if cidr == "" {
		return nil, nil
	}

	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	ips := expandCIDR(ipNet)
	if len(ips) == 0 {
		return nil, nil
	}

	ports := s.Ports
	if len(ports) == 0 {
		ports = []int{80, 443}
	}

	timeout := s.Timeout
	if timeout <= 0 {
		timeout = 500 * time.Millisecond
	}

	maxParallel := s.MaxParallel
	if maxParallel <= 0 {
		maxParallel = runtime.NumCPU() * 4
	}

	type job struct{ IP net.IP }
	jobs := make(chan job)
	results := make(chan IPScanEntry)

	var wg sync.WaitGroup

	for i := 0; i < maxParallel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
					results <- scanIP(ctx, j.IP, ports, timeout)
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, ip := range ips {
			select {
			case <-ctx.Done():
				return
			case jobs <- job{IP: ip}:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var out []IPScanEntry
	for r := range results {
		out = append(out, r)
	}

	return out, nil
}

func scanIP(ctx context.Context, ip net.IP, ports []int, timeout time.Duration) IPScanEntry {
	entry := IPScanEntry{
		IP:      ip,
		Method:  "tcp",
		PortOpen: false,
	}

	var responded bool
	var bestRTT time.Duration

	for _, port := range ports {
		select {
		case <-ctx.Done():
			return entry
		default:
		}

		addr := net.JoinHostPort(ip.String(), strconv.Itoa(port))
		dialer := net.Dialer{Timeout: timeout}

		start := time.Now()
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		rtt := time.Since(start)

		if err == nil {
			_ = conn.Close()
			responded = true
			entry.PortOpen = true

			if bestRTT == 0 || rtt < bestRTT {
				bestRTT = rtt
			}
			break
		}
	}

	entry.Responded = responded
	if responded {
		entry.RTT = bestRTT
	}
	return entry
}

func expandCIDR(ipNet *net.IPNet) []net.IP {
	var out []net.IP

	for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); ip = nextIP(ip) {
		tmp := make(net.IP, len(ip))
		copy(tmp, ip)
		out = append(out, tmp)
	}

	return out
}

func nextIP(ip net.IP) net.IP {
	ip = ip.To16()
	out := make(net.IP, len(ip))
	copy(out, ip)

	for i := len(out) - 1; i >= 0; i-- {
		out[i]++
		if out[i] != 0 {
			break
		}
	}
	return out
}
