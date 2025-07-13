package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	goodooHttp "goodoo/http"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	Config *goodooHttp.RequestConfig
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(config *goodooHttp.RequestConfig) *AuthHandler {
	return &AuthHandler{Config: config}
}

// Login handles user login
func (h *AuthHandler) Login(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Request context not found")
	}

	// Parse login parameters
	login := req.GetStringParam("login")
	password := req.GetStringParam("password")
	database := req.GetStringParam("db", req.GetDBName())

	if login == "" || password == "" {
		req.Logger.WarningCtx(req.Context, "Login attempt with missing credentials")
		return echo.NewHTTPError(http.StatusBadRequest, "Login and password required")
	}

	req.Logger.InfoCtx(req.Context, "Login attempt for user: %s on database: %s", login, database)

	// TODO: Implement actual authentication logic
	// For now, simulate authentication
	if login == "admin" && password == "admin" {
		userID := 1
		if err := req.Authenticate(database, login, userID); err != nil {
			req.Logger.ErrorCtx(req.Context, "Authentication failed: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Authentication failed")
		}

		req.Logger.InfoCtx(req.Context, "User %s successfully authenticated", login)

		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"user_id": userID,
			"login":   login,
			"db":      database,
		})
	}

	req.Logger.WarningCtx(req.Context, "Invalid credentials for user: %s", login)
	return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
}

// Logout handles user logout
func (h *AuthHandler) Logout(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Request context not found")
	}

	if !req.IsAuthenticated() {
		return echo.NewHTTPError(http.StatusBadRequest, "Not authenticated")
	}

	oldLogin := req.GetLogin()
	req.Logout(false) // Don't keep database

	req.Logger.InfoCtx(req.Context, "User %s logged out", oldLogin)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Logged out successfully",
	})
}

// SessionInfo returns current session information
func (h *AuthHandler) SessionInfo(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Request context not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"authenticated": req.IsAuthenticated(),
		"user_id":       req.GetUserID(),
		"login":         req.GetLogin(),
		"db":            req.GetDBName(),
		"session_id":    req.Session.SID,
		"context":       req.Session.GetContext(),
		"request_id":    req.GetRequestID(),
	})
}



