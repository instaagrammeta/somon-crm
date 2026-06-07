package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/instaagrammeta/somon-crm/backend/internal/i18n"
)

// Locale resolves the request locale from query, header, or default.
func Locale(defaultLocale string) gin.HandlerFunc {
	return func(c *gin.Context) {
		loc := c.Query("lang")
		if loc == "" {
			loc = c.GetHeader("Accept-Language")
		}
		loc = i18n.NormalizeLocale(loc)
		if loc == "" {
			loc = defaultLocale
		}
		c.Set("locale", loc)
		c.Next()
	}
}
