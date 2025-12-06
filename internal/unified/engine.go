package unified

import (
	"sort"
	"strings"
	"sync"
)

// Source represents the origin of a piece of discovery information.
type Source string

const (
	SourceLLDP    Source = "lldp"
	SourceUBNT    Source = "ubnt"
	SourceIPScan  Source = "ipscan"
	SourceTrusted Source = "trusted"
)

// FieldName represents a logical field on a device that can be populated
// from multiple discovery sources.
type FieldName string

const (
	FieldIP     FieldName = "ip"
	FieldVendor FieldName = "vendor"
	FieldNameID FieldName = "name"
)

// ValueWithSource binds a value to the discovery source that produced it.
type ValueWithSource struct {
	Value  string
	Source Source
}

// Conflict describes a disagreement between two sources about the value
// of a specific field on a device.
type Conflict struct {
	DeviceID string
	Field    FieldName
	Current  ValueWithSource
	Incoming ValueWithSource
}

// UnifiedDevice is the normalized representation of a device as exposed
// to the rest of the system and to the dashboard.
type UnifiedDevice struct {
	ID      string
	MAC     string
	IP      string
	Vendor  string
	Name    string
	Sources []Source
}

// LLDPInput contains the subset of LLDP data that the Unified Engine cares about.
type LLDPInput struct {
	MAC    string
	IP     string // management IP if available
	Name   string // system / host name
	Vendor string // derived vendor if available
}

// UBNTInput contains normalized fields from the UBNT discovery engine.
type UBNTInput struct {
	MAC    string
	IP     string
	Name   string // device name
	Vendor string // vendor / model info
}

// IPScanInput contains data coming from the IP scan engine.
type IPScanInput struct {
	MAC    string
	IP     string
	Name   string // reverse DNS / heuristic name
	Vendor string // vendor inferred from MAC / fingerprinting
}

// TrustedInput contains trusted overrides coming from the Trusted layer.
// All fields are optional; non-empty values will participate in the
// priority rules.
type TrustedInput struct {
	MAC    string
	IP     string
	Name   string
	Vendor string
}

// Snapshot represents a read-only view of the current engine state.
type Snapshot struct {
	Devices   []UnifiedDevice
	Conflicts []Conflict
}

// Engine maintains an in-memory, concurrency-safe view of all discovered
// devices and applies the priority rules described in Block 5.
type Engine struct {
	mu        sync.RWMutex
	devices   map[string]*deviceState
	conflicts []Conflict
}

type deviceState struct {
	id      string
	mac     string
	ip      fieldState
	vendor  fieldState
	name    fieldState
	sources map[Source]bool
}

type fieldState struct {
	ValueWithSource
	priority int
}

// NewEngine constructs an empty Unified Engine instance.
func NewEngine() *Engine {
	return &Engine{
		devices: make(map[string]*deviceState),
	}
}

// ApplyLLDP ingests a single LLDP record into the engine.
func (e *Engine) ApplyLLDP(in LLDPInput) {
	e.apply(SourceLLDP, in.MAC, in.IP, in.Vendor, in.Name)
}

// ApplyUBNT ingests a single UBNT discovery record into the engine.
func (e *Engine) ApplyUBNT(in UBNTInput) {
	e.apply(SourceUBNT, in.MAC, in.IP, in.Vendor, in.Name)
}

// ApplyIPScan ingests a single IP-scan discovery record into the engine.
func (e *Engine) ApplyIPScan(in IPScanInput) {
	e.apply(SourceIPScan, in.MAC, in.IP, in.Vendor, in.Name)
}

// ApplyTrusted ingests a trusted device definition / override into the engine.
// Trusted data follows the same merge rules but has higher priority for
// specific fields as defined in the Block 5 spec.
func (e *Engine) ApplyTrusted(in TrustedInput) {
	e.apply(SourceTrusted, in.MAC, in.IP, in.Vendor, in.Name)
}

// Snapshot returns a stable snapshot of the current engine state, including
// the list of unified devices and all recorded conflicts.
func (e *Engine) Snapshot() Snapshot {
	e.mu.RLock()
	defer e.mu.RUnlock()

	devices := make([]UnifiedDevice, 0, len(e.devices))
	for _, ds := range e.devices {
		var srcs []Source
		for s := range ds.sources {
			srcs = append(srcs, s)
		}
		sort.Slice(srcs, func(i, j int) bool {
			return srcs[i] < srcs[j]
		})

		devices = append(devices, UnifiedDevice{
			ID:      ds.id,
			MAC:     ds.mac,
			IP:      ds.ip.Value,
			Vendor:  ds.vendor.Value,
			Name:    ds.name.Value,
			Sources: srcs,
		})
	}

	// Sort devices by ID for a deterministic snapshot.
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].ID < devices[j].ID
	})

	conflicts := make([]Conflict, len(e.conflicts))
	copy(conflicts, e.conflicts)

	return Snapshot{
		Devices:   devices,
		Conflicts: conflicts,
	}
}

