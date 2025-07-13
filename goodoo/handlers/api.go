package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"goodoo/api"
	goodooHttp "goodoo/http"
	"goodoo/logging"
)

// APIHandler provides HTTP handlers for API calls
type APIHandler struct {
	registry *api.APIRegistry
	logger   *logging.Logger
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(registry *api.APIRegistry) *APIHandler {
	return &APIHandler{
		registry: registry,
		logger:   logging.GetLogger("goodoo.api.handler"),
	}
}

// CallMethod handles API method calls via HTTP
func (h *APIHandler) CallMethod(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	ctx := req.Context

	// Parse request
	var call api.APICall
	if err := c.Bind(&call); err != nil {
		h.logger.ErrorCtx(ctx, "Failed to parse API call: %v", err)
		return c.JSON(http.StatusBadRequest, api.APIResponse{
			Success: false,
			Error:   "Invalid request format",
		})
	}

	// Log the API call
	h.logger.InfoCtx(ctx, "API call: %s.%s", call.ModelName, call.Method)

	// Execute the call
	response := h.registry.ExecuteCall(ctx, &call, req)

	// Return appropriate HTTP status
	status := http.StatusOK
	if !response.Success {
		status = http.StatusBadRequest
		if strings.Contains(response.Error, "Access denied") ||
			strings.Contains(response.Error, "not accessible") {
			status = http.StatusForbidden
		}
		if strings.Contains(response.Error, "not found") {
			status = http.StatusNotFound
		}
	}

	return c.JSON(status, response)
}

// GetModelMethods returns available methods for a model
func (h *APIHandler) GetModelMethods(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	ctx := req.Context

	modelName := c.Param("model")
	if modelName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Model name is required",
		})
	}

	h.logger.InfoCtx(ctx, "Getting methods for model: %s", modelName)

	// Get public methods only for security
	methods := h.registry.GetPublicMethods(modelName)
	if methods == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Model not found",
		})
	}

	// Convert to response format
	response := make(map[string]interface{})
	for name, method := range methods {
		response[name] = map[string]interface{}{
			"type":       string(method.Type),
			"help":       method.Help,
			"constrains": method.Constrains,
			"depends":    method.Depends,
			"onchange":   method.OnChange,
			"returns":    method.Returns,
			"groups":     method.Groups,
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"model":   modelName,
		"methods": response,
	})
}

// GetMethodInfo returns detailed information about a specific method
func (h *APIHandler) GetMethodInfo(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	ctx := req.Context

	modelName := c.Param("model")
	methodName := c.Param("method")

	if modelName == "" || methodName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Model and method names are required",
		})
	}

	h.logger.InfoCtx(ctx, "Getting info for method: %s.%s", modelName, methodName)

	info := h.registry.GetMethodInfo(modelName, methodName)
	if info == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Method not found",
		})
	}

	return c.JSON(http.StatusOK, info)
}

// CallModelMethod handles calls to model-level methods via URL
func (h *APIHandler) CallModelMethod(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	ctx := req.Context

	modelName := c.Param("model")
	methodName := c.Param("method")

	if modelName == "" || methodName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Model and method names are required",
		})
	}

	// Parse arguments from query params or request body
	var args []interface{}
	var kwargs map[string]interface{}
	var callContext map[string]interface{}

	// For GET requests, use query parameters
	if c.Request().Method == "GET" {
		// Initialize kwargs
		kwargs = make(map[string]interface{})
		// Parse simple arguments from query
		for key, values := range c.QueryParams() {
			if key == "context" {
				// Skip context, handle separately
				continue
			}
			if len(values) == 1 {
				kwargs[key] = values[0]
			} else {
				kwargs[key] = values
			}
		}
	} else {
		// For POST/PUT, parse JSON body
		var body map[string]interface{}
		if err := c.Bind(&body); err == nil {
			if argsData, exists := body["args"]; exists {
				if argsList, ok := argsData.([]interface{}); ok {
					args = argsList
				}
			}
			if kwargsData, exists := body["kwargs"]; exists {
				if kwargsMap, ok := kwargsData.(map[string]interface{}); ok {
					kwargs = kwargsMap
				}
			}
			if contextData, exists := body["context"]; exists {
				if contextMap, ok := contextData.(map[string]interface{}); ok {
					callContext = contextMap
				}
			}
		}
	}

	// Build API call
	call := &api.APICall{
		ModelName: modelName,
		Method:    methodName,
		Args:      args,
		Kwargs:    kwargs,
		Context:   callContext,
	}

	h.logger.InfoCtx(ctx, "URL API call: %s.%s", modelName, methodName)

	// Execute the call
	response := h.registry.ExecuteCall(ctx, call, req)

	// Return appropriate HTTP status
	status := http.StatusOK
	if !response.Success {
		status = http.StatusBadRequest
		if strings.Contains(response.Error, "Access denied") ||
			strings.Contains(response.Error, "not accessible") {
			status = http.StatusForbidden
		}
		if strings.Contains(response.Error, "not found") {
			status = http.StatusNotFound
		}
	}

	return c.JSON(status, response)
}

