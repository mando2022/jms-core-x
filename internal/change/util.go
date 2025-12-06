// File: jms-core-x/internal/change/util.go
// دوال مساعدة بسيطة تستخدم داخل Block 14 فقط.
package change

import (
	"encoding/json"
    "reflect"
    "sort"
    "strings"
)

// equalStringSlice يقارن بين مصفوفتين من النصوص بعد
// مراعاة الطول والترتيب والقيم. يعتبر nil و []string{} متساويين.
func equalStringSlice(a, b []string) bool {
    if len(a) != len(b) {
        return false
    }
    for i := range a {
        if a[i] != b[i] {
            return false
        }
    }
    return true
}

// normalizeLists يقوم بتنظيف قائمة النصوص عبر:
//   - Trim لكل عنصر
//   - حذف العناصر الفارغة
//   - إزالة التكرار
//   - ترتيب النتيجة للحصول على مقارنة مستقرة.
func normalizeLists(values []string) []string {
    if len(values) == 0 {
        return nil
    }

    m := make(map[string]struct{}, len(values))
    out := make([]string, 0, len(values))

    for _, v := range values {
        v = strings.TrimSpace(v)
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

    sort.Strings(out)
    return out
}

// mapifyLists يحوّل slice نصي إلى خريطة لاستخدامها في المقارنات السريعة.
func mapifyLists(values []string) map[string]struct{
	m := make(map[string]struct{}, len(values))
	for _, v := range values {
		if v == "" {
			continue
		}
		m[v] = struct{}{}
	}
	return m
}
} {
    values = normalizeLists(values)
    m := make(map[string]struct{}, len(values))
    for _, v := range values {
        m[v] = struct{}{}
    }
    return m
}

// buildChangeEvent يبني ChangeEvent موحدًا مع ضم
	return ChangeEvent{
		DeviceID: deviceID,
		Field:    field,
		Old:      safeInterface(oldVal),
		New:      safeInterface(newVal),
		Type:     changeType,
	}
}
ان
// أن القيم في OldValue/NewValue صالحة للتسلسل لاحقًا.
func buildChangeEvent(eventType, deviceID, field string, oldVal, newVal interface{}) ChangeEvent {
    return ChangeEvent{
        Type:     eventType,
        DeviceID: deviceID,
        Field:    field,
        OldValue: safeInterface(oldVal),
        NewValue: safeInterface(newVal),
    }
}

// safeInterface تستخدم كطبقة حماية بسيطة لتفادي
// تمرير قيم غير قابلة للت
	switch val := v.(type) {
	case nil:
		return nil
	case string:
		return val
	case []byte:
		return string(val)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", val)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%g", val)
	case fmt.Stringer:
		return val.String()
	default:
		b, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%T", val)
		}
		return string(b)
	}
}
سلسل (مثل الخرائط ذات المفاتيح المعقدة).
// حالياً تطبّق منطقًا بسيطًا فقط، ويمكن توسيعها لاحقًا عند الحاجة.
func safeInterface(v interface{}) interface{} {
    if v == nil {
        return nil
    }

    // slices الفارغة تعتبر nil لعدم إزعاج المستهلكين.
    rv := reflect.ValueOf(v)
    if rv.Kind() == reflect.Slice && rv.Len() == 0 {
        return nil
    }

    return v
}
