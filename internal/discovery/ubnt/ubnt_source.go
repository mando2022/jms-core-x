// File: jms-core-x/internal/discovery/ubnt/ubnt_source.go
// Block 8 – UBNT Discovery Source (Data Collection Only)
//
// المسئوليات:
//   - إرسال broadcast probes على UDP 10001.
//   - استقبال ردود UBNT Discovery.
//   - تحليل الـ packets واستخراج UBNTDevice.
//   - إدارة TTL للأجهزة + Snapshot بدون أي Normalize.

package discovery

import (
	"context"
	"encoding/binary"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

// UBNTDevice يمثل جهاز تم اكتشافه عن طريق بروتوكول UBNT Discovery.
type UBNTDevice struct {
	MAC         string            // MAC address
	IPv4        string            // عنوان IPv4 الأساسي (إن وجد)
	Hostname    string            // Hostname للجهاز
	Model       string            // Model / Platform
	Firmware    string            // Firmware version
	Interface   string            // اسم الواجهة المحلي (Display Name)
	LastSeen    time.Time         // آخر مرة شوهد فيها هذا الجهاز
	ExtraFields map[string]string // حقول إضافية (ESSID, WirelessMode, ... إلخ)
}

// UBNTSource يقوم بـ:
//   - إرسال broadcast probe على UDP 10001.
//   - استقبال ردود UBNT Discovery.
//   - تحليل الـ packets واستخراج UBNTDevice.
//   - الحفاظ على TTL map + Snapshot.
type UBNTSource struct {
	ifaceName string

	mu      sync.Mutex
	devices map[string]UBNTDevice // keyed by MAC أو ip:xxx عند غياب MAC
	seen    map[string]time.Time  // keyed by نفس المفتاح المستخدم في devices
	ttl     time.Duration

	cancel  context.CancelFunc
	running bool
}

// UBNT constants (نفس الـ port المستخدم في sniffer/ubnt.go و ubnt-tools)
const (
	ubntPort       = 10001
	ubntScanPeriod = 20 * time.Second
	ubntTTL        = 2 * time.Minute
)

// Probe v1 + v2 (منطق مستوحى من sniffer/ubnt.go)
var (
	ubntProbeV1 = []byte{0x01, 0x00, 0x00, 0x00}
	ubntProbeV2 = []byte{0x02, 0x08, 0x00, 0x00}
)

// NewUBNTSource ينشئ مصدر UBNT جديد.
func NewUBNTSource(ifaceName string) *UBNTSource {
	return &UBNTSource{
		ifaceName: ifaceName,
		devices:   make(map[string]UBNTDevice),
		seen:      make(map[string]time.Time),
		ttl:       ubntTTL,
	}
}

// Start يشغّل Goroutine لإرسال الـ probes واستقبال ردود UBNT Discovery.
func (u *UBNTSource) Start(ctx context.Context) error {
	u.mu.Lock()
	if u.running {
		u.mu.Unlock()
		return nil
	}
	u.running = true
	u.mu.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	u.cancel = cancel

	go u.run(ctx)

	return nil
}

// Stop يوقِف الـ UBNTSource ويوقف كل الـ goroutines.
func (u *UBNTSource) Stop() {
	u.mu.Lock()
	if !u.running {
		u.mu.Unlock()
		return
	}
	u.running = false
	cancel := u.cancel
	u.cancel = nil
	u.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}

// Snapshot يعيد الأجهزة التي ما زالت ضمن TTL.
// بدون دمج أو Normalize – هذا مسئولية البلوكات اللاحقة.
func (u *UBNTSource) Snapshot() []UBNTDevice {
	now := time.Now()

	u.mu.Lock()
	defer u.mu.Unlock()

	out := make([]UBNTDevice, 0, len(u.devices))
	for key, dev := range u.devices {
		ts, ok := u.seen[key]
		if !ok {
			delete(u.devices, key)
			continue
		}
		if now.Sub(ts) > u.ttl {
			delete(u.devices, key)
			delete(u.seen, key)
			continue
		}
		out = append(out, dev)
	}

	return out
}

// run يقوم بفتح Socket UDP، إرسال probes، واستقبال الردود وتحليلها.
func (u *UBNTSource) run(ctx context.Context) {
	laddr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: ubntPort,
	}

	conn, err := net.ListenUDP("udp4", laddr)
	if err != nil {
		log.Println("[UBNT] ListenUDP error:", err)
		return
	}
	defer conn.Close()

	log.Println("[UBNT] Discovery socket started:", conn.LocalAddr())

	// عنوان broadcast الافتراضي
	bcast := &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: ubntPort,
	}

	// Goroutine للإرسال الدوري للـ probes
	go func() {
		ticker := time.NewTicker(ubntScanPeriod)
		defer ticker.Stop()

		// send first probes immediately
		if err := u.sendProbe(conn, bcast); err != nil {
			log.Println("[UBNT] initial probe error:", err)
		}

		for {
			select {
			case <-ctx.Done():
				log.Println("[UBNT] context cancelled, stopping probe sender")
				return
			case <-ticker.C:
				if err := u.sendProbe(conn, bcast); err != nil {
					log.Println("[UBNT] probe error:", err)
				}
			}
		}
	}()

	buf := make([]byte, 64*1024)

	for {
		_ = conn.SetReadDeadline(time.Now().Add(ubntScanPeriod))

		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				// timeout طبيعي، نُكمِل طالما الـ context لم يُنهَ.
				select {
				case <-ctx.Done():
					log.Println("[UBNT] context cancelled, stopping receiver")
					return
				default:
					continue
				}
			}

			log.Println("[UBNT] ReadFromUDP error:", err)
			select {
			case <-ctx.Done():
				log.Println("[UBNT] context cancelled after read error")
				return
			default:
				continue
			}
		}

		if n <= 0 {
			select {
			case <-ctx.Done():
				return
			default:
				continue
			}
		}

		data := make([]byte, n)
		copy(data, buf[:n])
		u.handlePacket(addr, data)
	}
}

