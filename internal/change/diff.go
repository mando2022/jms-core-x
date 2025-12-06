// File: jms-core-x/internal/change/diff.go
// يحتوي على منطق المقارنة بين حالة قديمة وحالة جديدة للأجهزة.
// لا يتعامل مع DB أو IO خارجي؛ فقط مقارنة وبناء ChangeEvents.
package change

import (
    "jms-core-x/internal/core/final"
)

// deviceState هو تمثيل مبسط لحالة جهاز كما يراها Block 14.
// يتم بناؤه من FinalSnapshot و/أو SnapshotHistory.
type deviceState struct {
    DeviceID string
    MAC      string

    IPs       []string
    Hostnames []string
    Vendor    string
    Ports     []string
    Sources   []string
}

// diffDevice يقارن بين نسخة قديمة وجديدة من نفس الجهاز
// ويعيد قائمة ChangeEvent على مستوى الحقول فقط.
func diffDevice(oldDev, newDev deviceState) []ChangeEvent {
    events := make([]ChangeEvent, 0)

    deviceID := newDev.DeviceID
    if deviceID == "" {
        deviceID = oldDev.DeviceID
    }

    emit := func(eventType, field string, oldVal, newVal interface{}) {
        ce := buildChangeEvent(eventType, deviceID, field, oldVal, newVal)
        events = append(events, ce)
    }

    // Vendor
    if oldDev.Vendor != newDev.Vendor {
        emit(VENDOR_CHANGED, "vendor", safeInterface(oldDev.Vendor), safeInterface(newDev.Vendor))
    }

    // IPs
    if !equalStringSlice(oldDev.IPs, newDev.IPs) {
        emit(IP_CHANGED, "ips", safeInterface(oldDev.IPs), safeInterface(newDev.IPs))
    }

    // Hostnames
    if !equalStringSlice(oldDev.Hostnames, newDev.Hostnames) {
        emit(HOSTNAME_CHANGED, "hostnames", safeInterface(oldDev.Hostnames), safeInterface(newDev.Hostnames))
    }

    // Ports
    if !equalStringSlice(oldDev.Ports, newDev.Ports) {
        emit(PORTS_CHANGED, "ports", safeInterface(oldDev.Ports), safeInterface(newDev.Ports))
    }

    // SeenSources
    if !equalStringSlice(oldDev.Sources, newDev.Sources) {
        emit(SOURCES_CHANGED, "sources", safeInterface(oldDev.Sources), safeInterface(newDev.Sources))
    }

    return events
}

// diffSnapshots يقارن بين خريطتين من الأجهزة:
//   - prev: الحالة السابقة (من SnapshotHistory)
//   - curr: الحالة الحالية (من FinalSnapshot)
// ويعيد جميع الأحداث الناتجة (جديدة + مختفية + تغييرات حقول).
func diffSnapshots(prev, curr map[string]deviceState) []ChangeEvent {
    events := make([]ChangeEvent, 0)

    // 1) أجهزة جديدة أو محدثة.
    for id, newDev := range curr {
        oldDev, ok := prev[id]
        if !ok {
            // جهاز جديد بالكامل.
            ce := buildChangeEvent(DEVICE_NEW, id, "device", nil, safeInterface(newDev))
            events = append(events, ce)
            continue
        }

        // جهاز موجود مسبقًا → مقارنة الحقول.
        devEvents := diffDevice(oldDev, newDev)
        events = append(events, devEvents...)
    }

    // 2) أجهزة اختفت.
    for id, oldDev := range prev {
        if _, stillPresent := curr[id]; stillPresent {
            continue
        }
        ce := buildChangeEvent(DEVICE_DISAPPEARED, id, "device", safeInterface(oldDev), nil)
        events = append(events, ce)
    }

    return events
}

// diffFinalSnapshots يبني خريطتين من FinalSnapshot ثم يستدعي diffSnapshots.
func diffFinalSnapshots(prevSnap final.FinalSnapshot, currSnap final.FinalSnapshot) []ChangeEvent {
    prev := make(map[string]deviceState, len(prevSnap.Devices))
    curr := make(map[string]deviceState, len(currSnap.Devices))

    // بناء خريطة الحالة السابقة
    for _, d := range prevSnap.Devices {
        prev[d.DeviceID] = deviceState{
            DeviceID: d.DeviceID,
            MAC:      d.MAC,
            IPs:       d.IPs,
            Hostnames: d.Hostnames,
            Vendor:    d.Vendor,
            Ports:     d.Ports,
            Sources:   d.SeenSources,
        }
    }

    // بناء خريطة الحالة الحالية
    for _, d := range currSnap.Devices {
        curr[d.DeviceID] = deviceState{
            DeviceID: d.DeviceID,
            MAC:      d.MAC,
            IPs:       d.IPs,
            Hostnames: d.Hostnames,
            Vendor:    d.Vendor,
            Ports:     d.Ports,
            Sources:   d.SeenSources,
        }
    }

    return diffSnapshots(prev, curr)
}
