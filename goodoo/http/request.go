package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"goodoo/database"
	"goodoo/logging"
	"gorm.io/gorm"
)

// Request wraps the HTTP request with session and context utilities (like Odoo's Request)
type Request struct {
	// Echo context
	Echo echo.Context
	
	// HTTP request
	HTTPRequest *http.Request
	
	// Session
	Session *Session
	
	// Database name
	DB string
	
	// Request parameters
	Params map[string]interface{}
	
	// Request context
	Context context.Context
	
	// Logger
	Logger *logging.Logger
	
	// Start time for performance tracking
	StartTime time.Time
	
	// User agent info
	UserAgent string
	
	// Remote address
	RemoteAddr string
	
	// Registry/Environment (placeholder for future ORM integration)
	Registry interface{}
	Env      interface{}
}

// RequestConfig holds configuration for request handling
type RequestConfig struct {
	SessionStore     SessionStore
	DefaultDBName    string
	SessionCookieName string
	Logger           *logging.Logger
}

// NewRequest creates a new Request wrapper from Echo context
func NewRequest(c echo.Context, config *RequestConfig) *Request {
	req := &Request{
		Echo:        c,
		HTTPRequest: c.Request(),
		Params:      make(map[string]interface{}),
		Context:     c.Request().Context(),
		Logger:      config.Logger,
		StartTime:   time.Now(),
		UserAgent:   c.Request().UserAgent(),
		RemoteAddr:  c.RealIP(),
	}
	
	// Initialize session
	req.initSession(config)
	
	// Parse request parameters
	req.parseParams()
	
	// Add request context
	req.Context = req.addRequestContext(req.Context)
	
	return req
}

// initSession initializes the session for this request
func (r *Request) initSession(config *RequestConfig) {
	cookieName := config.SessionCookieName
	if cookieName == "" {
		cookieName = "goodoo_session"
	}
	
	// Get session ID from cookie
	cookie, err := r.HTTPRequest.Cookie(cookieName)
	var sid string
	if err == nil && cookie != nil {
		sid = cookie.Value
	}
	
	// Get or create session
	if sid != "" && config.SessionStore.IsValidKey(sid) {
		r.Session = config.SessionStore.Get(sid)
	} else {
		r.Session = config.SessionStore.New()
		// Set session cookie
		r.setSessionCookie(cookieName, r.Session.SID)
	}
	
	// Determine database name
	r.DB = r.determineDatabase(config.DefaultDBName)
	
	// Update session context
	r.Session.UpdateContext(map[string]interface{}{
		"request_id":  r.generateRequestID(),
		"user_agent":  r.UserAgent,
		"remote_addr": r.RemoteAddr,
		"path":        r.HTTPRequest.URL.Path,
		"method":      r.HTTPRequest.Method,
	})
	
	r.Session.Touch()
}

// parseParams extracts and parses request parameters
func (r *Request) parseParams() {
	// Parse query parameters
	for key, values := range r.HTTPRequest.URL.Query() {
		if len(values) == 1 {
			r.Params[key] = values[0]
		} else {
			r.Params[key] = values
		}
	}
	
	// Parse form data for POST requests
	if r.HTTPRequest.Method == "POST" {
		contentType := r.HTTPRequest.Header.Get("Content-Type")
		
		if strings.Contains(contentType, "application/json") {
			r.parseJSONParams()
		} else if strings.Contains(contentType, "application/x-www-form-urlencoded") {
			r.parseFormParams()
		} else if strings.Contains(contentType, "multipart/form-data") {
			r.parseMultipartParams()
		}
	}
}

// parseJSONParams parses JSON request body
func (r *Request) parseJSONParams() {
	body, err := io.ReadAll(r.HTTPRequest.Body)
	if err != nil {
		r.Logger.ErrorCtx(r.Context, "Failed to read JSON body: %v", err)
		return
	}
	
	var jsonData map[string]interface{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		r.Logger.ErrorCtx(r.Context, "Failed to parse JSON body: %v", err)
		return
	}
	
	for key, value := range jsonData {
		r.Params[key] = value
	}
}

// parseFormParams parses form-encoded parameters
func (r *Request) parseFormParams() {
	if err := r.HTTPRequest.ParseForm(); err != nil {
		r.Logger.ErrorCtx(r.Context, "Failed to parse form: %v", err)
		return
	}
	
	for key, values := range r.HTTPRequest.PostForm {
		if len(values) == 1 {
			r.Params[key] = values[0]
		} else {
			r.Params[key] = values
		}
	}
}

// parseMultipartParams parses multipart form data
func (r *Request) parseMultipartParams() {
	if err := r.HTTPRequest.ParseMultipartForm(32 << 20); err != nil { // 32 MB max
		r.Logger.ErrorCtx(r.Context, "Failed to parse multipart form: %v", err)
		return
	}
	
	if r.HTTPRequest.MultipartForm != nil {
		for key, values := range r.HTTPRequest.MultipartForm.Value {
			if len(values) == 1 {
				r.Params[key] = values[0]
			} else {
				r.Params[key] = values
			}
		}
	}
}

// addRequestContext adds request-specific information to context
func (r *Request) addRequestContext(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, "request_id", r.generateRequestID())
	ctx = context.WithValue(ctx, "session_id", r.Session.SID)
	ctx = context.WithValue(ctx, "dbname", r.DB)
	ctx = context.WithValue(ctx, "user_id", r.Session.UserID)
	ctx = context.WithValue(ctx, "remote_addr", r.RemoteAddr)
	ctx = context.WithValue(ctx, "user_agent", r.UserAgent)
	ctx = context.WithValue(ctx, "start_time", r.StartTime)
	
	return ctx
}

