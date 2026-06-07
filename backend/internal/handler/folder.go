package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/instaagrammeta/somon-crm/backend/internal/middleware"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
)

type FolderHandler struct{ *App }

func NewFolderHandler(a *App) *FolderHandler { return &FolderHandler{a} }

// GET /api/folders?parent_id=
func (h *FolderHandler) List(c *gin.Context) {
	parent, _ := strconv.Atoi(c.DefaultQuery("parent_id", "0"))
	var rows []models.Folder
	h.DB.Where("parent_id = ?", parent).Order("name ASC").Find(&rows)
	c.JSON(http.StatusOK, rows)
}

// POST /api/folders
func (h *FolderHandler) Create(c *gin.Context) {
	var in models.Folder
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	uid := middleware.CurrentUserID(c)
	in.AuthorID = &uid
	in.AuthorName = middleware.CurrentUserName(c)
	if err := h.DB.Create(&in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "id": in.ID})
}

// PUT /api/folders/:id (rename)
func (h *FolderHandler) Rename(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var in struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	if err := h.DB.Model(&models.Folder{}).Where("id = ?", id).Update("name", in.Name).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// DELETE /api/folders/:id
func (h *FolderHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.Folder{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// GET /api/folders/:id/files
func (h *FolderHandler) ListFiles(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var rows []models.FolderFile
	h.DB.Where("folder_id = ?", id).Order("created_at DESC").Find(&rows)
	c.JSON(http.StatusOK, rows)
}

// POST /api/folders/:id/files (multipart "file")
func (h *FolderHandler) UploadFile(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	rel, ftype, err := utils.SaveUpload(file, h.Cfg.Upload.Dir, "folders")
	if err != nil {
		utils.ErrorRaw(c, http.StatusBadRequest, err.Error())
		return
	}
	uid := middleware.CurrentUserID(c)
	row := models.FolderFile{
		FolderID:     uint(id),
		Filename:     filepath.Base(rel),
		OriginalName: file.Filename,
		FilePath:     rel,
		FileType:     ftype,
		FileSize:     file.Size,
		AuthorID:     &uid,
		AuthorName:   middleware.CurrentUserName(c),
	}
	if err := h.DB.Create(&row).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, row)
}

// DELETE /api/files/:id
func (h *FolderHandler) DeleteFile(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var f models.FolderFile
	if err := h.DB.First(&f, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "file.not_found")
		return
	}
	if err := h.DB.Delete(&f).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	// best effort remove from disk
	if f.FilePath != "" {
		full := filepath.Join(h.Cfg.Upload.Dir, filepath.FromSlash(stripUploadsPrefix(f.FilePath)))
		_ = os.Remove(full)
	}
	utils.OK(c, gin.H{"success": true})
}

// GET /api/download/:id
func (h *FolderHandler) Download(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var f models.FolderFile
	if err := h.DB.First(&f, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "file.not_found")
		return
	}
	full := filepath.Join(h.Cfg.Upload.Dir, filepath.FromSlash(stripUploadsPrefix(f.FilePath)))
	c.FileAttachment(full, f.OriginalName)
}

func stripUploadsPrefix(p string) string {
	const pref = "/uploads/"
	if len(p) >= len(pref) && p[:len(pref)] == pref {
		return p[len(pref):]
	}
	return p
}
