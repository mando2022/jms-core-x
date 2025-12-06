package identity

import (
	"crypto/sha1"
	"encoding/hex"
	"sort"
	"strings"
)

// GenerateDeviceID ينشئ DeviceID ثابت اعتمادًا على مجموعة الخصائص الأساسية.
// القاعدة كما في الـ Blueprint:
//
//   device-<hash_of_sources_and_mac>
//
// حيث يعتمد الهاش على:
//   - MAC لو موجود
//   - وإلا مزيج من IPs + Hostnames + Vendor + Sources
func GenerateDeviceID(mac string, ips, hostnames []string, vendor string, sources []string) string {
	mac = strings.ToLower(strings.TrimSpace(mac))
	vendor = strings.ToLower(strings.TrimSpace(vendor))

	ips = normalizeList(ips)
	hostnames = normalizeList(hostnames)
	sources = normalizeList(sources)

	var parts []string
	if mac != "" {
		parts = append(parts, "mac="+mac)
	} else {
		parts = append(parts,
			"ips="+strings.Join(ips, ","),
			"hosts="+strings.Join(hostnames, ","),
			"vendor="+vendor,
		)
	}

	if len(sources) > 0 {
		parts = append(parts, "sources="+strings.Join(sources, ","))
	}

	key := strings.Join(parts, "|")

	h := sha1.Sum([]byte(key))
	return "device-" + hex.EncodeToString(h[:])
}

// mergeStringList يدمج عدة قوائم strings في قائمة واحدة بدون تكرار
// مع إزالة القيم الفارغة وترتيب الناتج بشكل ثابت.
func mergeStringList(base []string, lists ...[]string) []string {
	seen := make(map[string]struct{}, len(base))

	for _, v := range base {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		lv := strings.ToLower(v)
		seen[lv] = struct{}{}
	}

	for _, list := range lists {
		for _, v := range list {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			lv := strings.ToLower(v)
			seen[lv] = struct{}{}
		}
	}

	out := make([]string, 0, len(seen))
	for v := range seen {
		out = append(out, v)
	}
	sort.Strings(out)

	return out
}

// normalizeList يطبّق نفس منطق mergeStringList لكن على قائمة واحدة.
func normalizeList(in []string) []string {
	return mergeStringList(nil, in)
}
