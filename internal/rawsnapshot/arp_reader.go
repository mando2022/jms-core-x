// File: jms-core-x/internal/rawsnapshot/arp_reader.go
package rawsnapshot

import (
	"bufio"
	"context"
	"net"
	"os/exec"
	"strings"
)

type SystemARPReader struct{}

func (r *SystemARPReader) Read(ctx context.Context) ([]ARPEntry, error) {
	cmd := exec.CommandContext(ctx, "arp", "-a")
	raw, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var out []ARPEntry
	scanner := bufio.NewScanner(strings.NewReader(string(raw)))

	var currentIface string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "Interface:") {
			currentIface = parseInterfaceLine(line)
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		ip := net.ParseIP(fields[0])
		if ip == nil {
			continue
		}

		macNorm := strings.ReplaceAll(fields[1], "-", ":")
		mac, err := net.ParseMAC(macNorm)
		if err != nil {
			continue
		}

		out = append(out, ARPEntry{
			IP:        ip,
			MAC:       mac,
			Interface: currentIface,
			Source:    "arp-table",
		})
	}

	if err := scanner.Err(); err != nil {
		return out, err
	}

	return out, nil
}

func parseInterfaceLine(line string) string {
	line = strings.TrimPrefix(line, "Interface:")
	line = strings.TrimSpace(line)
	if idx := strings.Index(line, "---"); idx >= 0 {
		line = strings.TrimSpace(line[:idx])
	}
	return line
}
