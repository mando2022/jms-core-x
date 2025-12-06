// File: jms-core-x/internal/core/final/util.go
package final

import "sort"

// normalizeList ينظف قائمة من الـ strings عن طريق:
//   - إزالة العناصر الفارغة.
//   - إزالة التكرار.
//   - الترتيب حسب القيمة (lexicographical).
// بدون أي منطق خاص بالهوية أو الشبكة.
func normalizeList(in []string) []string {
    if len(in) == 0 {
        return nil
    }

    // إزالة الفارغ وبناء خريطة للتكرارات.
    m := make(map[string]struct{}, len(in))
    for _, v := range in {
        if v == "" {
            continue
        }
        m[v] = struct{}{}
    }

    if len(m) == 0 {
        return nil
    }

    out := make([]string, 0, len(m))
    for v := range m {
        out = append(out, v)
    }

    // ترتيب القيم.
    sort.Strings(out)
    return out
}

// uniqueStrings يبني قائمة فريدة من strings مع الحفاظ على الترتيب الأولي
// قدر الإمكان. تُستخدم في الأماكن التي نحتاج فيها إزالة التكرار فقط.
func uniqueStrings(in []string) []string {
    if len(in) == 0 {
        return nil
    }

    m := make(map[string]struct{}, len(in))
    out := make([]string, 0, len(in))

    for _, v := range in {
        if v == "" {
            continue
        }
        if _, exists := m[v]; exists {
            continue
        }
        m[v] = struct{}{}
        out = append(out, v)
    }

    if len(out) == 0 {
        return nil
    }

    return out
}
