// Package router wires HTTP routes for the Somon CRM backend.
package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/instaagrammeta/somon-crm/backend/internal/config"
	"github.com/instaagrammeta/somon-crm/backend/internal/handler"
	"github.com/instaagrammeta/somon-crm/backend/internal/middleware"
)

// Build composes a gin.Engine with all routes registered.
func Build(app *handler.App, cfg *config.Config) *gin.Engine {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins))
	r.Use(middleware.Locale(cfg.Locale.Default))
	// v-2: every mutating /api/* request is recorded into audit_logs.
	r.Use(middleware.Audit(app.DB))

	// File upload size limit
	r.MaxMultipartMemory = cfg.Upload.MaxMB << 20

	// Static uploads (dev convenience)
	uploadH := handler.NewUploadHandler(app)
	r.GET("/uploads/*path", uploadH.ServeFile)

	// Health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true, "time": time.Now().UTC()})
	})

	auth := handler.NewAuthHandler(app)
	users := handler.NewUserHandler(app)
	tasks := handler.NewTaskHandler(app)
	requests := handler.NewRequestsHandler(app)
	lids := handler.NewLidHandler(app)
	houses := handler.NewHouseHandler(app)
	realty := handler.NewRealtyHandler(app)
	objekt := handler.NewObjektHandler(app)
	sims := handler.NewSimHandler(app)
	posts := handler.NewPostHandler(app)
	msgs := handler.NewMessageHandler(app)
	folders := handler.NewFolderHandler(app)
	ipoteka := handler.NewIpotekaHandler(app)
	dash := handler.NewDashboardHandler(app)
	notifs := handler.NewNotificationHandler(app)

	// v-2 handlers
	pub := handler.NewPublicHandler(app)
	twofa := handler.NewTwoFAHandler(app)
	wsH := handler.NewWSHandler(app)
	aiH := handler.NewAIHandler(app)
	voiceH := handler.NewVoiceHandler(app)
	tourH := handler.NewTourHandler(app)
	wlH := handler.NewWebsiteLeadHandler(app)
	wfH := handler.NewWorkflowHandler(app)
	waWh := handler.NewWhatsAppWebhookHandler(app)

	api := r.Group("/api")

	// === public website (no auth) ===
	pubGroup := api.Group("/public")
	pubGroup.GET("/site/config", pub.Config)
	pubGroup.GET("/houses", pub.Houses)
	pubGroup.GET("/houses/:id", pub.HouseDetail)
	pubGroup.GET("/tours/:slug", pub.Tour)
	pubGroup.POST("/leads", pub.CaptureLead)
	pubGroup.GET("/apk", pub.APK)

	// === WhatsApp webhook (no auth — Meta calls it directly) ===
	api.GET("/webhooks/whatsapp", waWh.Verify)
	api.POST("/webhooks/whatsapp", waWh.Receive)

	// === public ===
	api.POST("/login", auth.Login)
	api.GET("/banks", ipoteka.ListBanks)
	api.GET("/banks/:id", ipoteka.GetBank)
	api.GET("/banks/by-slug/:slug", ipoteka.GetBankBySlug)
	api.GET("/banks/:id/conditions", ipoteka.ListBankConditions)
	api.GET("/installment-objects", ipoteka.ListInstallments)
	api.GET("/installment-objects/:id", ipoteka.GetInstallment)
	api.GET("/installment-objects/by-slug/:slug", ipoteka.GetInstallmentBySlug)
	api.GET("/installment-objects/:id/conditions", ipoteka.ListInstallmentConditions)

	// === protected (JWT) ===
	priv := api.Group("")
	priv.Use(middleware.JWTAuth(app.Auth))

	priv.POST("/logout", auth.Logout)
	priv.GET("/check-auth", auth.CheckAuth)

	// telegram link (current user)
	priv.POST("/me/telegram-link", auth.TelegramGenerateLink)
	priv.DELETE("/me/telegram-link", auth.TelegramUnlink)

	// users (admin)
	adm := priv.Group("")
	adm.Use(middleware.AdminRequired())
	adm.GET("/users", users.List)
	adm.POST("/users", users.Create)
	adm.PUT("/users/:id", users.Update)
	adm.DELETE("/users/:id", users.Delete)
	priv.GET("/users/list", users.ListBasic) // any authenticated

	// tasks
	priv.GET("/tasks", tasks.List)
	priv.POST("/tasks", tasks.Create)
	priv.PUT("/tasks/:id", tasks.Update)
	priv.DELETE("/tasks/:id", tasks.Delete)

	// kanban-requests
	priv.GET("/requests-boards", requests.ListBoards)
	priv.POST("/requests-boards", requests.CreateBoard)
	priv.PUT("/requests-boards/:id", requests.UpdateBoard)
	priv.DELETE("/requests-boards/:id", requests.DeleteBoard)
	priv.GET("/requests-boards/:id/columns", requests.ListColumns)
	priv.POST("/requests-columns", requests.CreateColumn)
	priv.PUT("/requests-columns/:id", requests.UpdateColumn)
	priv.DELETE("/requests-columns/:id", requests.DeleteColumn)
	priv.PUT("/requests-columns/:id/order", requests.SwapColumnOrder)
	priv.GET("/requests-board/:id/requests", requests.ListItems)
	priv.POST("/requests-new", requests.CreateItem)
	priv.PUT("/requests-update/:id", requests.UpdateItem)
	priv.DELETE("/requests-delete/:id", requests.DeleteItem)
	priv.POST("/requests-move", requests.MoveItem)
	priv.POST("/upload-request-file", requests.UploadFile)
	priv.POST("/requests-bulk-paste", requests.BulkPaste)
	priv.GET("/requests-export/excel", requests.ExportBoardExcel)
	priv.GET("/requests-column-export/:id", requests.ExportColumnExcel)

	// lids (legacy)
	priv.GET("/lids", lids.List)
	priv.POST("/lids", lids.Create)
	priv.PUT("/lids/:id", lids.Update)
	priv.DELETE("/lids/:id", lids.Delete)
	priv.GET("/lids/export", lids.Export)

	// kanban (modern)
	priv.GET("/kanban/boards", lids.KanbanBoards)
	priv.POST("/kanban/boards", lids.KanbanCreateBoard)
	priv.PUT("/kanban/boards/:id", lids.KanbanUpdateBoard)
	priv.DELETE("/kanban/boards/:id", lids.KanbanDeleteBoard)
	priv.PUT("/kanban/boards/:id/archive", lids.KanbanArchiveBoard)
	priv.GET("/kanban/boards/:id/columns", lids.KanbanColumns)
	priv.GET("/kanban/boards/:id/export/excel", lids.KanbanExport)
	priv.POST("/kanban/columns", lids.KanbanCreateColumn)
	priv.PUT("/kanban/columns/:id", lids.KanbanUpdateColumn)
	priv.DELETE("/kanban/columns/:id", lids.KanbanDeleteColumn)
	priv.PUT("/kanban/columns/:id/order", lids.KanbanColumnOrder)
	priv.GET("/kanban/leads", lids.KanbanLeads)
	priv.POST("/kanban/leads", lids.KanbanCreateLead)
	priv.PUT("/kanban/leads/:id", lids.KanbanUpdateLead)
	priv.DELETE("/kanban/leads/:id", lids.KanbanDeleteLead)
	priv.POST("/kanban/leads/move", lids.KanbanMoveLead)
	priv.POST("/kanban/actions/track", lids.KanbanTrack)

	// houses
	priv.GET("/houses", houses.List)
	priv.POST("/houses", houses.Create)
	priv.PUT("/houses/:id", houses.Update)
	priv.DELETE("/houses/:id", houses.Delete)
	priv.GET("/houses/export", houses.Export)

	// realty
	priv.GET("/realty/objects", realty.ListObjects)
	priv.POST("/realty/objects", realty.CreateObject)
	priv.PUT("/realty/objects/:id", realty.UpdateObject)
	priv.DELETE("/realty/objects/:id", realty.DeleteObject)
	priv.GET("/realty/objects/:id/blocks", realty.ListBlocks)
	priv.POST("/realty/blocks", realty.CreateBlock)
	priv.PUT("/realty/blocks/:id", realty.UpdateBlock)
	priv.DELETE("/realty/blocks/:id", realty.DeleteBlock)
	priv.GET("/realty/pricing", realty.ListPricing)
	priv.POST("/realty/pricing", realty.CreatePricing)
	priv.PUT("/realty/pricing/:id", realty.UpdatePricing)
	priv.DELETE("/realty/pricing/:id", realty.DeletePricing)
	priv.GET("/realty/layouts", realty.ListLayouts)
	priv.POST("/realty/layouts", realty.CreateLayout)
	priv.PUT("/realty/layouts/:id", realty.UpdateLayout)
	priv.DELETE("/realty/layouts/:id", realty.DeleteLayout)

	// objekt (shaxmatka)
	priv.GET("/objekt/projects", objekt.ListProjects)
	priv.POST("/objekt/projects", objekt.CreateProject)
	priv.PUT("/objekt/projects/:id", objekt.UpdateProject)
	priv.DELETE("/objekt/projects/:id", objekt.DeleteProject)
	priv.GET("/objekt/projects/:id/blocks", objekt.ListBlocks)
	priv.POST("/objekt/blocks", objekt.CreateBlock)
	priv.PUT("/objekt/blocks/:id", objekt.UpdateBlock)
	priv.DELETE("/objekt/blocks/:id", objekt.DeleteBlock)
	priv.POST("/objekt/blocks/order", objekt.UpdateBlocksOrder)
	priv.GET("/objekt/projects/:id/apartments", objekt.ListApartments)
	priv.PUT("/objekt/apartments/:id", objekt.UpdateApartment)
	priv.PATCH("/objekt/apartments/:id/status", objekt.UpdateApartmentStatus)
	priv.POST("/objekt/upload-image", objekt.UploadImage)
	priv.GET("/objekt/export/excel", objekt.ExportExcel)
	priv.GET("/objekt/export/pdf/:id", objekt.ExportPDF)

	// sim cards
	priv.GET("/company-phones", sims.ListPhones)
	priv.POST("/company-phones", sims.AddPhone)
	priv.PUT("/company-phones/:id", sims.UpdatePhone)
	priv.DELETE("/company-phones/:id", sims.DeletePhone)
	priv.GET("/sim-cards", sims.ListSims)
	priv.POST("/sim-cards", sims.AddSim)
	priv.PUT("/sim-cards/:id", sims.UpdateSim)
	priv.DELETE("/sim-cards/:id", sims.DeleteSim)
	priv.GET("/sim-tariffs", sims.ListTariffs)
	priv.POST("/sim-tariffs", sims.AddTariff)
	priv.PUT("/sim-tariffs/:id", sims.UpdateTariff)
	priv.DELETE("/sim-tariffs/:id", sims.DeleteTariff)
	priv.GET("/tariff-payments", sims.ListPayments)
	priv.DELETE("/tariff-payments/:id", sims.DeletePayment)
	priv.GET("/sim-phones/stats", sims.Stats)
	priv.GET("/sim-phones/export/excel", sims.ExportExcel)
	adm.POST("/sim-tariffs/notify-expiring", sims.SendExpiryNotifications)

	// posts
	priv.GET("/posts", posts.List)
	priv.POST("/posts", posts.Create)
	priv.PUT("/posts/:id", posts.Update)
	priv.DELETE("/posts/:id", posts.Delete)
	priv.GET("/posts/export/excel", posts.ExportExcel)
	priv.GET("/posts/export/pdf", posts.ExportPDF)

	// chat
	priv.GET("/messages", msgs.List)
	priv.POST("/messages", msgs.Send)
	priv.PUT("/messages/:id", msgs.Update)
	priv.DELETE("/messages/:id", msgs.Delete)

	// folders/files
	priv.GET("/folders", folders.List)
	priv.POST("/folders", folders.Create)
	priv.PUT("/folders/:id", folders.Rename)
	priv.DELETE("/folders/:id", folders.Delete)
	priv.GET("/folders/:id/files", folders.ListFiles)
	priv.POST("/folders/:id/files", folders.UploadFile)
	priv.DELETE("/files/:id", folders.DeleteFile)
	priv.GET("/download/:id", folders.Download)

	// admin: ipoteka & rasrochka
	adm.POST("/banks", ipoteka.CreateBank)
	adm.PUT("/banks/:id", ipoteka.UpdateBank)
	adm.DELETE("/banks/:id", ipoteka.DeleteBank)
	adm.POST("/banks/:id/conditions", ipoteka.CreateBankCondition)
	adm.PUT("/conditions/:id", ipoteka.UpdateBankCondition)
	adm.DELETE("/conditions/:id", ipoteka.DeleteBankCondition)
	adm.POST("/installment-objects", ipoteka.CreateInstallment)
	adm.PUT("/installment-objects/:id", ipoteka.UpdateInstallment)
	adm.DELETE("/installment-objects/:id", ipoteka.DeleteInstallment)
	adm.POST("/installment-objects/:id/conditions", ipoteka.CreateInstallmentCondition)
	adm.PUT("/installment-conditions/:id", ipoteka.UpdateInstallmentCondition)
	adm.DELETE("/installment-conditions/:id", ipoteka.DeleteInstallmentCondition)

	// dashboard
	priv.GET("/dashboard/stats", dash.Stats)

	// notifications (in-app feed)
	priv.GET("/notifications", notifs.List)
	priv.GET("/notifications/unread-count", notifs.UnreadCount)
	priv.POST("/notifications/read-all", notifs.MarkAllRead)
	priv.POST("/notifications/mark-read/:id", notifs.MarkRead)
	priv.DELETE("/notifications/:id", notifs.Delete)

	// ===== v-2: WebSocket realtime =====
	priv.GET("/ws", wsH.Connect)
	priv.GET("/ws/online", wsH.Online)

	// ===== v-2: 2FA =====
	priv.GET("/2fa/status", twofa.Status)
	priv.POST("/2fa/setup", twofa.Setup)
	priv.POST("/2fa/verify", twofa.Verify)
	priv.POST("/2fa/disable", twofa.Disable)

	// ===== v-2: AI features =====
	priv.POST("/ai/assistant", aiH.Assistant)
	priv.POST("/ai/score-lead/:id", aiH.ScoreLead)
	priv.POST("/ai/summarize", aiH.Summarize)
	priv.POST("/ai/categorize", aiH.Categorize)
	priv.POST("/ai/ocr", aiH.OCR)
	adm.GET("/ai/logs", aiH.Logs)

	// ===== v-2: voice notes =====
	priv.POST("/voice/upload", voiceH.Upload)
	priv.GET("/voice", voiceH.List)
	priv.DELETE("/voice/:id", voiceH.Delete)

	// ===== v-2: virtual tours (admin CRUD) =====
	priv.GET("/tours", tourH.List)
	priv.GET("/tours/:id", tourH.Get)
	adm.POST("/tours", tourH.Create)
	adm.PUT("/tours/:id", tourH.Update)
	adm.DELETE("/tours/:id", tourH.Delete)
	adm.POST("/tours/:id/panoramas", tourH.UploadPanorama)
	adm.PUT("/panoramas/:id", tourH.UpdatePanorama)
	adm.DELETE("/panoramas/:id", tourH.DeletePanorama)
	adm.POST("/panoramas/:id/hotspots", tourH.CreateHotspot)
	adm.DELETE("/hotspots/:id", tourH.DeleteHotspot)

	// ===== v-2: website leads (admin dashboard) =====
	priv.GET("/website-leads", wlH.List)
	priv.PUT("/website-leads/:id", wlH.Update)
	priv.POST("/website-leads/:id/promote", wlH.Promote)
	adm.DELETE("/website-leads/:id", wlH.Delete)

	// ===== v-2: workflow builder + audit =====
	priv.GET("/workflows", wfH.List)
	adm.POST("/workflows", wfH.Create)
	adm.PUT("/workflows/:id", wfH.Update)
	adm.DELETE("/workflows/:id", wfH.Delete)
	priv.GET("/workflows/:id/runs", wfH.Runs)
	adm.GET("/audit-logs", wfH.AuditList)

	return r
}
