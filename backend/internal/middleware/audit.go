package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/instaagrammeta/somon-crm/backend/internal/models"
)

// Audit returns a gin middleware that records every mutating request (POST /
// PUT / PATCH / DELETE) under /api/* into the audit_logs table. Read-only
// requests are skipped to keep the table small.
//
// Sensitive fields (`password`, `token`, `secret`) are scrubbed from the
// stored payload — we only ever keep a redacted JSON snapshot.
func Audit(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		write := method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE"
		if !write || !strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Next()
			return
		}

		// Capture and re-feed the request body so downstream handlers can read it.
		var bodySnapshot []byte
		if c.Request.Body != nil && c.Request.ContentLength > 0 && c.Request.ContentLength < 64*1024 {
			ct := c.Request.Header.Get("Content-Type")
			if strings.HasPrefix(ct, "application/json") || strings.HasPrefix(ct, "application/x-www-form-urlencoded") {
				bodySnapshot, _ = io.ReadAll(c.Request.Body)
				c.Request.Body = io.NopCloser(bytes.NewReader(bodySnapshot))
			}
		}

		c.Next()

		// Skip auth-noise that floods the table; otherwise record everything.
		path := c.Request.URL.Path
		if path == "/api/login" || path == "/api/logout" || strings.HasPrefix(path, "/api/notifications") {
			return
		}

		row := models.AuditLog{
			Method:    method,
			Path:      truncStr(path, 255),
			Status:    c.Writer.Status(),
			IP:        c.ClientIP(),
			UserAgent: truncStr(c.Request.UserAgent(), 255),
			Action:    actionFromMethod(method),
		}
		if uid := CurrentUserID(c); uid != 0 {
			id := uid
			row.UserID = &id
			row.UserLogin = CurrentUserName(c)
		}
		row.Entity, row.EntityID = entityFromPath(path)

		if len(bodySnapshot) > 0 {
			redacted := redactJSON(bodySnapshot)
			row.Payload = models.JSONB(redacted)
		}

		// Best-effort: never block the request on the audit write.
		go func(r models.AuditLog) { _ = db.Create(&r).Error }(row)
	}
}

// entityFromPath splits "/api/leads/42/foo" into ("leads", "42").
func entityFromPath(path string) (string, string) {
	parts := strings.Split(strings.TrimPrefix(path, "/api/"), "/")
	if len(parts) == 0 {
		return "", ""
	}
	entity := parts[0]
	id := ""
	if len(parts) > 1 && parts[1] != "" && parts[1][0] >= '0' && parts[1][0] <= '9' {
		id = parts[1]
	}
	return entity, id
}

func actionFromMethod(m string) string {
	switch m {
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	}
	return strings.ToLower(m)
}

// redactJSON parses a JSON object and replaces any sensitive value with "***".
// If the body is not valid JSON we return an empty {} so we never store raw
// password fields by accident.
func redactJSON(raw []byte) []byte {
	var v map[string]any
	if err := json.Unmarshal(raw, &v); err != nil {
		return []byte("{}")
	}
	for k := range v {
		lk := strings.ToLower(k)
		if strings.Contains(lk, "password") || strings.Contains(lk, "secret") || strings.Contains(lk, "token") {
			v[k] = "***"
		}
	}
	out, _ := json.Marshal(v)
	if len(out) > 4096 {
		return []byte(`{"truncated":true}`)
	}
	return out
}

func truncStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
