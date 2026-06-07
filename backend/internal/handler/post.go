package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/instaagrammeta/somon-crm/backend/internal/middleware"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
)

type PostHandler struct{ *App }

func NewPostHandler(a *App) *PostHandler { return &PostHandler{a} }

// GET /api/posts
func (h *PostHandler) List(c *gin.Context) {
	var rows []models.Post
	q := h.DB.Order("created_at DESC")
	if cat := c.Query("category"); cat != "" {
		q = q.Where("category = ?", cat)
	}
	if proj := c.Query("project"); proj != "" {
		q = q.Where("project = ?", proj)
	}
	q.Find(&rows)
	c.JSON(http.StatusOK, rows)
}

// POST /api/posts (multipart)
func (h *PostHandler) Create(c *gin.Context) {
	var in models.Post
	if err := c.ShouldBind(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	uid := middleware.CurrentUserID(c)
	in.UserID = &uid
	in.UserName = middleware.CurrentUserName(c)

	if file, err := c.FormFile("media"); err == nil {
		rel, ftype, err := utils.SaveUpload(file, h.Cfg.Upload.Dir, "posts")
		if err == nil {
			in.MediaPath = rel
			in.MediaType = ftype
		}
	}
	if err := h.DB.Create(&in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "id": in.ID})
}

// PUT /api/posts/:id
func (h *PostHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var p models.Post
	if err := h.DB.First(&p, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "post.not_found")
		return
	}
	var in models.Post
	if err := c.ShouldBind(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	in.ID = p.ID
	if file, err := c.FormFile("media"); err == nil {
		rel, ftype, err := utils.SaveUpload(file, h.Cfg.Upload.Dir, "posts")
		if err == nil {
			in.MediaPath = rel
			in.MediaType = ftype
		}
	}
	if err := h.DB.Model(&p).Updates(in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// DELETE /api/posts/:id
func (h *PostHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.Post{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// GET /api/posts/export/excel
func (h *PostHandler) ExportExcel(c *gin.Context) {
	var rows []models.Post
	h.DB.Order("created_at DESC").Find(&rows)
	headers := []string{"ID", "Сарлавҳа", "Категория", "Проект", "Корбар", "Лайкҳо",
		"Шарҳҳо", "Шейр", "Намоиш", "Дастрасӣ", "Сана"}
	body := make([][]any, len(rows))
	for i, r := range rows {
		body[i] = []any{r.ID, r.Title, r.Category, r.Project, r.UserName,
			r.Likes, r.Comments, r.Shares, r.Views, r.Reach, r.CreatedAt}
	}
	b, err := utils.BuildExcel("Posts", headers, body)
	if err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.Header("Content-Disposition", "attachment; filename=posts.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", b)
}

// GET /api/posts/export/pdf — fallback to Excel (PDF Cyrillic font is a future task).
func (h *PostHandler) ExportPDF(c *gin.Context) {
	h.ExportExcel(c)
}
