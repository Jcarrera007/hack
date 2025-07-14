package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	goodooHttp "goodoo/http"
	"goodoo/models"
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

	// Get database connection
	db := req.GetDB()
	if db == nil {
		req.Logger.ErrorCtx(req.Context, "Database connection not available")
		return echo.NewHTTPError(http.StatusInternalServerError, "Database connection error")
	}

	// Find user by login
	user, err := models.FindUserByLogin(db, login)
	if err != nil {
		req.Logger.WarningCtx(req.Context, "User not found: %s", login)
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	}

	// Check password
	if !user.CheckPassword(password) {
		req.Logger.WarningCtx(req.Context, "Invalid password for user: %s", login)
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	}

	// Authenticate user
	if err := req.Authenticate(database, login, int(user.ID)); err != nil {
		req.Logger.ErrorCtx(req.Context, "Authentication failed: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Authentication failed")
	}

	req.Logger.InfoCtx(req.Context, "User %s successfully authenticated", login)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"user_id": user.ID,
		"login":   login,
		"name":    user.Name,
		"email":   user.Email,
		"db":      database,
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Request context not found")
	}

	if !req.IsAuthenticated() {
		// If not authenticated, redirect to login for GET requests
		if c.Request().Method == "GET" {
			return c.Redirect(http.StatusFound, "/login")
		}
		return echo.NewHTTPError(http.StatusBadRequest, "Not authenticated")
	}

	oldLogin := req.GetLogin()
	req.Logout(false) // Don't keep database

	req.Logger.InfoCtx(req.Context, "User %s logged out", oldLogin)

	// For GET requests (from dashboard logout link), redirect to login page
	if c.Request().Method == "GET" {
		return c.Redirect(http.StatusFound, "/login")
	}

	// For POST requests (API calls), return JSON response
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



