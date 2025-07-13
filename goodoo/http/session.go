package http

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SessionStore interface for session storage backends
type SessionStore interface {
	New() *Session
	Get(sid string) *Session
	Save(session *Session) error
	Delete(sid string) error
	IsValidKey(sid string) bool
	Cleanup() error
}

// Session represents a user session with persistent data (like Odoo's Session)
type Session struct {
	SID          string                 `json:"sid"`
	Data         map[string]interface{} `json:"data"`
	IsDirty      bool                   `json:"-"`
	IsNew        bool                   `json:"-"`
	ShouldRotate bool                   `json:"-"`
	CanSave      bool                   `json:"-"`
	CreatedAt    time.Time              `json:"created_at"`
	LastAccessed time.Time              `json:"last_accessed"`
	
	// Authentication data
	DBName   string `json:"db_name,omitempty"`
	UserID   int    `json:"user_id,omitempty"`
	Login    string `json:"login,omitempty"`
	
	// Context data
	Context map[string]interface{} `json:"context"`
	
	mu sync.RWMutex
}

// NewSession creates a new session
func NewSession(sid string) *Session {
	now := time.Now()
	return &Session{
		SID:          sid,
		Data:         make(map[string]interface{}),
		IsDirty:      false,
		IsNew:        true,
		ShouldRotate: false,
		CanSave:      true,
		CreatedAt:    now,
		LastAccessed: now,
		Context:      getDefaultContext(),
	}
}

// Get retrieves a value from the session
func (s *Session) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Handle special keys
	switch key {
	case "db_name", "db":
		return s.DBName, s.DBName != ""
	case "user_id", "uid":
		return s.UserID, s.UserID != 0
	case "login":
		return s.Login, s.Login != ""
	}
	
	value, exists := s.Data[key]
	return value, exists
}

// Set stores a value in the session
func (s *Session) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Handle special keys
	switch key {
	case "db_name", "db":
		if str, ok := value.(string); ok {
			s.DBName = str
			s.IsDirty = true
		}
		return
	case "user_id", "uid":
		if id, ok := value.(int); ok {
			s.UserID = id
			s.IsDirty = true
		}
		return
	case "login":
		if str, ok := value.(string); ok {
			s.Login = str
			s.IsDirty = true
		}
		return
	}
	
	// Check if value actually changed
	if existing, exists := s.Data[key]; !exists || !deepEqual(existing, value) {
		s.IsDirty = true
	}
	
	s.Data[key] = value
}

// Delete removes a value from the session
func (s *Session) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Handle special keys
	switch key {
	case "db_name", "db":
		s.DBName = ""
		s.IsDirty = true
		return
	case "user_id", "uid":
		s.UserID = 0
		s.IsDirty = true
		return
	case "login":
		s.Login = ""
		s.IsDirty = true
		return
	}
	
	if _, exists := s.Data[key]; exists {
		delete(s.Data, key)
		s.IsDirty = true
	}
}

// Clear removes all data from the session
func (s *Session) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.Data = make(map[string]interface{})
	s.DBName = ""
	s.UserID = 0
	s.Login = ""
	s.Context = getDefaultContext()
	s.IsDirty = true
}

// Authenticate stores authentication information in the session
func (s *Session) Authenticate(dbname, login string, userID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.DBName = dbname
	s.Login = login
	s.UserID = userID
	s.IsDirty = true
	
	// Store in context as well
	s.Context["db_name"] = dbname
	s.Context["user_id"] = userID
	s.Context["login"] = login
}

// Logout clears authentication information
func (s *Session) Logout(keepDB bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !keepDB {
		s.DBName = ""
		delete(s.Context, "db_name")
	}
	
	s.UserID = 0
	s.Login = ""
	s.IsDirty = true
	
	delete(s.Context, "user_id")
	delete(s.Context, "login")
}

// IsAuthenticated checks if the session has valid authentication
func (s *Session) IsAuthenticated() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.UserID != 0 && s.Login != ""
}

// Touch updates the last accessed time
func (s *Session) Touch() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.LastAccessed = time.Now()
	s.IsDirty = true
}

// UpdateContext updates the session context
func (s *Session) UpdateContext(updates map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	for key, value := range updates {
		s.Context[key] = value
	}
	s.IsDirty = true
}

// GetContext returns a copy of the session context
func (s *Session) GetContext() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	context := make(map[string]interface{})
	for key, value := range s.Context {
		context[key] = value
	}
	return context
}

// FilesystemSessionStore implements SessionStore using filesystem (like Odoo's FilesystemSessionStore)
type FilesystemSessionStore struct {
	path         string
	renewMissing bool
	mu           sync.RWMutex
}

// NewFilesystemSessionStore creates a new filesystem session store
func NewFilesystemSessionStore(path string, renewMissing bool) (*FilesystemSessionStore, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}
	
	return &FilesystemSessionStore{
		path:         path,
		renewMissing: renewMissing,
	}, nil
}

// New creates a new session with a generated SID
func (fs *FilesystemSessionStore) New() *Session {
	sid := generateSessionID()
	return NewSession(sid)
}

// Get retrieves a session by SID
func (fs *FilesystemSessionStore) Get(sid string) *Session {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	
	if !fs.IsValidKey(sid) {
		if fs.renewMissing {
			return fs.New()
		}
		return nil
	}
	
	sessionFile := filepath.Join(fs.path, sid+".json")
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		if fs.renewMissing {
			return fs.New()
		}
		return nil
	}
	
	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		if fs.renewMissing {
			return fs.New()
		}
		return nil
	}
	
	session.IsNew = false
	session.IsDirty = false
	session.CanSave = true
	session.Touch()
	
	return &session
}

// Save persists a session to disk
func (fs *FilesystemSessionStore) Save(session *Session) error {
	if !session.CanSave || !session.IsDirty {
		return nil
	}
	
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	sessionFile := filepath.Join(fs.path, session.SID+".json")
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}
	
	if err := os.WriteFile(sessionFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}
	
	session.IsDirty = false
	session.IsNew = false
	
	return nil
}

// Delete removes a session from storage
func (fs *FilesystemSessionStore) Delete(sid string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	sessionFile := filepath.Join(fs.path, sid+".json")
	return os.Remove(sessionFile)
}

// IsValidKey checks if a session ID is valid
func (fs *FilesystemSessionStore) IsValidKey(sid string) bool {
	if len(sid) != 64 { // 32 bytes = 64 hex chars
		return false
	}
	
	// Check if file exists
	sessionFile := filepath.Join(fs.path, sid+".json")
	_, err := os.Stat(sessionFile)
	return err == nil
}

// Cleanup removes expired sessions
func (fs *FilesystemSessionStore) Cleanup() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	maxAge := 24 * time.Hour // Sessions expire after 24 hours
	cutoff := time.Now().Add(-maxAge)
	
	return filepath.Walk(fs.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		
		if info.ModTime().Before(cutoff) {
			return os.Remove(path)
		}
		
		return nil
	})
}

// Helper functions

// generateSessionID creates a new random session ID
func generateSessionID() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		panic(fmt.Sprintf("failed to generate session ID: %v", err))
	}
	return hex.EncodeToString(bytes)
}

// getDefaultContext returns default session context
func getDefaultContext() map[string]interface{} {
	return map[string]interface{}{
		"lang":     "en_US",
		"tz":       "UTC",
		"timezone": "UTC",
	}
}

// deepEqual compares two values for deep equality
func deepEqual(a, b interface{}) bool {
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return string(aJSON) == string(bJSON)
}