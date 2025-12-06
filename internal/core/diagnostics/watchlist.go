package diagnostics

import (
	"sort"
	"sync"
	"time"
)

type WatchlistEntry struct {
	DeviceID string
	AddedAt  time.Time
	Note     string
}

type Watchlist struct {
	mu      sync.RWMutex
	entries map[string]WatchlistEntry
}

func NewWatchlist() *Watchlist {
	return &Watchlist{
		entries: make(map[string]WatchlistEntry),
	}
}

func (w *Watchlist) Add(deviceID, note string) {
	if deviceID == "" {
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.entries == nil {
		w.entries = make(map[string]WatchlistEntry)
	}

	entry, ok := w.entries[deviceID]
	if !ok {
		entry = WatchlistEntry{
			DeviceID: deviceID,
			AddedAt:  time.Now(),
		}
	}
	entry.Note = note
	w.entries[deviceID] = entry
}

func (w *Watchlist) Remove(deviceID string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.entries == nil {
		return
	}
	delete(w.entries, deviceID)
}

func (w *Watchlist) List() []WatchlistEntry {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if len(w.entries) == 0 {
		return nil
	}

	out := make([]WatchlistEntry, 0, len(w.entries))
	for _, e := range w.entries {
		out = append(out, e)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].AddedAt.Before(out[j].AddedAt)
	})

	return out
}
