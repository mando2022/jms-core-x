// File: jms-core-x/internal/discovery/lldp_source.go
// Block 6 – LLDP Adapter (Data Collection Only)

package discovery

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// LLDPDevice يمثل جهاز تم اكتشافه عن طريق LLDP فقط.
// هذه هي البيانات الخام التي سيتم استهلاكها لاحقاً في Block 7/8.
type LLDPDevice struct {
	MAC       string    // MAC الخاص بالجهاز (من Ethernet Source MAC أو TLV)
	IP        string    // Management IP إن وُجد في TLVs
	Name      string    // System Name / Hostname
	PortID    string    // Port ID من TLVs
	ChassisID string    // Chassis ID من TLVs
	Interface string    // اسم الواجهة الظاهر للمستخدم (Display Name)
	LastSeen  time.Time // آخر مرة شوهد فيها الـ LLDP frame
}

// LLDPSource مسئول عن:
//   - فتح pcap handle على Windows Npcap interface.
//   - تطبيق BPF filter: ether proto 0x88cc
//   - قراءة LLDP frames.
//   - تحليل TLVs إلى LLDPDevice.
//   - حفظ النتائج في TTL map.
//   - توفير Snapshot للأجهزة الحية.
type LLDPSource struct {
	npfName   string
	ifaceName string

	mu      sync.Mutex
	devices map[string]LLDPDevice // keyed by MAC
	ttl     time.Duration

	cancel context.CancelFunc
	running bool
}

// NewLLDPSource ينشئ مصدر LLDP جديد على واجهة واحدة.
func NewLLDPSource(npfName, ifaceName string) *LLDPSource {
	return &LLDPSource{
		npfName:   npfName,
		ifaceName: ifaceName,
		devices:   make(map[string]LLDPDevice),
		ttl:       2 * time.Minute,
	}
}

// Start يشغّل Goroutine لالتقاط LLDP frames على الواجهة المحددة.
//
// ملاحظة: ctx القادم من الطبقة الأعلى هو المسئول عن إيقاف التشغيل.
func (s *LLDPSource) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.mu.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	go s.run(ctx)

	return nil
}

// Stop يوقِف LLDP sniffer بشكل نظيف.
func (s *LLDPSource) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	cancel := s.cancel
	s.cancel = nil
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}

// Snapshot يعيد الأجهزة التي ما زالت ضمن TTL.
// لا يقوم بأي دمج أو Normalize، فقط raw devices.
func (s *LLDPSource) Snapshot() []LLDPDevice {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]LLDPDevice, 0, len(s.devices))
	for mac, dev := range s.devices {
		if now.Sub(dev.LastSeen) > s.ttl {
			delete(s.devices, mac)
			continue
		}
		out = append(out, dev)
	}

	return out
}

// run هو الـ loop الداخلي لالتقاط LLDP frames من Npcap.
func (s *LLDPSource) run(ctx context.Context) {
	handle, err := pcap.OpenLive(s.npfName, 1600, true, pcap.BlockForever)
	if err != nil {
		log.Printf("[LLDP] pcap error on %s: %v\n", s.npfName, err)
		return
	}
	defer handle.Close()

	const filter = "ether proto 0x88cc" // LLDP only
	if err := handle.SetBPFFilter(filter); err != nil {
		log.Printf("[LLDP] BPF error on %s: %v\n", s.npfName, err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packetSource.NoCopy = true

	for {
		select {
		case <-ctx.Done():
			log.Println("[LLDP] context cancelled, stopping sniffer")
			return
		case pkt, ok := <-packetSource.Packets():
			if !ok {
				log.Println("[LLDP] packet source closed")
				return
			}
			s.handlePacket(pkt)
		}
	}
}

// handlePacket يحلل Packet واحد ويحدّث LLDPDevice map.
func (s *LLDPSource) handlePacket(pkt gopacket.Packet) {
	ethLayer := pkt.Layer(layers.LayerTypeEthernet)
	if ethLayer == nil {
		return
	}
	eth, _ := ethLayer.(*layers.Ethernet)
	if eth == nil {
		return
	}

	srcMAC := eth.SrcMAC.String()

	payload := pkt.Data()
	if len(payload) <= 14 {
		return
	}

	// تخطي Ethernet header (14 bytes) للوصول إلى LLDP TLVs
	llBytes := payload[14:]

	sysName, sysDesc, portID, mgmtAddr := parseLLDPTLVs(llBytes)
	name := sysName
	if name == "" {
		name = sysDesc
	}

	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	dev := s.devices[srcMAC]
	dev.MAC = srcMAC
	dev.IP = mgmtAddr
	dev.Name = name
	dev.PortID = portID
	dev.ChassisID = "" // يمكن توسعتها لاحقاً إذا احتجنا TLV chassis-specific
	dev.Interface = s.ifaceName
	dev.LastSeen = now

	s.devices[srcMAC] = dev
}

// parseLLDPTLVs مقتبس من منطق sniffer/lldp.go (مبسّط)
// يمر على TLVs ويستخرج System Name / Description / Port ID / Management Address.
func parseLLDPTLVs(b []byte) (sysName, sysDesc, portID, mgmtAddr string) {
	for len(b) >= 2 {
		tlvHeader := uint16(b[0])<<8 | uint16(b[1])
		tlvType := (tlvHeader & 0xFE00) >> 9
		tlvLen := tlvHeader & 0x01FF

		if tlvType == 0 {
			break
		}
		if len(b) < int(2+tlvLen) {
			break
		}

		value := b[2 : 2+tlvLen]

		switch tlvType {
		case 5: // System Name
			sysName = safeASCII(value)
		case 6: // System Description
			sysDesc = safeASCII(value)
		case 2: // Port ID
			if len(value) > 1 {
				// أول بايت = subtype
				portID = safeASCII(value[1:])
			}
		case 8: // Management Address
			// عادةً: length, addr subtype, address, interface subtype, interface number, OID len, OID
			if len(value) >= 2 {
				addrLen := int(value[0])
				if len(value) >= 1+addrLen && addrLen >= net.IPv4len {
					ip := net.IP(value[1 : 1+addrLen])
					mgmtAddr = ip.String()
				}
			}
		}

		b = b[2+tlvLen:]
	}

	return
}

// safeASCII يحوّل bytes إلى string مع إهمال الأحرف غير القابلة للطباعة.
func safeASCII(b []byte) string {
	out := make([]byte, 0, len(b))
	for _, r := range b {
		if r >= 32 && r <= 126 {
			out = append(out, r)
		}
	}
	return string(out)
}