// CallRecordMethod handles calls to record-level methods via URL
func (h *APIHandler) CallRecordMethod(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	ctx := req.Context

	modelName := c.Param("model")
	methodName := c.Param("method")
	idsParam := c.Param("ids")

	if modelName == "" || methodName == "" || idsParam == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Model name, method name, and IDs are required",
		})
	}

	// Parse IDs
	var ids []int
	idStrings := strings.Split(idsParam, ",")
	for _, idStr := range idStrings {
		if id, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
			ids = append(ids, id)
		}
	}

	if len(ids) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid IDs format",
		})
	}

	// Parse arguments similar to model method
	var args []interface{}
	var kwargs map[string]interface{}
	var callContext map[string]interface{}

	if c.Request().Method == "GET" {
		kwargs = make(map[string]interface{})
		for key, values := range c.QueryParams() {
			if key == "context" {
				continue
			}
			if len(values) == 1 {
				kwargs[key] = values[0]
			} else {
				kwargs[key] = values
			}
		}
	} else {
		var body map[string]interface{}
		if err := c.Bind(&body); err == nil {
			if argsData, exists := body["args"]; exists {
				if argsList, ok := argsData.([]interface{}); ok {
					args = argsList
				}
			}
			if kwargsData, exists := body["kwargs"]; exists {
				if kwargsMap, ok := kwargsData.(map[string]interface{}); ok {
					kwargs = kwargsMap
				}
			}
			if contextData, exists := body["context"]; exists {
				if contextMap, ok := contextData.(map[string]interface{}); ok {
					callContext = contextMap
				}
			}
		}
	}

	// Build API call
	call := &api.APICall{
		ModelName: modelName,
		Method:    methodName,
		Args:      args,
		Kwargs:    kwargs,
		Context:   callContext,
		IDs:       ids,
	}

	h.logger.InfoCtx(ctx, "Record API call: %s.%s on IDs %v", modelName, methodName, ids)

	// Execute the call
	response := h.registry.ExecuteCall(ctx, call, req)

	// Return appropriate HTTP status
	status := http.StatusOK
	if !response.Success {
		status = http.StatusBadRequest
		if strings.Contains(response.Error, "Access denied") ||
			strings.Contains(response.Error, "not accessible") {
			status = http.StatusForbidden
		}
		if strings.Contains(response.Error, "not found") {
			status = http.StatusNotFound
		}
	}

	return c.JSON(status, response)
}

// RegisterRoutes registers API routes with Echo
func (h *APIHandler) RegisterRoutes(e *echo.Echo) {
	// API group
	api := e.Group("/api")

	// Generic API call endpoint
	api.POST("/call", h.CallMethod)

	// Model methods
	api.GET("/models/:model/methods", h.GetModelMethods)
	api.GET("/models/:model/methods/:method", h.GetMethodInfo)
	api.Any("/models/:model/:method", h.CallModelMethod)

	// Record methods  
	api.Any("/models/:model/:ids/:method", h.CallRecordMethod)

	h.logger.Info("Registered API routes")
}

// Convenience function for default handler
func RegisterAPIRoutes(e *echo.Echo) {
	handler := NewAPIHandler(api.DefaultAPIRegistry)
	handler.RegisterRoutes(e)
}