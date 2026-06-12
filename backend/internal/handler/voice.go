package handler

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/instaagrammeta/somon-crm/backend/internal/middleware"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/instaagrammeta/somon-crm/backend/internal/service"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
)

// VoiceHandler accepts audio recordings produced by the in-app voice recorder
// (a small <button> that records via MediaRecorder and POSTs WebM/Opus). The
// audio is stored on disk; transcription runs asynchronously through the AI
// service so the upload response returns immediately.
type VoiceHandler struct{ *App }

func NewVoiceHandler(a *App) *VoiceHandler { return &VoiceHandler{a} }

// POST /api/voice/upload?entity=task&entity_id=12
//
//	multipart/form-data with file=audio/webm
//
// On success the row is created in voice_notes with status=pending, then a
// goroutine kicks off Whisper transcription and patches transcript+status.
func (h *VoiceHandler) Upload(c *gin.Context) {
	entity := c.Query("entity")
	if entity == "" {
		entity = c.PostForm("entity")
	}
	if entity == "" {
		entity = "task"
	}
	idStr := c.Query("entity_id")
	if idStr == "" {
		idStr = c.PostForm("entity_id")
	}
	entityID, _ := strconv.ParseUint(idStr, 10, 64)
	if entityID == 0 {
		utils.ErrorResp(c, http.StatusBadRequest, "voice.entity_id_required")
		return
	}

	fh, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	rel, _, err := utils.SaveUpload(fh, h.Cfg.Upload.Dir, "voice")
	if err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}

	uid := middleware.CurrentUserID(c)
	uidPtr := &uid
	row := models.VoiceNote{
		UserID:   uidPtr,
		Entity:   entity,
		EntityID: uint(entityID),
		AudioURL: rel,
		Status:   "pending",
	}
	if dur, _ := strconv.Atoi(c.PostForm("duration")); dur > 0 {
		row.DurationSec = dur
	}
	if err := h.DB.Create(&row).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Fire-and-forget transcription. We re-open the file on disk because the
	// multipart uploaded handle is closed once the request ends.
	if h.AI != nil && h.AI.Enabled() {
		fileBytes, ferr := readUploadedFile(fh)
		if ferr == nil {
			go h.transcribeAsync(row.ID, fileBytes, fh.Filename, c.GetString("locale"))
		}
	}

	c.JSON(http.StatusCreated, row)
}

func (h *VoiceHandler) transcribeAsync(id uint, audio []byte, filename, locale string) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	text, err := h.AI.Transcribe(ctx, audio, filename, locale, nil)
	updates := map[string]any{}
	if err != nil {
		updates["status"] = "failed"
	} else {
		updates["status"] = "done"
		updates["transcript"] = text
		updates["transcript_lang"] = locale
	}
	h.DB.Model(&models.VoiceNote{}).Where("id = ?", id).Updates(updates)

	// Push a realtime event so the recorder UI can swap "transcribing…" for
	// the actual transcript without polling.
	if h.Hub != nil {
		var row models.VoiceNote
		h.DB.First(&row, id)
		userID := uint(0)
		if row.UserID != nil {
			userID = *row.UserID
		}
		if userID != 0 {
			h.Hub.SendToUser(userID, service.Event{
				Topic:   "voice",
				Action:  "update",
				Payload: row,
			})
		}
	}
}

// GET /api/voice?entity=task&entity_id=12 — list voice notes attached to a row.
func (h *VoiceHandler) List(c *gin.Context) {
	entity := c.Query("entity")
	id, _ := strconv.ParseUint(c.Query("entity_id"), 10, 64)
	if entity == "" || id == 0 {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	var rows []models.VoiceNote
	if err := h.DB.Where("entity = ? AND entity_id = ?", entity, id).
		Order("created_at DESC").Find(&rows).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, rows)
}

// DELETE /api/voice/:id — author or admin.
func (h *VoiceHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var row models.VoiceNote
	if err := h.DB.First(&row, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "error.not_found")
		return
	}
	uid := middleware.CurrentUserID(c)
	if !middleware.IsAdmin(c) && (row.UserID == nil || *row.UserID != uid) {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}
	if err := h.DB.Delete(&row).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, nil)
}

// readUploadedFile opens the multipart handle and reads it fully into RAM.
// Audio notes are short (typically <1 MB) so this is safe.
func readUploadedFile(fh *multipart.FileHeader) ([]byte, error) {
	rc, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}
