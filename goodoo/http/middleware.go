package http

import (
	"time"

	"github.com/labstack/echo/v4"
	"goodoo/logging"
)

// RequestMiddleware creates middleware for handling Goodoo requests with session support
func RequestMiddleware(config *RequestConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Create Goodoo request wrapper
			req := NewRequest(c, config)
			
			// Add request to Echo context
			c.Set("goodoo_request", req)
			
			// Log request start
			req.Logger.DebugCtx(req.Context, "Request started: %s %s", 
				req.HTTPRequest.Method, req.HTTPRequest.URL.Path)
			
			// Process request
			err := next(c)
			
			// Save session if dirty
			if saveErr := req.SaveSession(config.SessionStore); saveErr != nil {
				req.Logger.ErrorCtx(req.Context, "Failed to save session: %v", saveErr)
			}
			
			// Log request completion
			req.LogRequest()
			
			return err
		}
	}
}

// AuthenticationMiddleware provides authentication checking
func AuthenticationMiddleware(required bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := GetGoodooRequest(c)
			if req == nil {
				return echo.NewHTTPError(500, "Goodoo request not found")
			}
			
			if required && !req.IsAuthenticated() {
				req.Logger.WarningCtx(req.Context, "Unauthenticated access attempt to %s", 
					req.HTTPRequest.URL.Path)
				return echo.NewHTTPError(401, "Authentication required")
			}
			
			if req.IsAuthenticated() {
				req.Logger.DebugCtx(req.Context, "Authenticated request from user %s (ID: %d)",
					req.GetLogin(), req.GetUserID())
			}
			
			return next(c)
		}
	}
}

// DatabaseMiddleware ensures database connection
func DatabaseMiddleware(required bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := GetGoodooRequest(c)
			if req == nil {
				return echo.NewHTTPError(500, "Goodoo request not found")
			}
			
			if required && req.GetDBName() == "" {
				req.Logger.WarningCtx(req.Context, "Database required but not set for %s", 
					req.HTTPRequest.URL.Path)
				return echo.NewHTTPError(400, "Database required")
			}
			
			if req.GetDBName() != "" {
				req.Logger.DebugCtx(req.Context, "Request using database: %s", req.GetDBName())
			}
			
			return next(c)
		}
	}
}

// SessionCleanupMiddleware periodically cleans up expired sessions
func SessionCleanupMiddleware(store SessionStore, interval time.Duration) echo.MiddlewareFunc {
	ticker := time.NewTicker(interval)
	
	go func() {
		for range ticker.C {
			if err := store.Cleanup(); err != nil {
				logging.Error("Session cleanup failed: %v", err)
			} else {
				logging.Debug("Session cleanup completed")
			}
		}
	}()
	
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return next
	}
}

// GetGoodooRequest retrieves the Goodoo request from Echo context
func GetGoodooRequest(c echo.Context) *Request {
	if req, ok := c.Get("goodoo_request").(*Request); ok {
		return req
	}
	return nil
}

// RequestLoggingMiddleware provides detailed request logging
func RequestLoggingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := GetGoodooRequest(c)
			if req == nil {
				return next(c)
			}
			
			start := time.Now()
			
			// Log request details
			req.Logger.InfoCtx(req.Context, "Request: %s %s from %s (User-Agent: %s)",
				req.HTTPRequest.Method,
				req.HTTPRequest.URL.Path,
				req.RemoteAddr,
				req.UserAgent,
			)
			
			// Process request
			err := next(c)
			
			// Log response details
			duration := time.Since(start)
			status := c.Response().Status
			
			if err != nil {
				req.Logger.ErrorCtx(req.Context, "Request failed: %s %s - Status: %d - Duration: %v - Error: %v",
					req.HTTPRequest.Method,
					req.HTTPRequest.URL.Path,
					status,
					duration,
					err,
				)
			} else {
				req.Logger.InfoCtx(req.Context, "Request completed: %s %s - Status: %d - Duration: %v",
					req.HTTPRequest.Method,
					req.HTTPRequest.URL.Path,
					status,
					duration,
				)
			}
			
			return err
		}
	}
}

// SecurityMiddleware adds security headers and checks
func SecurityMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := GetGoodooRequest(c)
			
			// Add security headers
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("X-Frame-Options", "DENY")
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			
			// Log security-related information
			if req != nil {
				req.Logger.DebugCtx(req.Context, "Security headers added for %s", req.HTTPRequest.URL.Path)
			}
			
			return next(c)
		}
	}
}

// ErrorHandlingMiddleware provides enhanced error handling with logging
func ErrorHandlingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			
			if err != nil {
				req := GetGoodooRequest(c)
				if req != nil {
					req.Logger.ErrorCtx(req.Context, "Request error: %v", err)
				}
				
				// Handle different error types
				if he, ok := err.(*echo.HTTPError); ok {
					return he
				}
				
				// Convert to HTTP error
				return echo.NewHTTPError(500, "Internal Server Error")
			}
			
			return nil
		}
	}
}