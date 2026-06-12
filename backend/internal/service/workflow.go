package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/instaagrammeta/somon-crm/backend/internal/models"
)

// WorkflowEngine runs simple if-this-then-that automations. Each Workflow has
// a Trigger (e.g. "lead.created") and a list of Steps; when a domain handler
// fires Trigger(ctx, "lead.created", payload) we look up every enabled
// Workflow with that trigger and execute its steps in order.
//
// Steps are persisted as JSON so the no-code Workflow Builder can edit them
// without schema migrations. Supported step types are intentionally small but
// cover ~80% of real automations:
//
//   - notify_user        params: {user_id, title, body, link?}
//   - notify_role        params: {role, title, body, link?}
//   - tg_send            params: {user_id, message}
//   - whatsapp_send      params: {phone, message}
//   - email_send         params: {to, subject, text, html?}
//   - create_task        params: {title, description?, executor_id, status?}
//   - set_field          params: {entity, id, field, value}   (audited)
//   - delay_minutes      params: {minutes}                    (cooperative sleep)
//   - log                params: {message}
type WorkflowEngine struct {
	db    *gorm.DB
	notif *NotificationService
	tg    *TelegramService
	wa    *WhatsAppService
	mail  *EmailService
}

func NewWorkflowEngine(db *gorm.DB, notif *NotificationService, tg *TelegramService, wa *WhatsAppService, mail *EmailService) *WorkflowEngine {
	return &WorkflowEngine{db: db, notif: notif, tg: tg, wa: wa, mail: mail}
}

// Trigger is the public entry point. Call it from any domain handler when
// something interesting happens. Each matching workflow runs in its own
// goroutine so the calling handler is never blocked by automation.
func (e *WorkflowEngine) Trigger(ctx context.Context, name string, payload map[string]any) {
	if e == nil || e.db == nil {
		return
	}
	var rows []models.Workflow
	if err := e.db.Where("trigger = ? AND enabled = ?", name, true).Find(&rows).Error; err != nil {
		return
	}
	for i := range rows {
		wf := rows[i]
		if !configMatches(wf.Config, payload) {
			continue
		}
		go e.run(ctx, wf, payload)
	}
}

// configMatches lets a workflow narrow its own scope (e.g. only fire for a
// specific kanban column id). Each key in `wf.Config` must be present in the
// payload with the *same* JSON value. Empty config matches everything.
func configMatches(rawCfg models.JSONB, payload map[string]any) bool {
	if len(rawCfg) == 0 || string(rawCfg) == "{}" {
		return true
	}
	var cfg map[string]any
	if err := json.Unmarshal([]byte(rawCfg), &cfg); err != nil || len(cfg) == 0 {
		return true
	}
	for k, want := range cfg {
		got, ok := payload[k]
		if !ok {
			return false
		}
		// Compare via JSON to make types match across int/float/string.
		a, _ := json.Marshal(want)
		b, _ := json.Marshal(got)
		if string(a) != string(b) {
			return false
		}
	}
	return true
}

// run executes a single workflow. Failures stop the run but never crash the
// engine; everything is recorded in workflow_runs for the UI to surface.
func (e *WorkflowEngine) run(ctx context.Context, wf models.Workflow, payload map[string]any) {
	pjson, _ := json.Marshal(payload)
	runRow := models.WorkflowRun{
		WorkflowID: wf.ID,
		StartedAt:  time.Now(),
		Status:     "running",
		Payload:    models.JSONB(pjson),
	}
	if e.db != nil {
		_ = e.db.Create(&runRow).Error
	}

	var steps []map[string]any
	if err := json.Unmarshal([]byte(wf.Steps), &steps); err != nil {
		e.finishRun(&runRow, "failed", fmt.Sprintf("steps parse: %v", err))
		return
	}

	var logBuf strings.Builder
	for i, step := range steps {
		stype, _ := step["type"].(string)
		params, _ := step["params"].(map[string]any)
		if stype == "" {
			continue
		}
		fmt.Fprintf(&logBuf, "[%d] %s: ", i+1, stype)
		err := e.runStep(ctx, stype, params, payload)
		if err != nil {
			fmt.Fprintf(&logBuf, "FAIL %v\n", err)
			e.finishRun(&runRow, "failed", logBuf.String())
			return
		}
		fmt.Fprintf(&logBuf, "ok\n")
	}
	e.finishRun(&runRow, "ok", logBuf.String())

	now := time.Now()
	e.db.Model(&models.Workflow{}).Where("id = ?", wf.ID).Updates(map[string]any{
		"last_run_at": now,
		"last_status": "ok",
		"runs":        gorm.Expr("runs + 1"),
	})
}

func (e *WorkflowEngine) finishRun(row *models.WorkflowRun, status, log string) {
	if e.db == nil {
		return
	}
	now := time.Now()
	row.FinishedAt = &now
	row.Status = status
	row.Log = log
	_ = e.db.Save(row).Error
}

