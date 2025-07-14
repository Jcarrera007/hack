package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	goodooHttp "goodoo/http"
	"goodoo/models"

	"github.com/labstack/echo/v4"
)

type DashboardHandler struct {
	config *goodooHttp.RequestConfig
}

type DashboardData struct {
	UserName string
	UserRole string
}

type MetricsResponse struct {
	ActiveUsers      int    `json:"active_users"`
	RequestCount     int    `json:"request_count"`
	AvgResponseTime  int    `json:"avg_response_time"`
	Status           string `json:"status"`
	SystemHealth     string `json:"system_health"`
	DatabaseSize     int    `json:"database_size_mb"`
	ActiveConnections int   `json:"active_connections"`
}

type ChartDataResponse struct {
	Requests      ChartData `json:"requests"`
	ResponseTimes ChartData `json:"response_times"`
}

type ChartData struct {
	Labels []string `json:"labels"`
	Data   []int    `json:"data"`
}

type ActivityItem struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	Level     string    `json:"level"`
}

type UserResponse struct {
	ID        uint      `json:"id"`
	Login     string    `json:"login"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	LastLogin time.Time `json:"last_login"`
	Active    bool      `json:"active"`
}

type SocialStatsResponse struct {
	ConnectedAccounts int `json:"connected_accounts"`
	PostsFetched      int `json:"posts_fetched"`
	ActiveFilters     int `json:"active_filters"`
}

type APIMetricsResponse struct {
	TotalRequests     int     `json:"total_requests"`
	SuccessRate       float64 `json:"success_rate"`
	ErrorRate         float64 `json:"error_rate"`
	AvgResponseTime   int     `json:"avg_response_time"`
}

type DatabaseInfoResponse struct {
	Status            string `json:"status"`
	ActiveConnections int    `json:"active_connections"`
	SizeMB            int    `json:"size_mb"`
}

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

type SettingsRequest struct {
	LogLevel             string `json:"log_level"`
	SessionTimeout       int    `json:"session_timeout"`
	PerformanceMonitoring bool   `json:"performance_monitoring"`
}

type CreateUserRequest struct {
	Login    string `json:"login"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Active   bool   `json:"active"`
}

// LLM Integration Types
type LLMProvider struct {
	ID       int                    `json:"id"`
	Name     string                 `json:"name"`
	Service  string                 `json:"service"`
	Active   bool                   `json:"active"`
	APIKey   string                 `json:"api_key,omitempty"`
	APIBase  string                 `json:"api_base,omitempty"`
	Config   map[string]interface{} `json:"config,omitempty"`
	Models   []LLMModel             `json:"models,omitempty"`
}

type LLMModel struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	ModelName  string `json:"model_name"`
	Active     bool   `json:"active"`
	ProviderID int    `json:"provider_id"`
	Type       string `json:"type"` // chat, embedding, etc.
}

type LLMAddonStatus struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Installed   bool   `json:"installed"`
	Active      bool   `json:"active"`
	Version     string `json:"version"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

type LLMToolsResponse struct {
	Providers []LLMProvider    `json:"providers"`
	Addons    []LLMAddonStatus `json:"addons"`
	Summary   LLMSummary       `json:"summary"`
}

type LLMSummary struct {
	TotalProviders   int `json:"total_providers"`
	ActiveProviders  int `json:"active_providers"`
	TotalModels      int `json:"total_models"`
	ActiveModels     int `json:"active_models"`
	InstalledAddons  int `json:"installed_addons"`
}

type LLMConfigRequest struct {
	ProviderID int                    `json:"provider_id"`
	Config     map[string]interface{} `json:"config"`
}

type LLMTestRequest struct {
	ProviderID int    `json:"provider_id"`
	ModelName  string `json:"model_name,omitempty"`
}

type LLMTestResponse struct {
	Success      bool   `json:"success"`
	ResponseTime int    `json:"response_time_ms"`
	Error        string `json:"error,omitempty"`
	ModelInfo    string `json:"model_info,omitempty"`
}

// Chat Integration Types
type ChatMessage struct {
	ID        string                 `json:"id"`
	Role      string                 `json:"role"` // "user" or "assistant"
	Content   string                 `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	Model     string                 `json:"model,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type ChatSession struct {
	ID        string        `json:"id"`
	UserID    int           `json:"user_id"`
	Title     string        `json:"title"`
	Model     string        `json:"model"`
	Messages  []ChatMessage `json:"messages"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Active    bool          `json:"active"`
}

type ChatRequest struct {
	Message   string `json:"message"`
	Model     string `json:"model"`
	SessionID string `json:"session_id,omitempty"`
}

type ChatResponse struct {
	ID            string    `json:"id"`
	Message       string    `json:"message"`
	Model         string    `json:"model"`
	SessionID     string    `json:"session_id"`
	Timestamp     time.Time `json:"timestamp"`
	ResponseTime  int       `json:"response_time_ms"`
	TokensUsed    int       `json:"tokens_used,omitempty"`
	FinishReason  string    `json:"finish_reason,omitempty"`
	Error         string    `json:"error,omitempty"`
}

type ChatSessionsResponse struct {
	Sessions []ChatSession `json:"sessions"`
	Total    int           `json:"total"`
}

type StreamChatResponse struct {
	Delta     string `json:"delta,omitempty"`
	Done      bool   `json:"done"`
	MessageID string `json:"message_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// User-to-User Chat Types
type UserChatMessage struct {
	ID          string    `json:"id"`
	FromUserID  int       `json:"from_user_id"`
	ToUserID    int       `json:"to_user_id"`
	Content     string    `json:"content"`
	MessageType string    `json:"message_type"` // "text", "file", "image"
	Timestamp   time.Time `json:"timestamp"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
	EditedAt    *time.Time `json:"edited_at,omitempty"`
}

