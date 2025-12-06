package api

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"
    "strings"
    "time"
)

// respondJSON يرسل استجابة JSON موحّدة مع كود HTTP مناسب.
func respondJSON(w http.ResponseWriter, status int, v interface{
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data == nil {
		return nil
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(data)
}
}) {
    if w == nil {
        return
    }
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(status)
    if v == nil {
        return
    }
    if err := json.NewEncoder(w).Encode(v); err != nil {
        log.Printf("api: failed to encode JSON response: %v", err)
    }
}

// errorJSON غلاف بسيط لأخطاء JSON.
func errorJSON(w http.ResponseWriter, status int, msg string) {
    type errorEnvelope struct {
        Error string `json:"error"`
    }
    respondJSON(w, status, errorEnvelope{Error: msg})
}

// methodNotAllowed يرسل 405 مع هيدر Allow.
func methodNotAllowed(w http.ResponseWriter, r *http.Request, allowed ...string) {
    if w == nil {
        return
    }
    if len(allowed) > 0 {
        w.Header().Set("Allow", strings.Join(allowed, ", "))
    }
    errorJSON(w, http.StatusMethodNotAllowed, "method not allowed")
}

// parseLimit يقرأ معامل limit من query parameters مع حدود افتراضية.
func parseLimit(r *http.Request, key string, def, max int) int {
    if r == nil {
        return def
    }
    raw := r.URL.Query().Get(key)
    if raw == "" {
        return def
    }
    n, err := strconv.Atoi(raw)
    if err != nil || n < 0 {
        return def
    }
    if max > 0 && n > max {
        return max
    }
    return n
}

// parseTimeQuery يحاول قراءة وقت بصيغة RFC3339 من query parameter.
func parseTimeQuery(r *http.Request, key string) *time.Time {
    if r == nil {
        return nil
    }
    raw := r.URL.Query().Get(key)
    if raw == "" {
        return nil
    }
    t, err := time.Parse(time.RFC3339, raw)
    if err != nil {
        return nil
    }
    return &t
}

// pathParam يستخرج الجزء المتبقي من الـ URL بعد prefix معين.
func pathParam(r *http.Request, prefix string) string {
    if r == nil || r.URL == nil {
        return ""
    }
    path := r.URL.Path
    if !strings.HasPrefix(path, prefix) {
        return ""
    }
    id := strings.TrimPrefix(path, prefix)
    id = strings.Trim(id, "/")
    return id
}

// applyMiddleware يضيف CORS + logging + recovery على كل الطلبات.
func applyMiddleware(next http.Handler) http.Handler {
    if next == nil {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            errorJSON(w, http.StatusInternalServerError, "router not configured")
        })
    }

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if rec := recover(); rec != nil {
                log.Printf("api: panic in HTTP handler: %v", rec)
                errorJSON(w, http.StatusInternalServerError, "internal server error")
            }
        }()

        // CORS بسيط للاستخدام من الواجهة الأمامية أو أدوات خارجية.
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }

        next.ServeHTTP(w, r)
    })
}
