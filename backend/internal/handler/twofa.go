package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/instaagrammeta/somon-crm/backend/internal/middleware"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
)

// TwoFAHandler exposes /api/2fa/* endpoints. All routes here are behind the
// regular auth middleware: the user must already be logged-in (with
// password) to view their 2FA status, set it up, or disable it.
type TwoFAHandler struct{ *App }

func NewTwoFAHandler(a *App) *TwoFAHandler { return &TwoFAHandler{a} }

// GET /api/2fa/status
func (h *TwoFAHandler) Status(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	c.JSON(http.StatusOK, h.TwoFA.Status(uid))
}

// POST /api/2fa/setup — generate a new secret and return the otpauth:// URL
// that the frontend renders as a QR code. Idempotent while Enabled=false.
func (h *TwoFAHandler) Setup(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	login := middleware.CurrentUserName(c)
	res, err := h.TwoFA.Setup(uid, login)
	if err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, res)
}

// POST /api/2fa/verify — confirm enrollment using the first 6-digit code.
// On success returns 8 plaintext recovery codes (shown ONCE).
func (h *TwoFAHandler) Verify(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	var body struct {
		Code string `json:"code"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.Code == "" {
		body.Code = strings.TrimSpace(c.PostForm("code"))
	}
	codes, err := h.TwoFA.Verify(uid, body.Code)
	if err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "twofa."+err.Error())
		return
	}
	utils.OK(c, gin.H{"recovery_codes": codes})
}

// POST /api/2fa/disable — disable 2FA. Requires the current password to be
// re-entered as a guard against session hijacking.
func (h *TwoFAHandler) Disable(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	var body struct {
		Password string `json:"password"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.Password == "" {
		body.Password = c.PostForm("password")
	}
	var u models.User
	if err := h.DB.First(&u, uid).Error; err != nil {
		utils.ErrorResp(c, http.StatusUnauthorized, "auth.unauthorized")
		return
	}
	if !h.Auth.CheckPassword(u.Password, body.Password) {
		utils.ErrorResp(c, http.StatusUnauthorized, "auth.wrong_password")
		return
	}
	if err := h.TwoFA.Disable(uid); err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, nil)
}
