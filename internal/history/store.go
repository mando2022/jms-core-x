// File: internal/history/store.go
package history

import (
    "context"
    "database/sql"
    "errors"
    "fmt"

    "jms-core-x/internal/core/final"
)

type Store struct {
    db *sql.DB
}

func NewStore(db *sql.DB) *Store {
    return &Store{db: db}
}

// تخزين Snapshot في snapshots_history
func (s Store) StoreSnapshotHistory(ctx context.Context, snap final.FinalSnapshot) (string, error) {
    if s.db == nil {
        return "", errors.New("history: nil db")
    }

    raw, err := encodeSnapshot(snap)
    if err != nil {
        return "", fmt.Errorf("encode snapshot error: %w", err)
    }

    id := generateSnapshotID(raw)

    _, err = s.db.ExecContext(
        ctx,
        "INSERT INTO "+snapshotsHistoryTable+" (snapshot_id, timestamp, raw_json) VALUES (?, ?, ?)",
        id,
        snap.Time,
        raw,
    )
    if err != nil {
        return "", fmt.Errorf("insert snapshot_history failed: %w", err)
    }

    return id, nil
}

// قراءة آخر Snapshot
func (s Store) LoadLastSnapshot() (SnapshotHistory, error) {
    if s.db == nil {
        return SnapshotHistory{}, errors.New("nil db")
    }

    row := s.db.QueryRow(
        "SELECT snapshot_id, timestamp, raw_json FROM "+snapshotsHistoryTable+" ORDER BY timestamp DESC LIMIT 1",
    )

    var out SnapshotHistory
    if err := row.Scan(&out.SnapshotID, &out.Timestamp, &out.RawJSON); err != nil {
        return SnapshotHistory{}, err
    }

    return out, nil
}

// تحديث تاريخ جهاز واحد
func (s Store) UpdateDeviceHistory(ctx context.Context, snap final.FinalSnapshot) error {
    if s.db == nil {
        return errors.New("nil db")
    }

    for _, d := range snap.Devices {
        _, err := s.db.ExecContext(
            ctx,
            `
            INSERT INTO `+devicesHistoryTable+` 
                (device_id, first_seen, last_seen, seen_count, mac, vendor)
            VALUES (?, ?, ?, ?, ?, ?)
            ON CONFLICT(device_id) DO UPDATE SET
                last_seen = excluded.last_seen,
                seen_count = devices_history.seen_count + 1,
                mac = excluded.mac,
                vendor = excluded.vendor
            `,
            d.DeviceID,
            snap.Time,
            snap.Time,
            1,
            d.MAC,
            d.Vendor,
        )
        if err != nil {
            return fmt.Errorf("update device history failed: %w", err)
        }
    }

    return nil
}

// كل الأجهزة
func (s Store) DeviceHistory(ctx context.Context) ([]DeviceHistory, error) {
    rows, err := s.db.QueryContext(
        ctx,
        `
        SELECT device_id, first_seen, last_seen, seen_count, mac, vendor
        FROM `+devicesHistoryTable+`
        ORDER BY last_seen DESC
        `,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    list := []DeviceHistory{}
    for rows.Next() {
        var d DeviceHistory
        if err := rows.Scan(&d.DeviceID, &d.FirstSeen, &d.LastSeen, &d.SeenCount, &d.MAC, &d.Vendor); err != nil {
            return nil, err
        }
        list = append(list, d)
    }

    return list, nil
}

// جهاز واحد
func (s Store) LoadDeviceHistory(deviceID string) (DeviceHistory, error) {
    row := s.db.QueryRow(
        `
        SELECT device_id, first_seen, last_seen, seen_count, mac, vendor
        FROM `+devicesHistoryTable+`
        WHERE device_id = ?
        `,
        deviceID,
    )

    var d DeviceHistory
    if err := row.Scan(&d.DeviceID, &d.FirstSeen, &d.LastSeen, &d.SeenCount, &d.MAC, &d.Vendor); err != nil {
        return DeviceHistory{}, err
    }

    return d, nil
}

// كتابة Timeline
func (s Store) AddTimelineEntries(ctx context.Context, snap final.FinalSnapshot) error {
    for _, d := range snap.Devices {
        for _, src := range d.SeenSources {
            _, err := s.db.ExecContext(
                ctx,
                `
                INSERT INTO `+devicesTimelineTable+` (device_id, timestamp, source)
                VALUES (?, ?, ?)
                `,
                d.DeviceID,
                snap.Time,
                safeDBValue(src),
            )
            if err != nil {
                return err
            }
        }
    }
    return nil
}

// قراءة Timeline
func (s Store) DeviceTimeline(ctx context.Context, deviceID string) ([]DeviceTimeline, error) {
    rows, err := s.db.QueryContext(
        ctx,
        `
        SELECT device_id, timestamp, source
        FROM `+devicesTimelineTable+`
        WHERE device_id = ?
        ORDER BY timestamp DESC
        `,
        deviceID,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    list := []DeviceTimeline{}
    for rows.Next() {
        var e DeviceTimeline
        if err := rows.Scan(&e.DeviceID, &e.Timestamp, &e.Source); err != nil {
            return nil, err
        }
        list = append(list, e)
    }

    return list, nil
}
