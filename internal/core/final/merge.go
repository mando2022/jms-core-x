// File: jms-core-x/internal/core/final/merge.go
package final

import (
    "jms-core-x/internal/core/identity"
    "jms-core-x/internal/core/normalization"
)

// decorateFromNormalized يدمج metadata الخفيفة من NormalizedSnapshot
// داخل FinalDevice دون تنفيذ أي منطق Identity جديد.
//
// القواعد الأساسية:
//   - عدم تغيير DeviceID.
//   - عدم إنشاء أو حذف أجهزة.
//   - عدم إعادة توزيع NormalizedDevice على أجهزة مختلفة.
//   - مسموح فقط بتحسين Vendor / Hostnames / Ports / SeenSources.
func decorateFromNormalized(fd *FinalDevice, idDev identity.IdentityDevice, normIndex map[string][]normalization.NormalizedDevice) {
    if fd == nil {
        return
    }

    // اختيار المفتاح الأساسي للبحث في الفهرس (MAC أولاً ثم IPs).
    var candidates []normalization.NormalizedDevice

    if idDev.MAC != "" {
        if list, ok := normIndex[idDev.MAC]; ok {
            candidates = append(candidates, list...)
        }
    }

    if len(candidates) == 0 {
        // استخدام أول IP متاح كمفتاح بديل إن لزم.
        for _, ip := range idDev.IPs {
            if ip == "" {
                continue
            }
            if list, ok := normIndex[ip]; ok {
                candidates = append(candidates, list...)
            }
            if len(candidates) > 0 {
                break
            }
        }
    }

    if len(candidates) == 0 {
        return
    }

    // تحسين Vendor/Hostnames/Ports/SeenSources من قائمة الـ candidates.
    mergeVendor(fd, candidates)
    mergeHostnames(fd, candidates)
    mergePorts(fd, candidates)
    mergeSources(fd, candidates)
}

// mergeVendor يحسن قيمة Vendor إن وجد Vendor أفضل في normalized.
// لا يغيّر Vendor إن كانت القيمة الحالية غير فارغة وواضحة.
func mergeVendor(fd *FinalDevice, candidates []normalization.NormalizedDevice) {
    if fd == nil {
        return
    }

    if fd.Vendor != "" && fd.Vendor != "Unknown" {
        return
    }

    for _, nd := range candidates {
        if nd.Vendor != "" && nd.Vendor != "Unknown" {
            fd.Vendor = nd.Vendor
            return
        }
    }
}

// mergeHostnames يضيف Hostnames إضافية من normalized إلى FinalDevice.Hostnames
// ثم يمرر القيمة النهائية عبر normalizeList لإزالة التكرار.
func mergeHostnames(fd *FinalDevice, candidates []normalization.NormalizedDevice) {
    if fd == nil {
        return
    }

    collected := make([]string, 0, len(fd.Hostnames)+len(candidates))
    collected = append(collected, fd.Hostnames...)

    for _, nd := range candidates {
        if nd.Hostname != "" {
            collected = append(collected, nd.Hostname)
        }
    }

    fd.Hostnames = normalizeList(collected)
}

// mergePorts يدمج Port من normalized مع FinalDevice.Ports كمجرد metadata.
// لا يوجد أي منطق Network Mapping هنا؛ مجرد قائمة بسيطة بالـ Ports.
func mergePorts(fd *FinalDevice, candidates []normalization.NormalizedDevice) {
    if fd == nil {
        return
    }

    collected := make([]string, 0, len(fd.Ports)+len(candidates))
    collected = append(collected, fd.Ports...)

    for _, nd := range candidates {
        if nd.Port != "" {
            collected = append(collected, nd.Port)
        }
    }

    fd.Ports = normalizeList(collected)
}

// mergeSources يضيف UnifiedSource من normalized (لو متوفر) إلى SeenSources
// بهدف التتبع فقط، دون أي تأثير على منطق الهوية.
func mergeSources(fd *FinalDevice, candidates []normalization.NormalizedDevice) {
    if fd == nil {
        return
    }

    collected := make([]string, 0, len(fd.SeenSources)+len(candidates))
    collected = append(collected, fd.SeenSources...)

    for _, nd := range candidates {
        if nd.UnifiedSource != "" {
            collected = append(collected, nd.UnifiedSource)
        }
    }

    fd.SeenSources = normalizeList(collected)
}
