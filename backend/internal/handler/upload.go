package handler

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
)

type UploadHandler struct{ *App }

func NewUploadHandler(a *App) *UploadHandler { return &UploadHandler{a} }

// GET /uploads/*path  — serve static files with simple traversal protection.
// In production prefer Nginx; this is a fallback for development.
func (h *UploadHandler) ServeFile(c *gin.Context) {
	rel := c.Param("path")
	rel = strings.TrimPrefix(rel, "/")
	if strings.Contains(rel, "..") {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	full := filepath.Join(h.Cfg.Upload.Dir, filepath.FromSlash(rel))
	c.File(full)
}
