package diagnostics

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ProbeFunc represents the underlying probe used by diagnostics.
type ProbeFunc func(ctx context.Context, ip string) (rtt time.Duration, ok bool)

// SessionListener is a stream of PingSample values for a running session.
type SessionListener chan PingSample

// PingSession represents a single manual diagnostics session for a device.
type PingSession struct {
	DeviceID string
	IP       string

	Duration time.Duration
	Interval time.Duration

	mu        sync.RWMutex
	Samples   []PingSample
	Summary   PingSummary
	listeners map[SessionListener]struct{}

	done   chan struct{}
	ctx    context.Context
	cancel context.CancelFunc
}

func (s *PingSession) Done() <-chan struct{} {
	return s.done
}


func (s *PingSession) addListener(ch SessionListener) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listeners == nil {
		s.listeners = make(map[SessionListener]struct{})
	}
	s.listeners[ch] = struct{}{}
}

func (s *PingSession) removeListener(ch SessionListener) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.listeners, ch)
}

func (s *PingSession) broadcastSample(sample PingSample) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for ch := range s.listeners {
		select {
		case ch <- sample:
		default:
		}
	}
}

func (s *PingSession) closeListeners() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for ch := range s.listeners {
		close(ch)
		delete(s.listeners, ch)
	}
}

// SessionManager manages the lifecycle of diagnostics sessions.
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*PingSession
	last     map[string]PingSummary

	probe ProbeFunc
}

func NewSessionManager(probe ProbeFunc) *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*PingSession),
		last:     make(map[string]PingSummary),
		probe:    probe,
	}
}

// StartSession starts a new diagnostics session for a device.
func (m *SessionManager) StartSession(
	ctx context.Context,
	deviceID, ip string,
	duration, interval time.Duration,
) (*PingSession, error) {
	if deviceID == "" {
		return nil, errors.New("diagnostics: empty deviceID")
	}
	if ip == "" {
		return nil, errors.New("diagnostics: empty IP")
	}
	if m.probe == nil {
		return nil, errors.New("diagnostics: probe function is nil")
	}

	if duration <= 0 {
		duration = 60 * time.Second
	}
	if interval <= 0 {
		interval = time.Second
	}

	m.mu.Lock()
	if _, exists := m.sessions[deviceID]; exists {
		m.mu.Unlock()
		return nil, fmt.Errorf("diagnostics: session already running for device %q", deviceID)
	}

	sessionCtx, cancel := context.WithCancel(ctx)
	s := &PingSession{
		DeviceID: deviceID,
		IP:       ip,
		Duration: duration,
		Interval: interval,
		done:     make(chan struct{}),
		ctx:      sessionCtx,
		cancel:   cancel,
	}
	m.sessions[deviceID] = s
	m.mu.Unlock()

	go m.runSession(s)

	return s, nil
}

func (m *SessionManager) runSession(s *PingSession) {
	defer close(s.done)

	start := time.Now()
	deadline := start.Add(s.Duration)

	for {
		now := time.Now()
		if !deadline.IsZero() && now.After(deadline) {
			break
		}

		select {
		case <-s.ctx.Done():
			m.finishSession(s)
			return
		default:
		}

		rtt, ok := m.probe(s.ctx, s.IP)
		sample := PingSample{
			Time:    time.Now(),
			Success: ok,
			RTT:     rtt,
		}

		s.mu.Lock()
		s.Samples = append(s.Samples, sample)
		s.mu.Unlock()

		s.broadcastSample(sample)

		select {
		case <-s.ctx.Done():
			m.finishSession(s)
			return
		case <-time.After(s.Interval):
		}
	}

	m.finishSession(s)
}

func (m *SessionManager) finishSession(s *PingSession) {
	s.mu.Lock()
	s.Summary = summarizeSamples(s.Duration, s.Samples)
	summary := s.Summary
	s.mu.Unlock()

	m.mu.Lock()
	m.last[s.DeviceID] = summary
	delete(m.sessions, s.DeviceID)
	m.mu.Unlock()

	s.closeListeners()
	if s.cancel != nil {
		s.cancel()
	}
}

func summarizeSamples(duration time.Duration, samples []PingSample) PingSummary {
	sent := len(samples)
	var received int
	var minRTT, maxRTT, totalRTT time.Duration

	for _, sample := range samples {
		if !sample.Success {
			continue
		}
		received++
		if minRTT == 0 || sample.RTT < minRTT {
			minRTT = sample.RTT
		}
		if sample.RTT > maxRTT {
			maxRTT = sample.RTT
		}
		totalRTT += sample.RTT
	}

	lost := sent - received
	var avgRTT time.Duration
	var lossPercent float64

	if received > 0 {
		avgRTT = time.Duration(int64(totalRTT) / int64(received))
	}
	if sent > 0 {
		lossPercent = float64(lost) * 100.0 / float64(sent)
	}

	return PingSummary{
		Duration:    duration,
		Sent:        sent,
		Received:    received,
		Lost:        lost,
		LossPercent: lossPercent,
		MinRTT:      minRTT,
		MaxRTT:      maxRTT,
		AvgRTT:      avgRTT,
	}
}

func (m *SessionManager) LastSummary(deviceID string) (PingSummary, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	summary, ok := m.last[deviceID]
	return summary, ok
}

func (m *SessionManager) Subscribe(deviceID string) (SessionListener, error) {
	m.mu.RLock()
	session, ok := m.sessions[deviceID]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("diagnostics: no active session for device %q", deviceID)
	}

	ch := make(SessionListener, 16)
	session.addListener(ch)
	return ch, nil
}

func (m *SessionManager) Unsubscribe(deviceID string, ch SessionListener) {
	m.mu.RLock()
	session, ok := m.sessions[deviceID]
	m.mu.RUnlock()
	if !ok {
		return
	}
	session.removeListener(ch)
}

func (m *SessionManager) StopAll() {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, s := range m.sessions {
		if s.cancel != nil {
			s.cancel()
		}
	}
}
