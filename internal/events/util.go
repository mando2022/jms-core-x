package events

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"
)

// يحوّل time.Time إلى Unix seconds
func parseTimestamp(t time.Time) int64 {
	return t.Unix()
}

// يحوّل Unix seconds إلى time.Time
func fromTimestamp(ts int64) time.Time {
	return time.Unix(ts, 0).UTC()
}

// تنظيف نصوص قبل إدخالها DB
func safeDBValue(s string) string {
	return strings.TrimSpace(s)
}

// توليد EventID ثابت
func generateEventID(
	deviceID, snapshotID, eventType, field, oldValue, newValue string,
	t time.Time,
) (string, error) {

	payload := map[string]interface{}{
		"device_id":   deviceID,
		"snapshot_id": snapshotID,
		"event_type":  eventType,
		"field":       field,
		"old_value":   oldValue,
		"new_value":   newValue,
		"timestamp":   t.UTC().Unix(),
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}
