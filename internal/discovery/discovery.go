// File: jms-core-x/internal/discovery/discovery.go
// Block 6 – Discovery Layer Base (Data Collection Only)

package discovery

import (
	"context"
)

// RawInput هو المخرج الرسمي لـ Block 6 كما في Block6_Blueprint.md
// هذا الـ struct سيتم استهلاكه لاحقاً في Block 7 (RawSnapshot Engine).
type RawInput struct {
	LLDP []LLDPDevice
	UBNT []UBNTDevice
}

// Options تمثل إعدادات تشغيل مصادر الاكتشاف على Windows + Npcap.
// ملاحظة: Block 6 لا يحمل مسئولية تحميل الـ config من الملفات أو الـ DB.
// الطبقات الأعلى هي اللي هتمرِّر القيم دي.
type Options struct {
	// NPFName هو اسم Npcap الحقيقي للواجهة
	// مثال: \Device\NPF_{XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX}
	NPFName string

	// InterfaceName هو الاسم الظاهر للمستخدم (Display Name)
	// مثال: "Ethernet", "Ethernet 2", "Wi-Fi"
	InterfaceName string
}

// Manager هو قلب Block 6 داخل JMS Core X.
// مسئول عن:
//
//   - تشغيل LLDPSource و UBNTSource في Goroutines منفصلة.
//   - إيقاف المصادر بشكل نظيف.
//   - إنتاج RawInput يحتوي على:
//       * قائمة LLDPDevice
//       * قائمة UBNTDevice
//
// Block 6 لا يقوم بأي دمج أو Normalize أو Unified Snapshot.
type Manager struct {
	opts Options

	lldp *LLDPSource
	ubnt *UBNTSource
}

// NewManager ينشئ Discovery Manager جديد بمصادر الاكتشاف المطلوبة.
func NewManager(opts Options) *Manager {
	return &Manager{
		opts: opts,
		lldp: NewLLDPSource(opts.NPFName, opts.InterfaceName),
		ubnt: NewUBNTSource(opts.InterfaceName),
	}
}

// Start يقوم بتشغيل LLDP + UBNT Sources.
//
// ملاحظات مهمة حسب Block6 Blueprint:
//   - Block 6 = Data Collection Only.
//   - لا يتم استدعاء أي كود من unified.Engine هنا.
//   - لا يوجد DB Write أو API هنا.
func (m *Manager) Start(ctx context.Context) error {
	if m.lldp != nil {
		if err := m.lldp.Start(ctx); err != nil {
			return err
		}
	}

	if m.ubnt != nil {
		if err := m.ubnt.Start(ctx); err != nil {
			// محاولة إيقاف LLDP في حالة فشل UBNT
			m.lldp.Stop()
			return err
		}
	}

	return nil
}

// Stop يوقِف جميع مصادر الاكتشاف (LLDP + UBNT) بشكل نظيف.
func (m *Manager) Stop() {
	if m.lldp != nil {
		m.lldp.Stop()
	}
	if m.ubnt != nil {
		m.ubnt.Stop()
	}
}

// Snapshot يعيد RawInput يحتوي على:
//
//   - LLDP: snapshot للأجهزة المكتشفة عبر LLDP
//   - UBNT: snapshot للأجهزة المكتشفة عبر بروتوكول UBNT
//
// بدون دمج أو Normalize أو Identity Resolution.
func (m *Manager) Snapshot() RawInput {
	var out RawInput
	if m.lldp != nil {
		out.LLDP = m.lldp.Snapshot()
	}
	if m.ubnt != nil {
		out.UBNT = m.ubnt.Snapshot()
	}
	return out
}
