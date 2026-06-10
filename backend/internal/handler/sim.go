package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/instaagrammeta/somon-crm/backend/internal/i18n"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
)

type SimHandler struct{ *App }

func NewSimHandler(a *App) *SimHandler { return &SimHandler{a} }

// ===== Company phones =====

// GET /api/company-phones
func (h *SimHandler) ListPhones(c *gin.Context) {
	type row struct {
		models.CompanyPhone
		AssignedName string `json:"assigned_name"`
	}
	var rows []row
	h.DB.Table("company_phones p").
		Select("p.*, u.full_name AS assigned_name").
		Joins("LEFT JOIN users u ON p.assigned_to = u.id").
		Order("p.created_at DESC").Scan(&rows)
	c.JSON(http.StatusOK, rows)
}

// POST /api/company-phones
func (h *SimHandler) AddPhone(c *gin.Context) {
	var in models.CompanyPhone
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	if err := h.DB.Create(&in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "id": in.ID})
}

// PUT /api/company-phones/:id
func (h *SimHandler) UpdatePhone(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var p models.CompanyPhone
	if err := h.DB.First(&p, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "phone.not_found")
		return
	}
	var in models.CompanyPhone
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	in.ID = p.ID
	if err := h.DB.Model(&p).Updates(in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// DELETE /api/company-phones/:id
func (h *SimHandler) DeletePhone(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.CompanyPhone{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// ===== SIM cards =====

// GET /api/sim-cards
func (h *SimHandler) ListSims(c *gin.Context) {
	type row struct {
		models.SimCard
		AssignedName string `json:"assigned_name"`
		PhoneModel   string `json:"phone_model"`
	}
	var rows []row
	h.DB.Table("sim_cards s").
		Select("s.*, u.full_name AS assigned_name, p.model AS phone_model").
		Joins("LEFT JOIN users u ON s.assigned_to = u.id").
		Joins("LEFT JOIN company_phones p ON s.phone_id = p.id").
		Order("s.created_at DESC").Scan(&rows)
	c.JSON(http.StatusOK, rows)
}

// POST /api/sim-cards
func (h *SimHandler) AddSim(c *gin.Context) {
	var in models.SimCard
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	if err := h.DB.Create(&in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "id": in.ID})
}

// PUT /api/sim-cards/:id
func (h *SimHandler) UpdateSim(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var s models.SimCard
	if err := h.DB.First(&s, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "sim.not_found")
		return
	}
	var in models.SimCard
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	in.ID = s.ID
	if err := h.DB.Model(&s).Updates(in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// DELETE /api/sim-cards/:id
func (h *SimHandler) DeleteSim(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.SimCard{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// ===== Tariffs =====

// GET /api/sim-tariffs?sim_id=
func (h *SimHandler) ListTariffs(c *gin.Context) {
	var rows []models.SimTariff
	q := h.DB.Order("created_at DESC")
	if sid := c.Query("sim_id"); sid != "" {
		q = q.Where("sim_id = ?", sid)
	}
	q.Find(&rows)
	c.JSON(http.StatusOK, rows)
}

// POST /api/sim-tariffs
func (h *SimHandler) AddTariff(c *gin.Context) {
	var in models.SimTariff
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	if err := h.DB.Create(&in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	// Auto-create matching payment row
	if in.Cost > 0 {
		pay := models.TariffPayment{
			SimID: in.SimID, TariffID: &in.ID, Amount: in.Cost,
			StartDate: in.StartDate, EndDate: in.EndDate,
			Status: "paid",
		}
		now := time.Now().UTC()
		pay.PaymentDate = &now
		_ = h.DB.Create(&pay).Error
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "id": in.ID})
}

// PUT /api/sim-tariffs/:id
func (h *SimHandler) UpdateTariff(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var t models.SimTariff
	if err := h.DB.First(&t, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "tariff.not_found")
		return
	}
	var in models.SimTariff
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	in.ID = t.ID
	if err := h.DB.Model(&t).Updates(in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// DELETE /api/sim-tariffs/:id
func (h *SimHandler) DeleteTariff(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.SimTariff{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// GET /api/tariff-payments?sim_id=
func (h *SimHandler) ListPayments(c *gin.Context) {
	var rows []models.TariffPayment
	q := h.DB.Order("payment_date DESC")
	if sid := c.Query("sim_id"); sid != "" {
		q = q.Where("sim_id = ?", sid)
	}
	q.Find(&rows)
	c.JSON(http.StatusOK, rows)
}

// DELETE /api/tariff-payments/:id
func (h *SimHandler) DeletePayment(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.TariffPayment{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// GET /api/sim-phones/stats
func (h *SimHandler) Stats(c *gin.Context) {
	var totalSims, activeSims, freePhones int64
	var totalPaid float64
	h.DB.Model(&models.SimCard{}).Count(&totalSims)
	h.DB.Model(&models.SimCard{}).Where("status = 'active'").Count(&activeSims)
	h.DB.Model(&models.CompanyPhone{}).Where("status = 'free'").Count(&freePhones)
	h.DB.Model(&models.TariffPayment{}).Select("COALESCE(SUM(amount), 0)").Row().Scan(&totalPaid)
	c.JSON(http.StatusOK, gin.H{
		"total_sims":  totalSims,
		"active_sims": activeSims,
		"free_phones": freePhones,
		"total_paid":  totalPaid,
	})
}

// GET /api/sim-phones/export/excel
func (h *SimHandler) ExportExcel(c *gin.Context) {
	type row struct {
		Phone        string     `json:"phone_number"`
		Operator     string     `json:"operator"`
		Status       string     `json:"status"`
		AssignedName string     `json:"assigned_name"`
		PhoneModel   string     `json:"phone_model"`
		LastCost     float64    `json:"last_cost"`
		EndDate      *time.Time `json:"end_date"`
	}
	var rows []row
	h.DB.Table("sim_cards s").
		Select(`s.phone_number, s.operator, s.status,
		        u.full_name AS assigned_name,
		        p.model AS phone_model,
		        (SELECT cost FROM sim_tariffs WHERE sim_id = s.id ORDER BY id DESC LIMIT 1) AS last_cost,
		        (SELECT end_date FROM sim_tariffs WHERE sim_id = s.id ORDER BY id DESC LIMIT 1) AS end_date`).
		Joins("LEFT JOIN users u ON s.assigned_to = u.id").
		Joins("LEFT JOIN company_phones p ON s.phone_id = p.id").
		Order("s.created_at DESC").Scan(&rows)

	headers := []string{"Телефон", "Оператор", "Ҳолат", "Корбар", "Модели телефон", "Тариф", "Охирин рӯзи тариф"}
	body := make([][]any, len(rows))
	for i, r := range rows {
		end := ""
		if r.EndDate != nil {
			end = r.EndDate.Format("2006-01-02")
		}
		body[i] = []any{r.Phone, r.Operator, r.Status, r.AssignedName, r.PhoneModel, r.LastCost, end}
	}
	bytes, err := utils.BuildExcel("SIM-cards", headers, body)
	if err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.Header("Content-Disposition", "attachment; filename=sim_cards.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", bytes)
}

// SendExpiryNotifications scans for tariffs ending within N days and sends Telegram alerts.
// Triggered by /admin/sim-tariffs/notify-expiring (admin) or a cron worker (future).
func (h *SimHandler) SendExpiryNotifications(c *gin.Context) {
	days := 3
	if d, err := strconv.Atoi(c.DefaultQuery("days", "3")); err == nil {
		days = d
	}
	cutoff := time.Now().UTC().AddDate(0, 0, days)
	type row struct {
		SimID      uint   `gorm:"column:sim_id"`
		Phone      string `gorm:"column:phone_number"`
		AssignedTo *uint  `gorm:"column:assigned_to"`
		EndDate    time.Time
	}
	var rows []row
	h.DB.Table("sim_tariffs t").
		Select("t.sim_id, s.phone_number, s.assigned_to, t.end_date").
		Joins("JOIN sim_cards s ON s.id = t.sim_id").
		Where("t.status = 'active' AND t.end_date IS NOT NULL AND t.end_date <= ?", cutoff).
		Scan(&rows)

	sent := 0
	for _, r := range rows {
		if r.AssignedTo == nil {
			continue
		}
		daysLeft := int(time.Until(r.EndDate).Hours() / 24)
		if daysLeft < 0 {
			daysLeft = 0
		}
		go h.Notif.Push(*r.AssignedTo, models.NotifyTariffExpiring,
			i18n.Translate(i18n.LocaleTG, "notify.tariff_expiring"),
			r.Phone, "/sim-cards",
			"tg.tariff_expiring", r.Phone, daysLeft)
		sent++
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "sent": sent, "scanned": len(rows)})
}
