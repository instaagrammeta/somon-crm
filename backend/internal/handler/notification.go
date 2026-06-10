package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/instaagrammeta/somon-crm/backend/internal/middleware"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
)

type NotificationHandler struct{ *App }

func NewNotificationHandler(a *App) *NotificationHandler { return &NotificationHandler{a} }

// GET /api/notifications?unread=1&limit=100
func (h *NotificationHandler) List(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	q := h.DB.Model(&models.Notification{}).
		Where("user_id = ?", uid).
		Order("created_at DESC")
	if c.Query("unread") == "1" {
		q = q.Where("is_read = ?", false)
	}
	limit := 100
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 && l <= 500 {
		limit = l
	}
	var rows []models.Notification
	if err := q.Limit(limit).Find(&rows).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []models.Notification{}
	}
	c.JSON(http.StatusOK, rows)
}

// GET /api/notifications/unread-count
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	var count int64
	h.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", uid, false).
		Count(&count)
	c.JSON(http.StatusOK, gin.H{"count": count})
}

// POST /api/notifications/:id/read
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", id, uid).
		Update("is_read", true).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// POST /api/notifications/read-all
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	if err := h.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", uid, false).
		Update("is_read", true).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// DELETE /api/notifications/:id
func (h *NotificationHandler) Delete(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.
		Where("id = ? AND user_id = ?", id, uid).
		Delete(&models.Notification{}).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, nil)
}
