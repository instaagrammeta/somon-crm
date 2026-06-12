package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/instaagrammeta/somon-crm/backend/internal/middleware"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
)

// WSHandler exposes a single endpoint /api/ws that upgrades to a WebSocket
// connection. Authentication is handled by the regular auth middleware before
// this handler runs (the JWT travels as `?token=` query param so browsers can
// open the socket without setting headers).
type WSHandler struct{ *App }

func NewWSHandler(a *App) *WSHandler { return &WSHandler{a} }

// Connect upgrades the HTTP request and hands it to the hub.
func (h *WSHandler) Connect(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	if uid == 0 {
		utils.ErrorResp(c, http.StatusUnauthorized, "auth.unauthorized")
		return
	}
	role := middleware.CurrentRole(c)
	if h.Hub == nil {
		utils.ErrorResp(c, http.StatusServiceUnavailable, "error.realtime_unavailable")
		return
	}
	if err := h.Hub.Serve(c.Writer, c.Request, uid, role); err != nil {
		// Once Serve has called Upgrade we must NOT touch the response writer
		// again; on error we just close the connection silently.
		return
	}
}

// Online returns the list of currently-connected user ids (admin tooling).
func (h *WSHandler) Online(c *gin.Context) {
	if h.Hub == nil {
		c.JSON(http.StatusOK, []uint{})
		return
	}
	c.JSON(http.StatusOK, h.Hub.OnlineUsers())
}
