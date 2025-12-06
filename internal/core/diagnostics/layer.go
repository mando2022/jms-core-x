package diagnostics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"jms-core-x/internal/core"
	"jms-core-x/internal/unified"
)

type UnifiedSnapshotProvider interface {
	UnifiedSnapshot(ctx context.Context) unified.Snapshot
}

type DiagnosticsLayer struct {
	core.BaseLayer

	unifiedProvider UnifiedSnapshotProvider
	sessionMgr      *SessionManager
	watchlist       *Watchlist
	eventsReader    EventsReader

	mu          sync.RWMutex
	diagnostics map[string]DeviceDiagnostics
}

func NewDiagnosticsLayer(
	unifiedProvider UnifiedSnapshotProvider,
	probe ProbeFunc,
	eventsReader EventsReader,
) *DiagnosticsLayer {
	return &DiagnosticsLayer{
		BaseLayer:       core.NewBaseLayer(core.LayerDiagnostics),
		unifiedProvider: unifiedProvider,
		sessionMgr:      NewSessionManager(probe),
		watchlist:       NewWatchlist(),
		eventsReader:    eventsReader,
		diagnostics:     make(map[string]DeviceDiagnostics),
	}
}

func (l *DiagnosticsLayer) Init(ctx context.Context, cfg core.Config) error {
	return nil
}

func (l *DiagnosticsLayer) Start(ctx context.Context) error {
	return nil
}

func (l *DiagnosticsLayer) Stop(ctx context.Context) error {
	if l.sessionMgr != nil {
		l.sessionMgr.StopAll()
	}
	return nil
}

func (l *DiagnosticsLayer) StartDevicePing(
	ctx context.Context,
	deviceID string,
	duration, interval time.Duration,
) (*PingSession, error) {
	if l.unifiedProvider == nil {
		return nil, fmt.Errorf("diagnostics: unified provider is nil")
	}

	snap := l.unifiedProvider.UnifiedSnapshot(ctx)
	device := findDeviceByID(snap, deviceID)
	if device == nil || device.IP == "" {
		return nil, fmt.Errorf("diagnostics: device %q has no IP", deviceID)
	}

	if l.sessionMgr == nil {
		return nil, fmt.Errorf("diagnostics: session manager is nil")
	}

	session, err := l.sessionMgr.StartSession(ctx, deviceID, device.IP, duration, interval)
	if err != nil {
		return nil, err
	}

	go l.trackSession(deviceID, device.IP, session)

	return session, nil
}

func (l *DiagnosticsLayer) LastSummary(deviceID string) (PingSummary, bool) {
	if l.sessionMgr == nil {
		return PingSummary{}, false
	}
	return l.sessionMgr.LastSummary(deviceID)
}

func (l *DiagnosticsLayer) DeviceDiagnostics(deviceID string) (DeviceDiagnostics, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	diag, ok := l.diagnostics[deviceID]
	return diag, ok
}

func (l *DiagnosticsLayer) AddToWatchlist(deviceID, note string) {
	l.watchlist.Add(deviceID, note)
}

func (l *DiagnosticsLayer) RemoveFromWatchlist(deviceID string) {
	l.watchlist.Remove(deviceID)
}

func (l *DiagnosticsLayer) ListWatchlist() []WatchlistEntry {
	return l.watchlist.List()
}

func (l *DiagnosticsLayer) DeviceEvents(
	ctx context.Context,
	deviceID string,
	limit int,
) ([]Event, error) {
	if l.eventsReader == nil {
		return nil, nil
	}
	return l.eventsReader.DeviceEvents(ctx, deviceID, limit)
}

//
// ------------------------------------------------------------
// 🔥 إضافات Block 8 الناقصة — Subscribe / Unsubscribe
// ------------------------------------------------------------
//

func (l *DiagnosticsLayer) Subscribe(deviceID string) (SessionListener, error) {
	if l.sessionMgr == nil {
		return nil, fmt.Errorf("diagnostics: session manager is nil")
	}
	return l.sessionMgr.Subscribe(deviceID)
}

func (l *DiagnosticsLayer) Unsubscribe(deviceID string, ch SessionListener) {
	if l.sessionMgr == nil {
		return
	}
	l.sessionMgr.Unsubscribe(deviceID, ch)
}

//
// ------------------------------------------------------------
// Session Tracking
// ------------------------------------------------------------
//

func (l *DiagnosticsLayer) trackSession(deviceID, ip string, s *PingSession) {
	if s == nil {
		return
	}

	<-s.Done()

	if l.sessionMgr == nil {
		return
	}

	summary, ok := l.sessionMgr.LastSummary(deviceID)
	if !ok {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	if l.diagnostics == nil {
		l.diagnostics = make(map[string]DeviceDiagnostics)
	}

	l.diagnostics[deviceID] = DeviceDiagnostics{
		DeviceID:   deviceID,
		IP:         ip,
		LastPing:   summary,
		LastUpdate: time.Now(),
	}
}

func findDeviceByID(snap unified.Snapshot, deviceID string) *unified.UnifiedDevice {
	for i := range snap.Devices {
		if snap.Devices[i].ID == deviceID {
			return &snap.Devices[i]
		}
	}
	return nil
}