// determineDatabase determines which database to use for this request
func (r *Request) determineDatabase(defaultDB string) string {
	// Check session first
	if r.Session.DBName != "" {
		return r.Session.DBName
	}
	
	// Check URL parameter
	if dbParam := r.HTTPRequest.URL.Query().Get("db"); dbParam != "" {
		return dbParam
	}
	
	// Use default
	return defaultDB
}

// setSessionCookie sets the session cookie
func (r *Request) setSessionCookie(name, value string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.HTTPRequest.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((24 * time.Hour).Seconds()), // 24 hours
	}
	
	r.Echo.SetCookie(cookie)
}

// generateRequestID generates a unique request ID
func (r *Request) generateRequestID() string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), r.Session.SID[:8])
}

// GetParam retrieves a parameter value with type conversion
func (r *Request) GetParam(key string) (interface{}, bool) {
	value, exists := r.Params[key]
	return value, exists
}

// GetStringParam retrieves a string parameter
func (r *Request) GetStringParam(key string, defaultValue ...string) string {
	if value, exists := r.Params[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
		return fmt.Sprintf("%v", value)
	}
	
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

// GetIntParam retrieves an integer parameter
func (r *Request) GetIntParam(key string, defaultValue ...int) int {
	if value, exists := r.Params[key]; exists {
		switch v := value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

// GetBoolParam retrieves a boolean parameter
func (r *Request) GetBoolParam(key string, defaultValue ...bool) bool {
	if value, exists := r.Params[key]; exists {
		switch v := value.(type) {
		case bool:
			return v
		case string:
			return v == "true" || v == "1" || v == "on" || v == "yes"
		case int:
			return v != 0
		case float64:
			return v != 0
		}
	}
	
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return false
}

// UpdateEnvironment updates the request environment (placeholder for ORM integration)
func (r *Request) UpdateEnvironment(userID int, context map[string]interface{}) {
	// Update session
	if userID != 0 {
		r.Session.UserID = userID
	}
	
	if context != nil {
		r.Session.UpdateContext(context)
	}
	
	// Update request context
	r.Context = r.addRequestContext(r.Context)
	
	r.Logger.DebugCtx(r.Context, "Environment updated for user %d", userID)
}

// Authenticate authenticates the user and updates the session
func (r *Request) Authenticate(dbname, login string, userID int) error {
	r.Session.Authenticate(dbname, login, userID)
	r.DB = dbname
	
	// Update request context
	r.Context = r.addRequestContext(r.Context)
	
	r.Logger.InfoCtx(r.Context, "User authenticated: %s (ID: %d) on database %s", login, userID, dbname)
	return nil
}

// Logout logs out the current user
func (r *Request) Logout(keepDB bool) {
	oldUserID := r.Session.UserID
	oldLogin := r.Session.Login
	
	r.Session.Logout(keepDB)
	
	if !keepDB {
		r.DB = ""
	}
	
	// Update request context
	r.Context = r.addRequestContext(r.Context)
	
	r.Logger.InfoCtx(r.Context, "User logged out: %s (ID: %d)", oldLogin, oldUserID)
}

// IsAuthenticated checks if the current request is authenticated
func (r *Request) IsAuthenticated() bool {
	return r.Session.IsAuthenticated()
}

// GetUserID returns the current user ID
func (r *Request) GetUserID() int {
	return r.Session.UserID
}

// GetLogin returns the current user login
func (r *Request) GetLogin() string {
	return r.Session.Login
}

// GetDBName returns the current database name
func (r *Request) GetDBName() string {
	return r.DB
}

// GetRequestID returns the unique request ID
func (r *Request) GetRequestID() string {
	if rid := r.Context.Value("request_id"); rid != nil {
		return rid.(string)
	}
	return ""
}

// SaveSession saves the session if it's dirty
func (r *Request) SaveSession(store SessionStore) error {
	if r.Session.IsDirty && r.Session.CanSave {
		return store.Save(r.Session)
	}
	return nil
}

// GetElapsedTime returns the time elapsed since request start
func (r *Request) GetElapsedTime() time.Duration {
	return time.Since(r.StartTime)
}

// AddToContext adds a value to the request context
func (r *Request) AddToContext(key string, value interface{}) {
	r.Context = context.WithValue(r.Context, key, value)
}

// GetFromContext retrieves a value from the request context
func (r *Request) GetFromContext(key string) interface{} {
	return r.Context.Value(key)
}

// GetDB returns the GORM database instance for the current database
func (r *Request) GetDB() *gorm.DB {
	if r.DB == "" {
		return nil
	}
	
	db, err := database.GetDatabase(r.DB)
	if err != nil {
		r.Logger.ErrorCtx(r.Context, "Failed to get database connection for %s: %v", r.DB, err)
		return nil
	}
	
	return db
}

// LogRequest logs request information
func (r *Request) LogRequest() {
	r.Logger.InfoCtx(r.Context, "%s %s - User: %s (ID: %d) - DB: %s - Duration: %v",
		r.HTTPRequest.Method,
		r.HTTPRequest.URL.Path,
		r.GetLogin(),
		r.GetUserID(),
		r.GetDBName(),
		r.GetElapsedTime(),
	)
}