// sendProbe يرسل probe v1 + v2 إلى broadcast address.
func (u *UBNTSource) sendProbe(conn *net.UDPConn, bcast *net.UDPAddr) error {
	if _, err := conn.WriteToUDP(ubntProbeV1, bcast); err != nil {
		log.Println("[UBNT] probe v1 error:", err)
		return err
	}
	if _, err := conn.WriteToUDP(ubntProbeV2, bcast); err != nil {
		log.Println("[UBNT] probe v2 error:", err)
		return err
	}
	return nil
}

// handlePacket يحلل Packet واحد من UBNT Discovery ويحدّث الـ map.
func (u *UBNTSource) handlePacket(addr *net.UDPAddr, data []byte) {
	dev := UBNTDevice{
		IPv4:        addr.IP.String(),
		Interface:   u.ifaceName,
		ExtraFields: map[string]string{},
	}

	macStr, hostname, firmware, modelStr := parseUBNTTLVs(data)

	now := time.Now()
	dev.MAC = macStr
	dev.Hostname = hostname
	dev.Firmware = firmware
	dev.Model = modelStr
	dev.LastSeen = now

	u.mu.Lock()
	defer u.mu.Unlock()

	key := dev.MAC
	if key == "" {
		// لو الـ MAC مش معروف، نستخدم الـ IP كمفتاح مؤقت
		key = "ip:" + dev.IPv4
	}

	u.devices[key] = dev
	u.seen[key] = now
}

// parseUBNTTLVs مقتبس ومبسّط مستوحى من sniffer/ubnt.go + ubnt-tools/discovery/packet.go.
// يمر على TLVs ويستخرج MAC + Firmware + Hostname + Model.
// الميثود الأصلي أكثر تعقيداً؛ هنا نأخذ الحد الأدنى المفيد للـ UI.
func parseUBNTTLVs(b []byte) (macStr, hostname, firmware, modelStr string) {
	var macStr, hostname, firmware, modelStr string

	i := 0
	for i+2 <= len(b) {
		typeID := b[i]
		length := int(b[i+1])
		i += 2
		if i+length > len(b) {
			break
		}
		value := b[i : i+length]
		i += length

		switch typeID {
		case 1: // MAC address (example)
			if len(value) == 6 {
				macStr = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
					value[0], value[1], value[2], value[3], value[4], value[5])
			}
		case 2: // hostname (example)
			hostname = string(value)
		case 3: // firmware (example)
			firmware = string(value)
		case 4: // model (example)
			modelStr = string(value)
		default:
			// ignore unknown TLV
		}
	}

	return macStr, hostname, firmware, modelStr
}
}

// cleanUBNTString ينضّف string من الفراغات والـ NULLs.
func cleanUBNTString(b []byte) string {
	s := string(b)
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\x00")
	return s
}
