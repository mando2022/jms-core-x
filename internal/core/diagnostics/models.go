package diagnostics

import "time"

// PingSample represents a single probe result in a diagnostics session.
type PingSample struct {
	Time    time.Time
	Success bool
	RTT     time.Duration
}

// PingSummary aggregates statistics for a diagnostics session.
type PingSummary struct {
	Duration    time.Duration
	Sent        int
	Received    int
	Lost        int
	LossPercent float64
	MinRTT      time.Duration
	MaxRTT      time.Duration
	AvgRTT      time.Duration
}

// DeviceDiagnostics keeps the last known diagnostics snapshot for a device.
type DeviceDiagnostics struct {
	DeviceID   string
	IP         string
	LastPing   PingSummary
	LastUpdate time.Time
}
