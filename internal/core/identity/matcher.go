package identity

import (
	"jms-core-x/internal/core/normalization"
)

// BuildIdentityDevices يطبّق منطق المطابقة على قائمة NormalizedDevice
// بحيث يتم دمج كل الأجهزة التي تشير لنفس الكيان في IdentityDevice واحد.
// المبدأ الأساسي:
//   Each NormalizedDevice -> belongs to exactly one IdentityDevice.
func BuildIdentityDevices(devs []normalization.NormalizedDevice) []IdentityDevice {
	if len(devs) == 0 {
		return nil
	}

	assigned := make(map[string]bool, len(devs))

	// خرائط تجميع حسب MAC و IP و Hostname
	macGroups := make(map[string][]normalization.NormalizedDevice)
	ipGroups := make(map[string][]normalization.NormalizedDevice)
	hostGroups := make(map[string][]normalization.NormalizedDevice)

	for _, d := range devs {
		if canUseMAC(d.MAC) {
			macGroups[d.MAC] = append(macGroups[d.MAC], d)
		}
		if d.IP != "" {
			ipGroups[d.IP] = append(ipGroups[d.IP], d)
		}
		if hk := hostGroupKey(d.Hostname); hk != "" {
			hostGroups[hk] = append(hostGroups[hk], d)
		}
	}

	out := make([]IdentityDevice, 0, len(devs))

	// 1) تجميع حسب MAC (أعلى أولوية).
	for mac, group := range macGroups {
		// استبعاد الأجهزة المعينة مسبقًا
		var bucket []normalization.NormalizedDevice
		for _, d := range group {
			if d.ID != "" && assigned[d.ID] {
				continue
			}
			bucket = append(bucket, d)
		}
		if len(bucket) == 0 {
			continue
		}

		idDev := mergeDevicesIntoIdentity(bucket)
		// MAC أساسي في هذه المجموعة
		idDev.MAC = mac
		idDev.DeviceID = GenerateDeviceID(idDev.MAC, idDev.IPs, idDev.Hostnames, idDev.Vendor, idDev.Sources)

		out = append(out, idDev)

		for _, d := range bucket {
			if d.ID != "" {
				assigned[d.ID] = true
			}
		}
	}

	// 2) تجميع حسب IP للأجهزة غير المعينة.
	for ip, group := range ipGroups {
		_ = ip // المفتاح هنا للمعلومة فقط

		var bucket []normalization.NormalizedDevice
		for _, d := range group {
			if d.ID != "" && assigned[d.ID] {
				continue
			}
			bucket = append(bucket, d)
		}
		if len(bucket) == 0 {
			continue
		}

		idDev := mergeDevicesIntoIdentity(bucket)
		// MAC قد يكون فارغًا هنا؛ DeviceID سيعتمد على IP/Hostname/Vendor.
		idDev.DeviceID = GenerateDeviceID(idDev.MAC, idDev.IPs, idDev.Hostnames, idDev.Vendor, idDev.Sources)

		out = append(out, idDev)

		for _, d := range bucket {
			if d.ID != "" {
				assigned[d.ID] = true
			}
		}
	}

	// 3) تجميع مساعد حسب Hostname للأجهزة المتبقية.
	for hk, group := range hostGroups {
		_ = hk // المفتاح مستخدم فقط للتجميع المنطقي

		var bucket []normalization.NormalizedDevice
		for _, d := range group {
			if d.ID != "" && assigned[d.ID] {
				continue
			}
			bucket = append(bucket, d)
		}
		if len(bucket) == 0 {
			continue
		}

		idDev := mergeDevicesIntoIdentity(bucket)
		idDev.DeviceID = GenerateDeviceID(idDev.MAC, idDev.IPs, idDev.Hostnames, idDev.Vendor, idDev.Sources)

		out = append(out, idDev)

		for _, d := range bucket {
			if d.ID != "" {
				assigned[d.ID] = true
			}
		}
	}

	// 4) أي أجهزة متبقية (بدون MAC/IP/Hostname مفيد) تعامل ككيانات منفصلة.
	for _, d := range devs {
		if d.ID != "" && assigned[d.ID] {
			continue
		}

		idDev := mergeDevicesIntoIdentity([]normalization.NormalizedDevice{d})
		idDev.DeviceID = GenerateDeviceID(idDev.MAC, idDev.IPs, idDev.Hostnames, idDev.Vendor, idDev.Sources)

		out = append(out, idDev)

		if d.ID != "" {
			assigned[d.ID] = true
		}
	}

	return out
}

// mergeDevicesIntoIdentity يدمج مجموعة من NormalizedDevice في IdentityDevice واحد
// مع دمج IPs/Hostnames/Ports/Sources واختيار Vendor مناسب.
func mergeDevicesIntoIdentity(devs []normalization.NormalizedDevice) IdentityDevice {
	var (
		ips       []string
		hostnames []string
		ports     []string
		sources   []string
		vendors   []string
		mac       string
	)

	for _, d := range devs {
		if mac == "" && d.MAC != "" {
			mac = d.MAC
		}
		if d.IP != "" {
			ips = mergeStringList(ips, []string{d.IP})
		}
		if d.Hostname != "" {
			hostnames = mergeStringList(hostnames, []string{d.Hostname})
		}
		if d.Port != "" {
			ports = mergeStringList(ports, []string{d.Port})
		}
		if d.UnifiedSource != "" {
			sources = mergeStringList(sources, []string{d.UnifiedSource})
		}
		if d.Vendor != "" {
			vendors = append(vendors, d.Vendor)
		}
	}

	vendor := chooseVendor(vendors)

	return IdentityDevice{
		DeviceID:  "", // يتم تعيينه لاحقًا عن طريق GenerateDeviceID
		MAC:       mac,
		IPs:       ips,
		Hostnames: hostnames,
		Vendor:    vendor,
		Ports:     ports,
		Sources:   sources,
	}
}