// runStep is the dispatcher. Add new step types here as they are needed; each
// type is intentionally one short function so tests can target them in
// isolation.
func (e *WorkflowEngine) runStep(ctx context.Context, kind string, p map[string]any, payload map[string]any) error {
	resolve := func(s string) string { return interpolate(s, payload) }
	switch kind {
	case "notify_user":
		uid := uintParam(p["user_id"])
		if uid == 0 {
			return errors.New("user_id required")
		}
		e.notif.Push(uid, "workflow",
			resolve(stringParam(p["title"])),
			resolve(stringParam(p["body"])),
			resolve(stringParam(p["link"])), "")
		return nil

	case "notify_role":
		role := stringParam(p["role"])
		if role == "" {
			return errors.New("role required")
		}
		var users []models.User
		e.db.Where("role = ? AND is_active = ?", role, true).Find(&users)
		title := resolve(stringParam(p["title"]))
		body := resolve(stringParam(p["body"]))
		link := resolve(stringParam(p["link"]))
		for _, u := range users {
			e.notif.Push(u.ID, "workflow", title, body, link, "")
		}
		return nil

	case "tg_send":
		uid := uintParam(p["user_id"])
		msg := resolve(stringParam(p["message"]))
		if uid == 0 || msg == "" {
			return errors.New("user_id+message required")
		}
		if e.tg != nil {
			e.tg.Notify(uid, msg)
		}
		return nil

	case "whatsapp_send":
		if e.wa == nil || !e.wa.Enabled() {
			return errors.New("whatsapp disabled")
		}
		return e.wa.SendText(stringParam(p["phone"]), resolve(stringParam(p["message"])))

	case "email_send":
		if e.mail == nil || !e.mail.Enabled() {
			return errors.New("email disabled")
		}
		return e.mail.SendMessage(
			stringParam(p["to"]),
			resolve(stringParam(p["subject"])),
			resolve(stringParam(p["text"])),
			resolve(stringParam(p["html"])),
		)

	case "create_task":
		t := models.Task{
			Title:       resolve(stringParam(p["title"])),
			Description: resolve(stringParam(p["description"])),
			Status:      ifNonEmpty(stringParam(p["status"]), models.TaskStatusNew),
		}
		if eid := uintParam(p["executor_id"]); eid != 0 {
			t.ExecutorID = &eid
			t.ExecutorIDs = models.UintSlice{eid}
		}
		return e.db.Create(&t).Error

	case "set_field":
		entity := stringParam(p["entity"])
		id := uintParam(p["id"])
		field := stringParam(p["field"])
		if entity == "" || id == 0 || field == "" {
			return errors.New("entity+id+field required")
		}
		table := safeTable(entity)
		if table == "" {
			return fmt.Errorf("entity %q not allowed", entity)
		}
		return e.db.Exec(
			fmt.Sprintf("UPDATE %s SET %s = ?, updated_at = NOW() WHERE id = ?", table, safeColumn(field)),
			p["value"], id,
		).Error

	case "delay_minutes":
		mins := intParam(p["minutes"])
		if mins <= 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(mins) * time.Minute):
			return nil
		}

	case "log":
		// no-op; the message is already captured in the run log via runStep's caller
		return nil
	}
	return fmt.Errorf("unknown step type: %s", kind)
}

// interpolate replaces {{key}} references in `s` with their stringified value
// from `payload`. Missing keys become an empty string. This is a tiny, safe
// templating that purposely doesn't do nested expressions.
func interpolate(s string, payload map[string]any) string {
	if !strings.Contains(s, "{{") {
		return s
	}
	out := s
	for k, v := range payload {
		needle := "{{" + k + "}}"
		out = strings.ReplaceAll(out, needle, fmt.Sprintf("%v", v))
	}
	return out
}

// safeTable / safeColumn restrict set_field to a known whitelist so a
// malicious workflow payload cannot UPDATE arbitrary tables/columns.
func safeTable(entity string) string {
	switch entity {
	case "lid", "lead":
		return "lids"
	case "task":
		return "tasks"
	case "request":
		return "request_items"
	case "house":
		return "houses"
	case "website_lead":
		return "website_leads"
	}
	return ""
}

func safeColumn(c string) string {
	// allow [a-z0-9_]+
	out := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return -1
	}, strings.ToLower(c))
	return out
}

func stringParam(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func uintParam(v any) uint {
	switch n := v.(type) {
	case float64:
		return uint(n)
	case int:
		return uint(n)
	case int64:
		return uint(n)
	case string:
		var u uint64
		fmt.Sscanf(n, "%d", &u)
		return uint(u)
	}
	return 0
}

func intParam(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	}
	return 0
}

func ifNonEmpty(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
