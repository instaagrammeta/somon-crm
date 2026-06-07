package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/instaagrammeta/somon-crm/backend/internal/i18n"
)

// LocaleFromCtx returns the language code chosen by middleware (or default).
func LocaleFromCtx(c *gin.Context) string {
	if v, ok := c.Get("locale"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return i18n.LocaleTG
}

// ErrorResp writes a standard error response in the user's locale.
func ErrorResp(c *gin.Context, status int, key string, args ...any) {
	loc := LocaleFromCtx(c)
	c.AbortWithStatusJSON(status, gin.H{
		"success": false,
		"error":   i18n.Translate(loc, key, args...),
		"code":    key,
	})
}

// ErrorRaw — when the message is already a final string (e.g., DB error).
func ErrorRaw(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, gin.H{
		"success": false,
		"error":   msg,
	})
}

func OK(c *gin.Context, data any) {
	if data == nil {
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}
	c.JSON(http.StatusOK, data)
}

func OKMessage(c *gin.Context, data gin.H) {
	if data == nil {
		data = gin.H{}
	}
	data["success"] = true
	c.JSON(http.StatusOK, data)
}

// Created — 201 with object.
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, data)
}
