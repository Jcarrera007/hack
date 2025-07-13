package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	goodooHttp "goodoo/http"
)

// SessionHandler handles session management
type SessionHandler struct {
	Config *goodooHttp.RequestConfig
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(config *goodooHttp.RequestConfig) *SessionHandler {
	return &SessionHandler{Config: config}
}

// GetSession returns current session data
func (h *SessionHandler) GetSession(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)

	req.Logger.DebugCtx(req.Context, "Session data requested")

	return c.JSON(http.StatusOK, map[string]interface{}{
		"sid":           req.Session.SID,
		"authenticated": req.IsAuthenticated(),
		"user_id":       req.GetUserID(),
		"login":         req.GetLogin(),
		"db":            req.GetDBName(),
		"context":       req.Session.GetContext(),
		"created_at":    req.Session.CreatedAt,
		"last_accessed": req.Session.LastAccessed,
	})
}

// ClearSession clears all session data
func (h *SessionHandler) ClearSession(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)

	req.Logger.InfoCtx(req.Context, "Session cleared by request")

	req.Session.Clear()

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Session cleared",
	})
}

// SetSessionData sets data in the session
func (h *SessionHandler) SetSessionData(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)

	var body struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
	}

	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if body.Key == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Key is required",
		})
	}

	req.Session.Set(body.Key, body.Value)

	req.Logger.DebugCtx(req.Context, "Session data set: %s", body.Key)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"key":     body.Key,
		"value":   body.Value,
	})
}