type UserChatRoom struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Type         string              `json:"type"` // "direct", "group"
	Participants []UserChatParticipant `json:"participants"`
	LastMessage  *UserChatMessage    `json:"last_message,omitempty"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
	UnreadCount  int                 `json:"unread_count"`
}

type UserChatParticipant struct {
	UserID     int       `json:"user_id"`
	UserName   string    `json:"user_name"`
	UserEmail  string    `json:"user_email"`
	IsOnline   bool      `json:"is_online"`
	LastSeen   time.Time `json:"last_seen"`
	JoinedAt   time.Time `json:"joined_at"`
}

type SendUserMessageRequest struct {
	ToUserID    int    `json:"to_user_id,omitempty"`
	RoomID      string `json:"room_id,omitempty"`
	Content     string `json:"content"`
	MessageType string `json:"message_type"`
}

type UserChatResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type CreateGroupChatRequest struct {
	Name         string `json:"name"`
	ParticipantIDs []int `json:"participant_ids"`
}

type UserPresenceUpdate struct {
	UserID   int       `json:"user_id"`
	IsOnline bool      `json:"is_online"`
	LastSeen time.Time `json:"last_seen"`
}

func NewDashboardHandler(config *goodooHttp.RequestConfig) *DashboardHandler {
	return &DashboardHandler{
		config: config,
	}
}

// DashboardPage renders the main dashboard page
func (h *DashboardHandler) DashboardPage(c echo.Context) error {
	// Get user information from Goodoo request
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil {
		return echo.NewHTTPError(500, "Request context not found")
	}
	
	data := DashboardData{
		UserName: "Administrator",
		UserRole: "Admin",
	}
	
	if req.IsAuthenticated() {
		// Get actual user data from database
		db := req.GetDB()
		if db != nil {
			var user models.User
			if err := db.First(&user, req.GetUserID()).Error; err == nil {
				data.UserName = user.Name
				data.UserRole = "User" // You can extend this with actual roles
			}
		}
	}
	
	return c.Render(http.StatusOK, "dashboard.html", data)
}

// GetMetrics returns dashboard metrics
func (h *DashboardHandler) GetMetrics(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil {
		return echo.NewHTTPError(500, "Request context not found")
	}
	db := req.GetDB()
	if db == nil {
		return echo.NewHTTPError(500, "Database not available")
	}
	
	// Count total users
	var totalUsers int64
	db.Model(&models.User{}).Count(&totalUsers)
	
	// Count active users (users with recent activity - using WriteDate as proxy for last activity)
	var activeUsers int64
	db.Model(&models.User{}).Where("write_date > ? AND active = true", time.Now().Add(-24*time.Hour)).Count(&activeUsers)
	
	// Calculate system health based on database connectivity and user activity
	systemHealth := "Healthy"
	healthStatus := "healthy"
	if activeUsers == 0 {
		systemHealth = "Warning"
		healthStatus = "warning"
	}
	
	// Get database size info
	var dbSize int
	sqlDB, err := db.DB()
	if err == nil {
		stats := sqlDB.Stats()
		dbSize = stats.OpenConnections * 10 // Rough estimate
	}
	
	// Calculate average response time based on recent performance
	avgResponseTime := 85 + int(time.Now().Unix()%50) // Dynamic mock data
	
	// Request count - simulate increasing numbers
	requestCount := 1200 + int(time.Now().Unix()%500)
	
	response := MetricsResponse{
		ActiveUsers:       int(activeUsers),
		RequestCount:      requestCount,
		AvgResponseTime:   avgResponseTime,
		Status:            healthStatus,
		SystemHealth:      systemHealth,
		DatabaseSize:      dbSize,
		ActiveConnections: int(totalUsers),
	}
	
	return c.JSON(http.StatusOK, response)
}

// GetChartData returns data for dashboard charts
func (h *DashboardHandler) GetChartData(c echo.Context) error {
	// Generate realistic data for the last 24 hours
	now := time.Now()
	labels := make([]string, 24)
	requestData := make([]int, 24)
	responseData := make([]int, 24)
	
	// Base values that change throughout the day
	baseRequests := 50
	baseResponse := 120
	
	for i := 0; i < 24; i++ {
		hourTime := now.Add(time.Duration(i-23) * time.Hour)
		labels[i] = hourTime.Format("15:04")
		
		// Simulate realistic traffic patterns (higher during business hours)
		hour := hourTime.Hour()
		trafficMultiplier := 1.0
		if hour >= 9 && hour <= 17 { // Business hours
			trafficMultiplier = 2.0 + float64(hour-9)*0.1
		} else if hour >= 18 && hour <= 22 { // Evening
			trafficMultiplier = 1.5
		} else { // Night/early morning
			trafficMultiplier = 0.5
		}
		
		// Add some randomness but keep it realistic
		requests := int(float64(baseRequests) * trafficMultiplier * (0.8 + 0.4*float64(i%7)/6.0))
		response := int(float64(baseResponse) * (1.0 + 0.3*float64(i%5)/4.0))
		
		// Add some noise
		requests += int(time.Now().Unix()+int64(i)) % 20
		response += int(time.Now().Unix()+int64(i*2)) % 30
		
		requestData[i] = requests
		responseData[i] = response
	}
	
	response := ChartDataResponse{
		Requests: ChartData{
			Labels: labels,
			Data:   requestData,
		},
		ResponseTimes: ChartData{
			Labels: labels,
			Data:   responseData,
		},
	}
	
	return c.JSON(http.StatusOK, response)
}

// GetRecentActivity returns recent system activity
func (h *DashboardHandler) GetRecentActivity(c echo.Context) error {
	activities := []ActivityItem{
		{
			Timestamp: time.Now().Add(-2 * time.Minute),
			Message:   "User admin logged in",
			Level:     "INFO",
		},
		{
			Timestamp: time.Now().Add(-5 * time.Minute),
			Message:   "API endpoint /api/models called",
			Level:     "SUCCESS",
		},
		{
			Timestamp: time.Now().Add(-10 * time.Minute),
			Message:   "Database connection established",
			Level:     "INFO",
		},
		{
			Timestamp: time.Now().Add(-15 * time.Minute),
			Message:   "Session cleanup completed",
			Level:     "INFO",
		},
		{
			Timestamp: time.Now().Add(-20 * time.Minute),
			Message:   "System health check passed",
			Level:     "SUCCESS",
		},
	}
	
	return c.JSON(http.StatusOK, activities)
}

// GetUsers returns user list for user management section
func (h *DashboardHandler) GetUsers(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil {
		return echo.NewHTTPError(500, "Request context not found")
	}
	db := req.GetDB()
	if db == nil {
		return echo.NewHTTPError(500, "Database not available")
	}
	
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch users",
		})
	}
	
	response := make([]UserResponse, len(users))
	for i, user := range users {
		response[i] = UserResponse{
			ID:        user.ID,
			Login:     user.Login,
			Name:      user.Name,
			Email:     user.Email,
			LastLogin: user.WriteDate, // Using WriteDate as a proxy for last activity
			Active:    true, // You can add an Active field to the User model
		}
	}
	
	return c.JSON(http.StatusOK, response)
}

// GetSocialStats returns social media integration statistics
func (h *DashboardHandler) GetSocialStats(c echo.Context) error {
	// In a real implementation, this would query the Odoo social media module
	response := SocialStatsResponse{
		ConnectedAccounts: 3,
		PostsFetched:      1247,
		ActiveFilters:     8,
	}
	
	return c.JSON(http.StatusOK, response)
}

// GetAPIMetrics returns API performance metrics
func (h *DashboardHandler) GetAPIMetrics(c echo.Context) error {
	// In a real implementation, this would aggregate actual API metrics
	response := APIMetricsResponse{
		TotalRequests:   15432,
		SuccessRate:     98.7,
		ErrorRate:       1.3,
		AvgResponseTime: 127,
	}
	
	return c.JSON(http.StatusOK, response)
}

// GetDatabaseInfo returns database information
func (h *DashboardHandler) GetDatabaseInfo(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil {
		return echo.NewHTTPError(500, "Request context not found")
	}
	db := req.GetDB()
	if db == nil {
		return echo.NewHTTPError(500, "Database not available")
	}
	
	// Check database connection
	sqlDB, err := db.DB()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, DatabaseInfoResponse{
			Status:            "Error",
			ActiveConnections: 0,
			SizeMB:            0,
		})
	}
	
	stats := sqlDB.Stats()
	
	response := DatabaseInfoResponse{
		Status:            "Connected",
		ActiveConnections: stats.OpenConnections,
		SizeMB:            248, // This would require a database-specific query
	}
	
	return c.JSON(http.StatusOK, response)
}

// GetRecentLogs returns recent system logs
func (h *DashboardHandler) GetRecentLogs(c echo.Context) error {
	// Get optional level filter
	levelFilter := c.QueryParam("level")
	limitStr := c.QueryParam("limit")
	
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}
	
	// In a real implementation, this would read from the actual log files or database
	logs := []LogEntry{
		{
			Timestamp: time.Now().Add(-1 * time.Minute),
			Level:     "INFO",
			Message:   "Dashboard API endpoint accessed",
		},
		{
			Timestamp: time.Now().Add(-3 * time.Minute),
			Level:     "INFO",
			Message:   "User authentication successful",
		},
		{
			Timestamp: time.Now().Add(-5 * time.Minute),
			Level:     "DEBUG",
			Message:   "Database query executed in 45ms",
		},
		{
			Timestamp: time.Now().Add(-8 * time.Minute),
			Level:     "WARN",
			Message:   "High memory usage detected: 78%",
		},
		{
			Timestamp: time.Now().Add(-12 * time.Minute),
			Level:     "INFO",
			Message:   "Session cleanup completed successfully",
		},
	}
	
	// Filter by level if specified
	if levelFilter != "" && levelFilter != "all" {
		filtered := make([]LogEntry, 0)
		for _, log := range logs {
			if log.Level == levelFilter {
				filtered = append(filtered, log)
			}
		}
		logs = filtered
	}
	
	// Apply limit
	if len(logs) > limit {
		logs = logs[:limit]
	}
	
	return c.JSON(http.StatusOK, logs)
}

// GetSettings returns current system settings
func (h *DashboardHandler) GetSettings(c echo.Context) error {
	// In a real implementation, this would read from configuration
	settings := map[string]interface{}{
		"log_level":             "info",
		"session_timeout":       60,
		"performance_monitoring": true,
	}
	
	return c.JSON(http.StatusOK, settings)
}

// SaveSettings saves system settings
func (h *DashboardHandler) SaveSettings(c echo.Context) error {
	var req SettingsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}
	
	// Validate log level
	validLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true, "critical": true,
	}
	if !validLevels[req.LogLevel] {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid log level",
		})
	}
	
	// Validate session timeout
	if req.SessionTimeout < 5 || req.SessionTimeout > 480 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Session timeout must be between 5 and 480 minutes",
		})
	}
	
	// In a real implementation, you would save these settings to configuration
	h.config.Logger.Info("Settings updated: log_level=%s, session_timeout=%d, performance_monitoring=%t",
		req.LogLevel, req.SessionTimeout, req.PerformanceMonitoring)
	
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Settings saved successfully",
	})
}

// CreateUser creates a new user (admin only)
func (h *DashboardHandler) CreateUser(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil {
		return echo.NewHTTPError(500, "Request context not found")
	}

	// Check if user is authenticated
	if !req.IsAuthenticated() {
		return echo.NewHTTPError(401, "Authentication required")
	}

	// Get database connection
	db := req.GetDB()
	if db == nil {
		return echo.NewHTTPError(500, "Database not available")
	}

	// Parse request using Echo's native JSON binding
	var createReq CreateUserRequest
	if err := c.Bind(&createReq); err != nil {
		req.Logger.ErrorCtx(req.Context, "Error binding request: %v", err)
		return echo.NewHTTPError(400, "Invalid request format: "+err.Error())
	}

	// Validate required fields
	if createReq.Login == "" || createReq.Name == "" || createReq.Email == "" || createReq.Password == "" {
		return echo.NewHTTPError(400, "All fields (login, name, email, password) are required")
	}

	// Check if user already exists
	var existingUser models.User
	if err := db.Where("login = ?", createReq.Login).First(&existingUser).Error; err == nil {
		return echo.NewHTTPError(409, "User with this login already exists")
	}

	// Create the user
	user, err := models.CreateUser(db, createReq.Login, createReq.Name, createReq.Email, createReq.Password)
	if err != nil {
		req.Logger.ErrorCtx(req.Context, "Failed to create user: %v", err)
		return echo.NewHTTPError(500, "Failed to create user")
	}

	// Update active status if specified
	if !createReq.Active {
		user.Active = false
		db.Save(&user)
	}

	req.Logger.InfoCtx(req.Context, "User created: %s (ID: %d) by admin %s", user.Login, user.ID, req.GetLogin())

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "User created successfully",
		"user": map[string]interface{}{
			"id":     user.ID,
			"login":  user.Login,
			"name":   user.Name,
			"email":  user.Email,
			"active": user.Active,
		},
	})
}

// GetLLMTools returns comprehensive LLM tools information
func (h *DashboardHandler) GetLLMTools(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil {
		return echo.NewHTTPError(500, "Request context not found")
	}

	// For now, return mock data representing the available Odoo LLM addons
	// In a real implementation, this would query the Odoo database
	addons := []LLMAddonStatus{
		{Name: "llm", DisplayName: "LLM Integration Base", Installed: true, Active: true, Version: "16.0.1.4.0", Category: "Core", Description: "Base LLM integration module"},
		{Name: "llm_openai", DisplayName: "OpenAI Integration", Installed: true, Active: true, Version: "16.0.1.1.3", Category: "Provider", Description: "OpenAI provider integration"},
		{Name: "llm_anthropic", DisplayName: "Anthropic Integration", Installed: true, Active: false, Version: "16.0.1.1.0", Category: "Provider", Description: "Anthropic Claude integration"},
		{Name: "llm_ollama", DisplayName: "Ollama Integration", Installed: true, Active: true, Version: "16.0.1.0.0", Category: "Provider", Description: "Local Ollama integration"},
		{Name: "llm_chroma", DisplayName: "Chroma Vector Store", Installed: true, Active: true, Version: "16.0.1.0.0", Category: "Vector Store", Description: "ChromaDB integration"},
		{Name: "llm_knowledge", DisplayName: "Knowledge Base", Installed: true, Active: true, Version: "16.0.1.0.0", Category: "Knowledge", Description: "Knowledge management"},
		{Name: "llm_assistant", DisplayName: "AI Assistant", Installed: false, Active: false, Version: "", Category: "Interface", Description: "Conversational AI interface"},
	}

	providers := []LLMProvider{
		{ID: 1, Name: "OpenAI", Service: "openai", Active: true, APIBase: "https://api.openai.com/v1"},
		{ID: 2, Name: "Local Ollama", Service: "ollama", Active: true, APIBase: "http://localhost:11434"},
		{ID: 3, Name: "Anthropic", Service: "anthropic", Active: false},
	}

	// Calculate summary
	totalProviders := len(providers)
	activeProviders := 0
	totalModels := 0
	activeModels := 0
	installedAddons := 0

	for _, provider := range providers {
		if provider.Active {
			activeProviders++
		}
		totalModels += len(provider.Models)
		for _, model := range provider.Models {
			if model.Active {
				activeModels++
			}
		}
	}

	for _, addon := range addons {
		if addon.Installed {
			installedAddons++
		}
	}

	summary := LLMSummary{
		TotalProviders:  totalProviders,
		ActiveProviders: activeProviders,
		TotalModels:     totalModels,
		ActiveModels:    activeModels,
		InstalledAddons: installedAddons,
	}

	response := LLMToolsResponse{
		Providers: providers,
		Addons:    addons,
		Summary:   summary,
	}

	return c.JSON(http.StatusOK, response)
}

// GetLLMProviders returns LLM providers from Odoo
func (h *DashboardHandler) GetLLMProviders(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil {
		return echo.NewHTTPError(500, "Request context not found")
	}

	// Mock providers data - in real implementation, query Odoo's llm.provider model
	providers := []LLMProvider{
		{
			ID:      1,
			Name:    "OpenAI Production",
			Service: "openai",
			Active:  true,
			APIBase: "https://api.openai.com/v1",
			Models: []LLMModel{
				{ID: 1, Name: "GPT-4", ModelName: "gpt-4", Active: true, ProviderID: 1, Type: "chat"},
				{ID: 2, Name: "GPT-3.5 Turbo", ModelName: "gpt-3.5-turbo", Active: true, ProviderID: 1, Type: "chat"},
				{ID: 3, Name: "Text Embedding Ada", ModelName: "text-embedding-ada-002", Active: true, ProviderID: 1, Type: "embedding"},
			},
		},
		{
			ID:      2,
			Name:    "Local Ollama",
			Service: "ollama",
			Active:  true,
			APIBase: "http://localhost:11434",
			Models: []LLMModel{
				{ID: 4, Name: "Llama 2", ModelName: "llama2", Active: true, ProviderID: 2, Type: "chat"},
				{ID: 5, Name: "Code Llama", ModelName: "codellama", Active: false, ProviderID: 2, Type: "chat"},
			},
		},
		{
			ID:      3,
			Name:    "Anthropic Claude",
			Service: "anthropic",
			Active:  false,
			Models:  []LLMModel{},
		},
	}

	return c.JSON(http.StatusOK, providers)
}

// GetLLMModels returns available models
func (h *DashboardHandler) GetLLMModels(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil {
		return echo.NewHTTPError(500, "Request context not found")
	}

	providerID := c.QueryParam("provider_id")
	
	// Mock models data - in real implementation, query Odoo's llm.model model
	allModels := []LLMModel{
		{ID: 1, Name: "GPT-4", ModelName: "gpt-4", Active: true, ProviderID: 1, Type: "chat"},
		{ID: 2, Name: "GPT-3.5 Turbo", ModelName: "gpt-3.5-turbo", Active: true, ProviderID: 1, Type: "chat"},
		{ID: 3, Name: "Text Embedding Ada", ModelName: "text-embedding-ada-002", Active: true, ProviderID: 1, Type: "embedding"},
		{ID: 4, Name: "Llama 2", ModelName: "llama2", Active: true, ProviderID: 2, Type: "chat"},
		{ID: 5, Name: "Code Llama", ModelName: "codellama", Active: false, ProviderID: 2, Type: "chat"},
	}

	// Filter by provider if specified
	if providerID != "" {
		if pid, err := strconv.Atoi(providerID); err == nil {
			filteredModels := make([]LLMModel, 0)
			for _, model := range allModels {
				if model.ProviderID == pid {
					filteredModels = append(filteredModels, model)
				}
			}
			return c.JSON(http.StatusOK, filteredModels)
		}
	}

	return c.JSON(http.StatusOK, allModels)
}

// GetLLMAddonStatus returns status of LLM addons
func (h *DashboardHandler) GetLLMAddonStatus(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil {
		return echo.NewHTTPError(500, "Request context not found")
	}

	// Mock addon status - in real implementation, query Odoo's ir.module.module model
	addons := []LLMAddonStatus{
		{Name: "llm", DisplayName: "LLM Integration Base", Installed: true, Active: true, Version: "16.0.1.4.0", Category: "Core"},
		{Name: "llm_openai", DisplayName: "OpenAI Integration", Installed: true, Active: true, Version: "16.0.1.1.3", Category: "Provider"},
		{Name: "llm_anthropic", DisplayName: "Anthropic Integration", Installed: true, Active: false, Version: "16.0.1.1.0", Category: "Provider"},
		{Name: "llm_ollama", DisplayName: "Ollama Integration", Installed: true, Active: true, Version: "16.0.1.0.0", Category: "Provider"},
		{Name: "llm_mistral", DisplayName: "Mistral Integration", Installed: false, Active: false, Version: "", Category: "Provider"},
		{Name: "llm_chroma", DisplayName: "Chroma Vector Store", Installed: true, Active: true, Version: "16.0.1.0.0", Category: "Vector Store"},
		{Name: "llm_qdrant", DisplayName: "Qdrant Vector Store", Installed: false, Active: false, Version: "", Category: "Vector Store"},
		{Name: "llm_pgvector", DisplayName: "PostgreSQL Vector", Installed: true, Active: false, Version: "16.0.1.0.0", Category: "Vector Store"},
		{Name: "llm_knowledge", DisplayName: "Knowledge Base", Installed: true, Active: true, Version: "16.0.1.0.0", Category: "Knowledge"},
		{Name: "llm_training", DisplayName: "Model Training", Installed: false, Active: false, Version: "", Category: "Training"},
		{Name: "llm_assistant", DisplayName: "AI Assistant", Installed: true, Active: false, Version: "16.0.1.0.0", Category: "Interface"},
		{Name: "llm_replicate", DisplayName: "Replicate Integration", Installed: false, Active: false, Version: "", Category: "Specialized"},
		{Name: "llm_litellm", DisplayName: "LiteLLM Gateway", Installed: false, Active: false, Version: "", Category: "Specialized"},
		{Name: "llm_mcp", DisplayName: "MCP Integration", Installed: true, Active: false, Version: "16.0.1.0.0", Category: "Specialized"},
	}

	return c.JSON(http.StatusOK, addons)
}

// SaveLLMConfiguration saves LLM configuration
func (h *DashboardHandler) SaveLLMConfiguration(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil {
		return echo.NewHTTPError(500, "Request context not found")
	}

	var configReq LLMConfigRequest
	if err := c.Bind(&configReq); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	// In real implementation, save to Odoo's llm.provider model
	req.Logger.InfoCtx(req.Context, "LLM configuration saved for provider %d", configReq.ProviderID)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "LLM configuration saved successfully",
	})
}

// TestLLMConnection tests connection to LLM provider
func (h *DashboardHandler) TestLLMConnection(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil {
		return echo.NewHTTPError(500, "Request context not found")
	}

	var testReq LLMTestRequest
	if err := c.Bind(&testReq); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	// Simulate testing - in real implementation, actually test the provider
	start := time.Now()
	
	// Simulate different response times and success rates based on provider
	var success bool
	var errorMsg string
	var modelInfo string

	switch testReq.ProviderID {
	case 1: // OpenAI
		success = true
		modelInfo = "GPT-4 available, 8k context window"
	case 2: // Ollama
		success = true
		modelInfo = "Llama2 loaded, 4k context window"
	case 3: // Anthropic (inactive)
		success = false
		errorMsg = "Provider not configured or inactive"
	default:
		success = false
		errorMsg = "Unknown provider"
	}

	responseTime := int(time.Since(start).Milliseconds())
	if success {
		responseTime += 150 + (testReq.ProviderID * 50) // Simulate realistic response times
	}

	response := LLMTestResponse{
		Success:      success,
		ResponseTime: responseTime,
		Error:        errorMsg,
		ModelInfo:    modelInfo,
	}

	return c.JSON(http.StatusOK, response)
}

// SendChatMessage handles chat message sending and AI response
func (h *DashboardHandler) SendChatMessage(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil {
		return echo.NewHTTPError(500, "Request context not found")
	}

	var chatReq ChatRequest
	if err := c.Bind(&chatReq); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	if chatReq.Message == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Message cannot be empty",
		})
	}

	start := time.Now()
	
	// Generate unique message ID
	messageID := fmt.Sprintf("msg_%d_%d", req.GetUserID(), time.Now().UnixNano())
	
	// Simulate AI response generation based on selected model
	aiResponse, tokensUsed := h.generateAIResponse(chatReq.Message, chatReq.Model)
	
	responseTime := int(time.Since(start).Milliseconds())

	// Create session ID if not provided
	sessionID := chatReq.SessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("session_%d_%d", req.GetUserID(), time.Now().Unix())
	}

	response := ChatResponse{
		ID:           messageID,
		Message:      aiResponse,
		Model:        chatReq.Model,
		SessionID:    sessionID,
		Timestamp:    time.Now(),
		ResponseTime: responseTime,
		TokensUsed:   tokensUsed,
		FinishReason: "stop",
	}

	// Log the chat interaction
	req.Logger.InfoCtx(req.Context, "Chat message processed: user=%d, model=%s, tokens=%d, time=%dms", 
		req.GetUserID(), chatReq.Model, tokensUsed, responseTime)

	return c.JSON(http.StatusOK, response)
}

// GetChatSessions returns user's chat sessions
func (h *DashboardHandler) GetChatSessions(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil {
		return echo.NewHTTPError(500, "Request context not found")
	}

	// Mock sessions data - in real implementation, query database
	sessions := []ChatSession{
		{
			ID:        "session_1",
			UserID:    req.GetUserID(),
			Title:     "LLM Configuration Help",
			Model:     "gpt-3.5-turbo",
			CreatedAt: time.Now().Add(-2 * time.Hour),
			UpdatedAt: time.Now().Add(-30 * time.Minute),
			Active:    true,
			Messages: []ChatMessage{
				{ID: "msg_1", Role: "user", Content: "How do I configure OpenAI?", Timestamp: time.Now().Add(-2 * time.Hour)},
				{ID: "msg_2", Role: "assistant", Content: "To configure OpenAI, you need to...", Timestamp: time.Now().Add(-2*time.Hour + time.Minute)},
			},
		},
		{
			ID:        "session_2",
			UserID:    req.GetUserID(),
			Title:     "Go Programming Help",
			Model:     "gpt-4",
			CreatedAt: time.Now().Add(-1 * time.Hour),
			UpdatedAt: time.Now().Add(-10 * time.Minute),
			Active:    false,
			Messages: []ChatMessage{
				{ID: "msg_3", Role: "user", Content: "Help me write a Go function", Timestamp: time.Now().Add(-1 * time.Hour)},
			},
		},
	}

	response := ChatSessionsResponse{
		Sessions: sessions,
		Total:    len(sessions),
	}

	return c.JSON(http.StatusOK, response)
}

// GetChatSession returns a specific chat session
func (h *DashboardHandler) GetChatSession(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil {
		return echo.NewHTTPError(500, "Request context not found")
	}

	sessionID := c.Param("id")
	if sessionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Session ID is required",
		})
	}

	// Mock session data - in real implementation, query database
	session := ChatSession{
		ID:        sessionID,
		UserID:    req.GetUserID(),
		Title:     "Current Chat Session",
		Model:     "gpt-3.5-turbo",
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now(),
		Active:    true,
		Messages: []ChatMessage{
			{
				ID:        "msg_welcome",
				Role:      "assistant",
				Content:   "Hello! I'm your AI assistant. How can I help you today?",
				Timestamp: time.Now().Add(-1 * time.Hour),
				Model:     "gpt-3.5-turbo",
			},
		},
	}

	return c.JSON(http.StatusOK, session)
}

// CreateChatSession creates a new chat session
func (h *DashboardHandler) CreateChatSession(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil {
		return echo.NewHTTPError(500, "Request context not found")
	}

	var sessionReq struct {
		Title string `json:"title"`
		Model string `json:"model"`
	}

	if err := c.Bind(&sessionReq); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	sessionID := fmt.Sprintf("session_%d_%d", req.GetUserID(), time.Now().Unix())
	
	session := ChatSession{
		ID:        sessionID,
		UserID:    req.GetUserID(),
		Title:     sessionReq.Title,
		Model:     sessionReq.Model,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Active:    true,
		Messages:  []ChatMessage{},
	}

	req.Logger.InfoCtx(req.Context, "Chat session created: %s for user %d", sessionID, req.GetUserID())

	return c.JSON(http.StatusCreated, session)
}

// DeleteChatSession deletes a chat session
func (h *DashboardHandler) DeleteChatSession(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil {
		return echo.NewHTTPError(500, "Request context not found")
	}

	sessionID := c.Param("id")
	if sessionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Session ID is required",
		})
	}

	// In real implementation, delete from database
	req.Logger.InfoCtx(req.Context, "Chat session deleted: %s by user %d", sessionID, req.GetUserID())

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Session deleted successfully",
	})
}

// GetAvailableChatModels returns available models for chat
func (h *DashboardHandler) GetAvailableChatModels(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil {
		return echo.NewHTTPError(500, "Request context not found")
	}

	// Get available models from active providers
	models := []map[string]interface{}{
		{
			"id":          "gpt-3.5-turbo",
			"name":        "GPT-3.5 Turbo",
			"provider":    "OpenAI",
			"context":     4096,
			"available":   true,
			"cost_per_1k": 0.002,
		},
		{
			"id":          "gpt-4",
			"name":        "GPT-4",
			"provider":    "OpenAI",
			"context":     8192,
			"available":   true,
			"cost_per_1k": 0.03,
		},
		{
			"id":          "claude-3",
			"name":        "Claude 3",
			"provider":    "Anthropic",
			"context":     100000,
			"available":   false,
			"cost_per_1k": 0.015,
		},
		{
			"id":          "llama2",
			"name":        "Llama 2",
			"provider":    "Ollama (Local)",
			"context":     4096,
			"available":   true,
			"cost_per_1k": 0.0,
		},
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"models": models,
		"default": "gpt-3.5-turbo",
	})
}

// generateAIResponse simulates AI response generation
func (h *DashboardHandler) generateAIResponse(userMessage, model string) (string, int) {
	// Simulate different response styles based on model
	responses := map[string][]string{
		"gpt-3.5-turbo": {
			"I'd be happy to help you with that! Based on your question about %s, here's what I can tell you:",
			"That's a great question! Let me break this down for you:",
			"I can help you with that. Here's a comprehensive answer:",
		},
		"gpt-4": {
			"I'll provide you with a detailed analysis of %s. Let me walk you through this step by step:",
			"Excellent question! I'll give you a thorough explanation:",
			"Let me provide you with a comprehensive response to your inquiry about %s:",
		},
		"claude-3": {
			"I appreciate your question about %s. I'll provide a helpful and accurate response:",
			"Thank you for asking! I'll be thorough and precise in my answer:",
			"I'm here to help! Let me give you a clear explanation:",
		},
		"llama2": {
			"Based on my training, I can help you understand %s:",
			"Here's what I know about your question:",
			"I'll do my best to answer your question about %s:",
		},
	}

	// Get response templates for the model
	templates := responses[model]
	if templates == nil {
		templates = responses["gpt-3.5-turbo"] // fallback
	}

	// Select a random template
	template := templates[int(time.Now().UnixNano())%len(templates)]
	
	// Generate a contextual response
	var response string
	if strings.Contains(template, "%s") {
		// Extract key topic from user message
		topic := h.extractTopic(userMessage)
		response = fmt.Sprintf(template, topic)
	} else {
		response = template
	}

	// Add specific content based on user message
	if strings.Contains(strings.ToLower(userMessage), "llm") || strings.Contains(strings.ToLower(userMessage), "ai") {
		response += "\n\nRegarding LLM tools in your Goodoo system:\n"
		response += "• You have several LLM providers configured (OpenAI, Ollama, etc.)\n"
		response += "• Vector databases like Chroma and Qdrant are available for embeddings\n"
		response += "• The knowledge base system can help with document processing\n"
		response += "• You can monitor all LLM tools from the dashboard"
	} else if strings.Contains(strings.ToLower(userMessage), "go") || strings.Contains(strings.ToLower(userMessage), "golang") {
		response += "\n\nFor Go programming:\n"
		response += "• Go is excellent for building APIs and microservices\n"
		response += "• Your Goodoo framework is built in Go\n"
		response += "• Focus on clean error handling and proper interfaces\n"
		response += "• Use go mod for dependency management"
	} else if strings.Contains(strings.ToLower(userMessage), "config") {
		response += "\n\nFor configuration:\n"
		response += "• Check the LLM Tools tab for provider settings\n"
		response += "• API keys should be stored securely\n"
		response += "• Test connections using the built-in test functionality\n"
		response += "• Monitor system health from the overview"
	} else {
		response += "\n\nIf you need help with:\n"
		response += "• LLM configuration → Check the LLM Tools tab\n"
		response += "• System monitoring → View the Overview dashboard\n"
		response += "• API performance → Check API Metrics\n"
		response += "• Programming help → I can assist with code examples"
	}

	// Simulate token usage
	tokenCount := len(strings.Fields(userMessage+response)) * 2 // rough estimate

	return response, tokenCount
}

// extractTopic extracts key topic from user message
func (h *DashboardHandler) extractTopic(message string) string {
	message = strings.ToLower(message)
	
	// Common topics mapping
	topics := map[string]string{
		"llm":         "LLM configuration",
		"ai":          "AI tools",
		"openai":      "OpenAI setup",
		"anthropic":   "Anthropic configuration",
		"claude":      "Claude models",
		"gpt":         "GPT models",
		"go":          "Go programming",
		"golang":      "Go development",
		"api":         "API development",
		"database":    "database configuration",
		"vector":      "vector databases",
		"embedding":   "text embeddings",
		"chat":        "chat functionality",
		"config":      "system configuration",
		"dashboard":   "dashboard usage",
	}

	for keyword, topic := range topics {
		if strings.Contains(message, keyword) {
			return topic
		}
	}

	return "your question"
}

// User-to-User Chat Methods

// GetUserChatRooms returns all chat rooms for the current user
func (h *DashboardHandler) GetUserChatRooms(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil || !req.IsAuthenticated() {
		return echo.NewHTTPError(401, "Authentication required")
	}

	userID := req.GetUserID()
	db := req.GetDB()

	// Get all users for direct chat options
	var users []models.User
	db.Where("id != ?", userID).Find(&users)

	// Mock chat rooms data (in real implementation, query from database)
	rooms := []UserChatRoom{
		{
			ID:   "general",
			Name: "General Discussion",
			Type: "group",
			Participants: []UserChatParticipant{
				{UserID: 1, UserName: "Admin", UserEmail: "admin@goodoo.com", IsOnline: true, LastSeen: time.Now()},
				{UserID: 2, UserName: "User 1", UserEmail: "user1@goodoo.com", IsOnline: false, LastSeen: time.Now().Add(-15 * time.Minute)},
			},
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now().Add(-5 * time.Minute),
			UnreadCount: 3,
		},
	}

	// Add direct chat rooms for each user
	for _, user := range users {
		roomID := fmt.Sprintf("direct_%d_%d", min(userID, int(user.ID)), max(userID, int(user.ID)))
		rooms = append(rooms, UserChatRoom{
			ID:   roomID,
			Name: user.Name,
			Type: "direct",
			Participants: []UserChatParticipant{
				{UserID: int(user.ID), UserName: user.Name, UserEmail: user.Email, IsOnline: true, LastSeen: time.Now()},
			},
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			UpdatedAt:   time.Now().Add(-10 * time.Minute),
			UnreadCount: 0,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"rooms": rooms,
		"total": len(rooms),
	})
}

// Helper functions for min/max
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GetUserChatMessages returns messages for a specific chat room
func (h *DashboardHandler) GetUserChatMessages(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil || !req.IsAuthenticated() {
		return echo.NewHTTPError(401, "Authentication required")
	}

	roomID := c.Param("id")
	if roomID == "" {
		return echo.NewHTTPError(400, "Room ID is required")
	}

	// Mock messages data (in real implementation, query from database)
	messages := []UserChatMessage{
		{
			ID:          "msg_1",
			FromUserID:  1,
			ToUserID:    2,
			Content:     "Hey! How's the new dashboard coming along?",
			MessageType: "text",
			Timestamp:   time.Now().Add(-2 * time.Hour),
		},
		{
			ID:          "msg_2",
			FromUserID:  2,
			ToUserID:    1,
			Content:     "It's looking great! The LLM integration is working well.",
			MessageType: "text",
			Timestamp:   time.Now().Add(-1 * time.Hour),
			ReadAt:      &time.Time{},
		},
		{
			ID:          "msg_3",
			FromUserID:  1,
			ToUserID:    2,
			Content:     "Awesome! Can't wait to test the chat features.",
			MessageType: "text",
			Timestamp:   time.Now().Add(-30 * time.Minute),
		},
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"messages": messages,
		"room_id":  roomID,
		"total":    len(messages),
	})
}

// SendUserMessage sends a message between users
func (h *DashboardHandler) SendUserMessage(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil || !req.IsAuthenticated() {
		return echo.NewHTTPError(401, "Authentication required")
	}

	var request SendUserMessageRequest
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(400, "Invalid request format")
	}

	if request.Content == "" {
		return echo.NewHTTPError(400, "Message content is required")
	}

	userID := req.GetUserID()

	// Create new message
	message := UserChatMessage{
		ID:          fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		FromUserID:  userID,
		ToUserID:    request.ToUserID,
		Content:     request.Content,
		MessageType: request.MessageType,
		Timestamp:   time.Now(),
	}

	// In real implementation, save to database and broadcast via WebSocket

	return c.JSON(http.StatusOK, UserChatResponse{
		Success: true,
		Message: "Message sent successfully",
		Data:    message,
	})
}

// GetChatUsers returns all users available for chat
func (h *DashboardHandler) GetChatUsers(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil || !req.IsAuthenticated() {
		return echo.NewHTTPError(401, "Authentication required")
	}

	userID := req.GetUserID()
	db := req.GetDB()

	var users []models.User
	if err := db.Where("id != ?", userID).Find(&users).Error; err != nil {
		return echo.NewHTTPError(500, "Failed to fetch users")
	}

	var chatUsers []UserChatParticipant
	for _, user := range users {
		chatUsers = append(chatUsers, UserChatParticipant{
			UserID:    int(user.ID),
			UserName:  user.Name,
			UserEmail: user.Email,
			IsOnline:  true, // Mock data - in real implementation, check user presence
			LastSeen:  time.Now().Add(-5 * time.Minute),
			JoinedAt:  time.Now().Add(-24 * time.Hour), // Mock join time
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"users": chatUsers,
		"total": len(chatUsers),
	})
}

// CreateGroupChat creates a new group chat room
func (h *DashboardHandler) CreateGroupChat(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil || !req.IsAuthenticated() {
		return echo.NewHTTPError(401, "Authentication required")
	}

	var request CreateGroupChatRequest
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(400, "Invalid request format")
	}

	if request.Name == "" {
		return echo.NewHTTPError(400, "Group name is required")
	}

	userID := req.GetUserID()

	// Create new group chat room
	room := UserChatRoom{
		ID:        fmt.Sprintf("group_%d", time.Now().UnixNano()),
		Name:      request.Name,
		Type:      "group",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add creator and participants
	room.Participants = append(room.Participants, UserChatParticipant{
		UserID:   userID,
		UserName: "Current User", // Get from database in real implementation
		JoinedAt: time.Now(),
	})

	// In real implementation, save to database

	return c.JSON(http.StatusOK, UserChatResponse{
		Success: true,
		Message: "Group chat created successfully",
		Data:    room,
	})
}

// JoinChatRoom allows a user to join a chat room
func (h *DashboardHandler) JoinChatRoom(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil || !req.IsAuthenticated() {
		return echo.NewHTTPError(401, "Authentication required")
	}

	roomID := c.Param("id")
	if roomID == "" {
		return echo.NewHTTPError(400, "Room ID is required")
	}

	// In real implementation, add user to room in database

	return c.JSON(http.StatusOK, UserChatResponse{
		Success: true,
		Message: "Successfully joined chat room",
	})
}

// LeaveChatRoom allows a user to leave a chat room
func (h *DashboardHandler) LeaveChatRoom(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil || !req.IsAuthenticated() {
		return echo.NewHTTPError(401, "Authentication required")
	}

	roomID := c.Param("id")
	if roomID == "" {
		return echo.NewHTTPError(400, "Room ID is required")
	}

	// In real implementation, remove user from room in database

	return c.JSON(http.StatusOK, UserChatResponse{
		Success: true,
		Message: "Successfully left chat room",
	})
}

// GetUserPresence returns online status of users
func (h *DashboardHandler) GetUserPresence(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil || !req.IsAuthenticated() {
		return echo.NewHTTPError(401, "Authentication required")
	}

	// Mock presence data
	presence := []UserPresenceUpdate{
		{UserID: 1, IsOnline: true, LastSeen: time.Now()},
		{UserID: 2, IsOnline: false, LastSeen: time.Now().Add(-15 * time.Minute)},
		{UserID: 3, IsOnline: true, LastSeen: time.Now().Add(-2 * time.Minute)},
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"presence": presence,
		"timestamp": time.Now(),
	})
}

// UpdateUserPresence updates current user's presence status
func (h *DashboardHandler) UpdateUserPresence(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil || !req.IsAuthenticated() {
		return echo.NewHTTPError(401, "Authentication required")
	}

	var update UserPresenceUpdate
	if err := c.Bind(&update); err != nil {
		return echo.NewHTTPError(400, "Invalid request format")
	}

	userID := req.GetUserID()
	update.UserID = userID
	update.LastSeen = time.Now()

	// In real implementation, update presence in database/cache

	return c.JSON(http.StatusOK, UserChatResponse{
		Success: true,
		Message: "Presence updated successfully",
		Data:    update,
	})
}

// MarkMessageRead marks a message as read
func (h *DashboardHandler) MarkMessageRead(c echo.Context) error {
	req := goodooHttp.GetGoodooRequest(c)
	if req == nil || !req.IsAuthenticated() {
		return echo.NewHTTPError(401, "Authentication required")
	}

	messageID := c.Param("id")
	if messageID == "" {
		return echo.NewHTTPError(400, "Message ID is required")
	}

	// In real implementation, update message read status in database
	readTime := time.Now()

	return c.JSON(http.StatusOK, UserChatResponse{
		Success: true,
		Message: "Message marked as read",
		Data: map[string]interface{}{
			"message_id": messageID,
			"read_at":    readTime,
		},
	})
}

// RegisterDashboardRoutes registers all dashboard routes
func RegisterDashboardRoutes(e *echo.Echo, config *goodooHttp.RequestConfig) {
	handler := NewDashboardHandler(config)
	
	// Dashboard page (requires authentication)
	protected := e.Group("")
	protected.Use(goodooHttp.AuthenticationMiddleware(true))
	protected.Use(goodooHttp.DatabaseMiddleware(true))
	
	protected.GET("/dashboard", handler.DashboardPage)
	
	// API endpoints for dashboard data
	api := protected.Group("/api")
	api.GET("/metrics", handler.GetMetrics)
	api.GET("/metrics/charts", handler.GetChartData)
	api.GET("/metrics/api", handler.GetAPIMetrics)
	api.GET("/activity/recent", handler.GetRecentActivity)
	api.GET("/users", handler.GetUsers)
	api.GET("/social/stats", handler.GetSocialStats)
	api.GET("/database/info", handler.GetDatabaseInfo)
	api.GET("/logs/recent", handler.GetRecentLogs)
	api.GET("/settings", handler.GetSettings)
	api.POST("/settings", handler.SaveSettings)
	api.POST("/users/create", handler.CreateUser)
	
	// LLM Tools API endpoints
	api.GET("/llm/tools", handler.GetLLMTools)
	api.GET("/llm/providers", handler.GetLLMProviders)
	api.GET("/llm/models", handler.GetLLMModels)
	api.GET("/llm/addons/status", handler.GetLLMAddonStatus)
	api.POST("/llm/config", handler.SaveLLMConfiguration)
	api.POST("/llm/test", handler.TestLLMConnection)
	
	// Chat API endpoints
	api.POST("/chat/send", handler.SendChatMessage)
	api.GET("/chat/sessions", handler.GetChatSessions)
	api.GET("/chat/session/:id", handler.GetChatSession)
	api.POST("/chat/session/new", handler.CreateChatSession)
	api.DELETE("/chat/session/:id", handler.DeleteChatSession)
	api.GET("/chat/models", handler.GetAvailableChatModels)
	
	// User-to-User Chat API endpoints
	api.GET("/user-chat/rooms", handler.GetUserChatRooms)
	api.GET("/user-chat/room/:id/messages", handler.GetUserChatMessages)
	api.POST("/user-chat/send", handler.SendUserMessage)
	api.GET("/user-chat/users", handler.GetChatUsers)
	api.POST("/user-chat/room/create", handler.CreateGroupChat)
	api.POST("/user-chat/room/:id/join", handler.JoinChatRoom)
	api.POST("/user-chat/room/:id/leave", handler.LeaveChatRoom)
	api.GET("/user-chat/presence", handler.GetUserPresence)
	api.POST("/user-chat/presence", handler.UpdateUserPresence)
	api.POST("/user-chat/message/:id/read", handler.MarkMessageRead)
}