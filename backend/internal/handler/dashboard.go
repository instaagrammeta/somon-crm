package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/instaagrammeta/somon-crm/backend/internal/middleware"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
)

type DashboardHandler struct{ *App }

func NewDashboardHandler(a *App) *DashboardHandler { return &DashboardHandler{a} }

// GET /api/dashboard/stats
func (h *DashboardHandler) Stats(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	role := middleware.CurrentRole(c)

	var totalUsers, totalRequests, openRequests, totalLeads, totalTasks, myTasks, totalSims, totalApartments, freeApartments, soldApartments int64
	h.DB.Model(&models.User{}).Where("is_active = TRUE").Count(&totalUsers)
	h.DB.Model(&models.RequestItem{}).Count(&totalRequests)
	h.DB.Model(&models.RequestItem{}).Joins("JOIN requests_columns c ON c.id = requests_items.column_id").
		Where("c.title <> ?", "Анҷом").Count(&openRequests)
	h.DB.Model(&models.KanbanLead{}).Count(&totalLeads)
	h.DB.Model(&models.Task{}).Count(&totalTasks)
	if role == models.RoleAdmin {
		myTasks = totalTasks
	} else {
		h.DB.Model(&models.Task{}).Where("executor_id = ?", uid).Count(&myTasks)
	}
	h.DB.Model(&models.SimCard{}).Where("status = 'active'").Count(&totalSims)
	h.DB.Model(&models.ObjektApartment{}).Count(&totalApartments)
	h.DB.Model(&models.ObjektApartment{}).Where("status = 'free'").Count(&freeApartments)
	h.DB.Model(&models.ObjektApartment{}).Where("status = 'sold'").Count(&soldApartments)

	// Last 7-day chart
	type bucket struct {
		Day   time.Time `json:"day"`
		Count int       `json:"count"`
	}
	var leads7 []bucket
	h.DB.Raw(`SELECT date_trunc('day', created_at) AS day, COUNT(*) AS count
              FROM kanban_leads
              WHERE created_at >= NOW() - INTERVAL '7 days'
              GROUP BY 1 ORDER BY 1`).Scan(&leads7)
	var requests7 []bucket
	h.DB.Raw(`SELECT date_trunc('day', created_at) AS day, COUNT(*) AS count
              FROM requests_items
              WHERE created_at >= NOW() - INTERVAL '7 days'
              GROUP BY 1 ORDER BY 1`).Scan(&requests7)

	c.JSON(http.StatusOK, gin.H{
		"users":           totalUsers,
		"requests":        totalRequests,
		"requests_open":   openRequests,
		"leads":           totalLeads,
		"tasks":           totalTasks,
		"my_tasks":        myTasks,
		"sim_cards":       totalSims,
		"apartments":      totalApartments,
		"apartments_free": freeApartments,
		"apartments_sold": soldApartments,
		"leads_7d":        leads7,
		"requests_7d":     requests7,
	})
}
