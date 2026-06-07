package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/instaagrammeta/somon-crm/backend/internal/middleware"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
)

type MessageHandler struct{ *App }

func NewMessageHandler(a *App) *MessageHandler { return &MessageHandler{a} }

// GET /api/messages?limit=50
func (h *MessageHandler) List(c *gin.Context) {
	limit := 100
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 && l <= 500 {
		limit = l
	}
	var rows []models.Message
	h.DB.Order("created_at DESC").Limit(limit).Find(&rows)
	// reverse to oldest-first for chat UI
	for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
		rows[i], rows[j] = rows[j], rows[i]
	}
	c.JSON(http.StatusOK, rows)
}

// POST /api/messages (multipart: text + optional file)
func (h *MessageHandler) Send(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	uname := middleware.CurrentUserName(c)
	msg := models.Message{
		UserID:   &uid,
		UserName: uname,
		Message:  c.PostForm("message"),
	}
	if file, err := c.FormFile("file"); err == nil {
		rel, ftype, err := utils.SaveUpload(file, h.Cfg.Upload.Dir, "chat")
		if err == nil {
			msg.FilePath = rel
			msg.FileName = file.Filename
			msg.FileType = ftype
		}
	}
	if msg.Message == "" && msg.FilePath == "" {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	if err := h.DB.Create(&msg).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, msg)
}

// PUT /api/messages/:id
func (h *MessageHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var m models.Message
	if err := h.DB.First(&m, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "error.not_found")
		return
	}
	uid := middleware.CurrentUserID(c)
	if !middleware.IsAdmin(c) && m.UserID != nil && *m.UserID != uid {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}
	var in struct {
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	if err := h.DB.Model(&m).Update("message", in.Message).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// DELETE /api/messages/:id
func (h *MessageHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var m models.Message
	if err := h.DB.First(&m, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "error.not_found")
		return
	}
	uid := middleware.CurrentUserID(c)
	if !middleware.IsAdmin(c) && m.UserID != nil && *m.UserID != uid {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}
	if err := h.DB.Delete(&m).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}
