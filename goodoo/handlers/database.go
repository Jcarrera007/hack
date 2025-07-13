package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	goodooHttp "goodoo/http"
)

// DatabaseHandler handles database-related requests
type DatabaseHandler struct {
	Config *goodooHttp.RequestConfig
}

// NewDatabaseHandler creates a new database handler
func NewDatabaseHandler(config *goodooHttp.RequestConfig) *DatabaseHandler {
	return &DatabaseHandler{Config: config}
}

// ListDatabases returns available databases
func (h *DatabaseHandler) ListDatabases(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)

	// Mock database list - in real implementation, this would query the database server
	databases := []string{
		"goodoo_demo",
		"goodoo_production", 
		"goodoo_test",
	}

	req.Logger.InfoCtx(req.Context, "Database list requested")

	return c.JSON(http.StatusOK, map[string]interface{}{
		"databases": databases,
		"current":   req.GetDBName(),
	})
}

// SetDatabase sets the current database for the session
func (h *DatabaseHandler) SetDatabase(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)

	var body struct {
		Database string `json:"database"`
	}

	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if body.Database == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Database name is required",
		})
	}

	// Set database in session
	req.Session.Set("db_name", body.Database)
	// Session will be saved automatically by middleware

	req.Logger.InfoCtx(req.Context, "Database changed to: %s", body.Database)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"database": body.Database,
		"message":  "Database updated successfully",
	})
}