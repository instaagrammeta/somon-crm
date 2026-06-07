package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/instaagrammeta/somon-crm/backend/internal/service"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
)

const (
	CtxUserID   = "user_id"
	CtxUserName = "user_name"
	CtxRole     = "user_role"
	CtxClaims   = "jwt_claims"
)

// JWTAuth checks the bearer token, ensures it is not blacklisted, then loads the user.
func JWTAuth(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			utils.ErrorResp(c, http.StatusUnauthorized, "error.unauthorized")
			return
		}
		claims, err := authSvc.ParseToken(token)
		if err != nil {
			utils.ErrorResp(c, http.StatusUnauthorized, "auth.token_invalid")
			return
		}

		// Check blacklist (logout)
		ctx, cancel := context.WithTimeout(c.Request.Context(), authSvc.RedisTimeout())
		defer cancel()
		if blocked, _ := authSvc.IsBlacklisted(ctx, claims.ID); blocked {
			utils.ErrorResp(c, http.StatusUnauthorized, "auth.token_invalid")
			return
		}

		c.Set(CtxClaims, claims)
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxUserName, claims.FullName)
		c.Set(CtxRole, claims.Role)
		c.Next()
	}
}

// AdminRequired must follow JWTAuth.
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(CtxRole)
		if role != models.RoleAdmin {
			utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
			return
		}
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	if t, err := c.Cookie("access_token"); err == nil {
		return t
	}
	if t := c.Query("token"); t != "" {
		return t
	}
	return ""
}

// CurrentUserID returns the authenticated user id from gin.Context.
func CurrentUserID(c *gin.Context) uint {
	v, _ := c.Get(CtxUserID)
	if id, ok := v.(uint); ok {
		return id
	}
	return 0
}

func CurrentUserName(c *gin.Context) string {
	v, _ := c.Get(CtxUserName)
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func CurrentRole(c *gin.Context) string {
	v, _ := c.Get(CtxRole)
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func IsAdmin(c *gin.Context) bool {
	return CurrentRole(c) == models.RoleAdmin
}
