package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"
	goodooHttp "goodoo/http"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	Config *goodooHttp.RequestConfig
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(config *goodooHttp.RequestConfig) *HealthHandler {
	return &HealthHandler{Config: config}
}

// Health returns basic health status
func (h *HealthHandler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "goodoo",
	})
}

// DetailedHealth returns detailed health information
func (h *HealthHandler) DetailedHealth(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "goodoo",
		"version":   "1.0.0",
		"uptime":    time.Since(req.StartTime),
		"memory": map[string]interface{}{
			"alloc_mb":      bToMb(m.Alloc),
			"total_alloc_mb": bToMb(m.TotalAlloc),
			"sys_mb":        bToMb(m.Sys),
			"num_gc":        m.NumGC,
		},
		"goroutines": runtime.NumGoroutine(),
	}

	if req.Session != nil {
		health["session_id"] = req.Session.SID
	}

	req.Logger.InfoCtx(req.Context, "Detailed health check requested")

	return c.JSON(http.StatusOK, health)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}