// apply is the core merge function used by all discovery sources.
func (e *Engine) apply(src Source, mac, ip, vendor, name string) {
	key := identityKey(mac, ip)
	if key == "" {
		// بدون هوية (MAC/IP) لا يمكننا إنشاء جهاز موثوق.
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	ds, ok := e.devices[key]
	if !ok {
		ds = &deviceState{
			id:      key,
			mac:     mac,
			sources: make(map[Source]bool),
		}
		e.devices[key] = ds
	}
	ds.sources[src] = true

	// تحديث الحقول الأساسية مع احترام قواعد الأولوية.
	e.updateFieldLocked(ds, FieldIP, src, &ds.ip, ip)
	e.updateFieldLocked(ds, FieldVendor, src, &ds.vendor, vendor)
	e.updateFieldLocked(ds, FieldNameID, src, &ds.name, name)
}

func (e *Engine) updateFieldLocked(ds *deviceState, field FieldName, src Source, st *fieldState, val string) {
	if val == "" {
		return
	}
	priority := fieldPriority(field, src)

	// إذا لم يكن هناك قيمة حالية، نستخدم القادمة مباشرة.
	if st.Value == "" {
		st.Value = val
		st.Source = src
		st.priority = priority
		return
	}

	// نفس القيمة: يمكننا فقط ترقية الـ Source إذا كانت أولويته أعلى.
	if st.Value == val {
		if priority > st.priority {
			st.Source = src
			st.priority = priority
		}
		return
	}

	// قيمة مختلفة → تعارض محتمل.
	current := ValueWithSource{Value: st.Value, Source: st.Source}
	incoming := ValueWithSource{Value: val, Source: src}

	// لو الأولوية أعلى، نحدّث القيمة لكن نحتفظ بسجل التعارض.
	if priority > st.priority {
		e.conflicts = append(e.conflicts, Conflict{
			DeviceID: ds.id,
			Field:    field,
			Current:  current,
			Incoming: incoming,
		})
		st.Value = val
		st.Source = src
		st.priority = priority
		return
	}

	// لو الأولوية أقل أو مساوية، نسجل التعارض لكن لا نغيّر القيمة الحالية.
	e.conflicts = append(e.conflicts, Conflict{
		DeviceID: ds.id,
		Field:    field,
		Current:  current,
		Incoming: incoming,
	})
}

// identityKey يبني مفتاح داخلي موحّد للجهاز اعتمادًا على MAC أو IP.
// يفضّل MAC إن وجد، وبعده IP.
func identityKey(mac, ip string) string {
	if mac != "" {
		return "mac:" + normalizeMAC(mac)
	}
	if ip != "" {
		return "ip:" + strings.TrimSpace(strings.ToLower(ip))
	}
	return ""
}

// normalizeMAC يحوّل الـ MAC لشكل موحّد (بدون فواصل وبحروف صغيرة).
func normalizeMAC(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ':' || c == '-' || c == '.' || c == ' ' {
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}

// fieldPriority يطبّق قواعد الأولوية لكل حقل حسب Block 5 spec.
//
//  * IP:     trusted > ubnt > lldp > ipscan
//  * Vendor: ubnt > lldp > ipscan
//  * Name:   trusted > other sources
func fieldPriority(field FieldName, src Source) int {
	switch field {
	case FieldIP:
		switch src {
		case SourceTrusted:
			return 40
		case SourceUBNT:
			return 30
		case SourceLLDP:
			return 20
		case SourceIPScan:
			return 10
		default:
			return 0
		}
	case FieldVendor:
		switch src {
		case SourceUBNT:
			return 30
		case SourceLLDP:
			return 20
		case SourceIPScan:
			return 10
		case SourceTrusted:
			// في بعض الحالات قد يأتي Vendor من trusted، نمنحه أعلى أولوية.
			return 40
		default:
			return 0
		}
	case FieldNameID:
		switch src {
		case SourceTrusted:
			return 40
		case SourceUBNT:
			return 20
		case SourceLLDP:
			return 15
		case SourceIPScan:
			return 10
		default:
			return 0
		}
	default:
		return 0
	}
}
