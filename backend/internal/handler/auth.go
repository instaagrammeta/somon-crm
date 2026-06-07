package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/instaagrammeta/somon-crm/backend/internal/middleware"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/instaagrammeta/somon-crm/backend/internal/service"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
)

type AuthHandler struct{ *App }

func NewAuthHandler(a *App) *AuthHandler { return &AuthHandler{a} }

type loginInput struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// POST /api/login
func (h *AuthHandler) Login(c *gin.Context) {
	var in loginInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	var u models.User
	if err := h.DB.Where("login = ? AND is_active = TRUE", in.Login).First(&u).Error; err != nil {
		utils.ErrorResp(c, http.StatusUnauthorized, "auth.invalid_credentials")
		return
	}
	if !h.Auth.CheckPassword(u.Password, in.Password) {
		utils.ErrorResp(c, http.StatusUnauthorized, "auth.invalid_credentials")
		return
	}
	token, exp, err := h.Auth.IssueToken(&u)
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "error.internal")
		return
	}
	now := time.Now().UTC()
	h.DB.Model(&u).Update("last_login", now)

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"token":      token,
		"expires_at": exp,
		"user":       u.ToPublic(),
	})
}

// POST /api/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	v, _ := c.Get(middleware.CtxClaims)
	if claims, ok := v.(*service.Claims); ok && claims != nil {
		_ = h.Auth.Blacklist(c.Request.Context(), claims)
	}
	utils.OKMessage(c, gin.H{"message": "logged_out"})
}

// GET /api/check-auth
func (h *AuthHandler) CheckAuth(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	if uid == 0 {
		c.JSON(http.StatusOK, gin.H{"authenticated": false})
		return
	}
	var u models.User
	if err := h.DB.First(&u, uid).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"authenticated": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"user":          u.ToPublic(),
	})
}

// POST /api/me/telegram-link  -> returns code + deep link
func (h *AuthHandler) TelegramGenerateLink(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	if uid == 0 {
		utils.ErrorResp(c, http.StatusUnauthorized, "error.unauthorized")
		return
	}
	if !h.Telegram.Enabled() {
		utils.ErrorRaw(c, http.StatusServiceUnavailable, "telegram disabled")
		return
	}
	code, err := h.Telegram.GenerateLinkCode(c.Request.Context(), uid)
	if err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	deep := "https://t.me/" + h.Telegram.BotUsername() + "?start=" + code
	c.JSON(http.StatusOK, gin.H{
		"code":          code,
		"deep_link":     deep,
		"bot_username":  h.Telegram.BotUsername(),
		"valid_minutes": 15,
	})
}

// DELETE /api/me/telegram-link  -> unlink current user's telegram
func (h *AuthHandler) TelegramUnlink(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	if err := h.DB.Model(&models.User{}).Where("id = ?", uid).Updates(map[string]any{
		"telegram_chat_id":   0,
		"telegram_username":  "",
		"telegram_linked_at": nil,
	}).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}
