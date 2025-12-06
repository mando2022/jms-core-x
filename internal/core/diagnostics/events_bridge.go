package diagnostics

import "context"

// Event is a placeholder type for device events.
//
// The concrete shape of events will be defined in later blocks (Events Layer).
// Using interface{} here keeps the diagnostics interface flexible without
// introducing extra assumptions in Block 8.
type Event interface{}

// EventsReader provides read-only access to device events.
//
// Block 8 only defines this integration interface; the actual event store and
// implementations are added by later blocks.
type EventsReader interface {
	DeviceEvents(ctx context.Context, deviceID string, limit int) ([]Event, error)
}
