package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/instaagrammeta/somon-crm/backend/internal/middleware"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
	"gorm.io/gorm"
)

type UserHandler struct{ *App }

func NewUserHandler(a *App) *UserHandler { return &UserHandler{a} }

// GET /api/users  (admin)
func (h *UserHandler) List(c *gin.Context) {
	var users []models.User
	q := h.DB.Order("created_at DESC")
	if cat := c.Query("category"); cat != "" {
		q = q.Where("category = ?", cat)
	}
	if role := c.Query("role"); role != "" {
		q = q.Where("role = ?", role)
	}
	if err := q.Find(&users).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]models.PublicUser, len(users))
	for i, u := range users {
		out[i] = u.ToPublic()
	}
	c.JSON(http.StatusOK, out)
}

// GET /api/users/list  (any authenticated)
func (h *UserHandler) ListBasic(c *gin.Context) {
	type basic struct {
		ID       uint   `json:"id"`
		FullName string `json:"full_name"`
		Role     string `json:"role"`
		Category string `json:"category"`
	}
	var rows []basic
	if err := h.DB.Model(&models.User{}).
		Select("id, full_name, role, category").
		Where("is_active = TRUE").
		Order("full_name ASC").
		Find(&rows).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, rows)
}

type userInput struct {
	FullName         string             `form:"full_name" json:"full_name"`
	Age              int                `form:"age" json:"age"`
	PersonalPhones   models.StringSlice `form:"personal_phones" json:"personal_phones"`
	WorkPhones       models.StringSlice `form:"work_phones" json:"work_phones"`
	Login            string             `form:"login" json:"login"`
	Password         string             `form:"password" json:"password"`
	Category         string             `form:"category" json:"category"`
	Role             string             `form:"role" json:"role"`
	TelegramUsername string             `form:"telegram_username" json:"telegram_username"`
	NotifyTelegram   *bool              `form:"notify_telegram" json:"notify_telegram"`
	IsActive         *bool              `form:"is_active" json:"is_active"`
}

// POST /api/users  (admin)
func (h *UserHandler) Create(c *gin.Context) {
	var in userInput
	if err := c.ShouldBind(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	if in.Login == "" || in.Password == "" || in.FullName == "" {
		utils.ErrorResp(c, http.StatusBadRequest, "error.validation", "login/password/full_name")
		return
	}

	// Optional photo upload (multipart)
	photoPath := ""
	if file, err := c.FormFile("photo"); err == nil {
		rel, _, err := utils.SaveUpload(file, h.Cfg.Upload.Dir, "users")
		if err == nil {
			photoPath = rel
		}
	}

	hashed, err := h.Auth.HashPassword(in.Password)
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "error.internal")
		return
	}
	if in.Role == "" {
		in.Role = models.RoleEmployee
	}
	user := models.User{
		FullName:         in.FullName,
		Age:              in.Age,
		PersonalPhones:   in.PersonalPhones,
		WorkPhones:       in.WorkPhones,
		Login:            in.Login,
		Password:         hashed,
		Photo:            photoPath,
		Category:         in.Category,
		Role:             in.Role,
		TelegramUsername: in.TelegramUsername,
		IsActive:         true,
		NotifyTelegram:   true,
	}
	if in.NotifyTelegram != nil {
		user.NotifyTelegram = *in.NotifyTelegram
	}
	if in.IsActive != nil {
		user.IsActive = *in.IsActive
	}

	if err := h.DB.Create(&user).Error; err != nil {
		if isUniqueViolation(err) {
			utils.ErrorResp(c, http.StatusConflict, "user.exists")
			return
		}
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Pre-issue a Telegram linking code; returned to the admin so they can hand it to the new user.
	resp := gin.H{"success": true, "user": user.ToPublic()}
	if h.Telegram.Enabled() {
		if code, err := h.Telegram.GenerateLinkCode(c.Request.Context(), user.ID); err == nil {
			resp["telegram_link_code"] = code
			resp["telegram_deep_link"] = "https://t.me/" + h.Telegram.BotUsername() + "?start=" + code
			resp["telegram_bot"] = h.Telegram.BotUsername()
		}
	}
	c.JSON(http.StatusCreated, resp)
}

// PUT /api/users/:id  (admin)
func (h *UserHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var u models.User
	if err := h.DB.First(&u, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "user.not_found")
		return
	}
	var in userInput
	if err := c.ShouldBind(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	updates := map[string]any{}
	if in.FullName != "" {
		updates["full_name"] = in.FullName
	}
	if in.Age != 0 {
		updates["age"] = in.Age
	}
	if in.PersonalPhones != nil {
		updates["personal_phones"] = in.PersonalPhones
	}
	if in.WorkPhones != nil {
		updates["work_phones"] = in.WorkPhones
	}
	if in.Login != "" {
		updates["login"] = in.Login
	}
	if in.Category != "" {
		updates["category"] = in.Category
	}
	if in.Role != "" {
		updates["role"] = in.Role
	}
	if in.TelegramUsername != "" {
		updates["telegram_username"] = in.TelegramUsername
	}
	if in.NotifyTelegram != nil {
		updates["notify_telegram"] = *in.NotifyTelegram
	}
	if in.IsActive != nil {
		updates["is_active"] = *in.IsActive
	}
	if in.Password != "" {
		hashed, err := h.Auth.HashPassword(in.Password)
		if err != nil {
			utils.ErrorResp(c, http.StatusInternalServerError, "error.internal")
			return
		}
		updates["password"] = hashed
	}
	if file, err := c.FormFile("photo"); err == nil {
		rel, _, err := utils.SaveUpload(file, h.Cfg.Upload.Dir, "users")
		if err == nil {
			updates["photo"] = rel
		}
	}
	if err := h.DB.Model(&u).Updates(updates).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	_ = h.DB.First(&u, id)
	c.JSON(http.StatusOK, gin.H{"success": true, "user": u.ToPublic()})
}

// DELETE /api/users/:id  (admin)
func (h *UserHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if uint(id) == middleware.CurrentUserID(c) {
		utils.ErrorRaw(c, http.StatusBadRequest, "cannot delete self")
		return
	}
	if err := h.DB.Delete(&models.User{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// pgx returns SQLSTATE 23505; gorm wraps the message — substring is OK
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := err.Error()
	return contains(msg, "duplicate") || contains(msg, "23505") || contains(msg, "UNIQUE")
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
