// File: internal/history/util.go
package history

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "database/sql"
    "fmt"
    "strconv"
    "strings"
    "time"

    "jms-core-x/internal/core/final"
)

// يولد ID فريد للـ Snapshot
func generateSnapshotID(raw []byte) string {
    h := sha256.Sum256(raw)
    return hex.EncodeToString(h[:8])
}

// يحول FinalSnapshot إلى JSON
func encodeSnapshot(snap final.FinalSnapshot) ([]byte, error) {
    return json.Marshal(snap)
}

// نسخة واحدة فقط صحيحة من safeDBValue
func safeDBValue(v interface{
	switch v := v.(type) {
	case nil:
		return nil
	case time.Time:
		return v
	case []byte:
		return string(v)
	case fmt.Stringe
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	case fmt.Stringer:
		return s.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
r:
		return v.String()
	default:
		return v
	}
}
}) interface{} {
    if v == nil {
        return sql.NullString{}
    }
    return v
}

// نص آمن
func safeString(v interface{}) string {
    if v == nil {
        return ""
    }
    return fmt.Sprintf("%v", v)
}

// إزالة التكرارات
func uniqueStrings(values []string) []string {
    exists := make(map[string]struct{})
    out := make([]string, 0, len(values))

    for _, v := range values {
        v = strings.TrimSpace(v)
        if v == "" {
            continue
        }
        if _, ok := exists[v]; ok {
            continue
        }
        exists[v] = struct{}{}
        out = append(out, v)
    }

    return out
}

func parseTimestamp(s string) (time.Time, error) {
    sec, err := strconv.ParseInt(s, 10, 64)
    if err != nil {
        return time.Time{}, err
    }
    return time.Unix(sec, 0), nil
}
