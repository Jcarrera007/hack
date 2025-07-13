package logging

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Log level constants based on loglevels.py
const (
	LOG_NOTSET   = "notset"
	LOG_DEBUG    = "debug"
	LOG_INFO     = "info"
	LOG_WARNING  = "warn"
	LOG_ERROR    = "error"
	LOG_CRITICAL = "critical"
)

// Additional log level constants to extend the existing system
const (
	NOTSET_LEVEL = iota * 10  // 0
	DEBUG_LEVEL                       // 10
	INFO_LEVEL                        // 20
	WARNING_LEVEL                     // 30
	ERROR_LEVEL                       // 40
	CRITICAL_LEVEL                    // 50
)

// GetNumericValue returns the numeric value of the log level
func GetLogLevelNumericValue(level LogLevel) int {
	switch level {
	case DEBUG:
		return DEBUG_LEVEL
	case INFO:
		return INFO_LEVEL
	case WARNING:
		return WARNING_LEVEL
	case ERROR:
		return ERROR_LEVEL
	case CRITICAL:
		return CRITICAL_LEVEL
	default:
		return INFO_LEVEL
	}
}

// ParseLogLevelString converts string to LogLevel (case-insensitive)
func ParseLogLevelString(level string) LogLevel {
	normalized := strings.ToLower(strings.TrimSpace(level))
	switch normalized {
	case LOG_NOTSET, "":
		return INFO // NOTSET defaults to INFO in our system
	case LOG_DEBUG:
		return DEBUG
	case LOG_INFO:
		return INFO
	case LOG_WARNING, "warning":
		return WARNING
	case LOG_ERROR:
		return ERROR
	case LOG_CRITICAL, "crit":
		return CRITICAL
	default:
		return INFO // Default fallback
	}
}

// IsValidLogLevel checks if a string represents a valid log level
func IsValidLogLevel(level string) bool {
	normalizedLevel := strings.ToLower(strings.TrimSpace(level))
	validLevels := []string{
		LOG_NOTSET, "notset",
		LOG_DEBUG, "debug", 
		LOG_INFO, "info",
		LOG_WARNING, "warn", "warning",
		LOG_ERROR, "error",
		LOG_CRITICAL, "critical", "crit",
	}
	
	for _, valid := range validLevels {
		if normalizedLevel == valid {
			return true
		}
	}
	return false
}

// GetAllLogLevels returns all available log levels
func GetAllLogLevels() []LogLevel {
	return []LogLevel{DEBUG, INFO, WARNING, ERROR, CRITICAL}
}

// GetLogLevelNames returns all available log level names
func GetLogLevelNames() []string {
	return []string{LOG_NOTSET, LOG_DEBUG, LOG_INFO, LOG_WARNING, LOG_ERROR, LOG_CRITICAL}
}

// CompareLogLevels returns true if level1 is >= level2 (should be logged)
func CompareLogLevels(level1, level2 LogLevel) bool {
	return GetLogLevelNumericValue(level1) >= GetLogLevelNumericValue(level2)
}

// GetEffectiveLevel returns the effective log level, handling inheritance
func GetEffectiveLevel(level LogLevel, parentLevel LogLevel) LogLevel {
	// In our system, we don't have NOTSET, so just return the level
	return level
}

// LogLevelRange represents a range of log levels
type LogLevelRange struct {
	Min LogLevel
	Max LogLevel
}

// Contains checks if a level falls within the range
func (r LogLevelRange) Contains(level LogLevel) bool {
	return GetLogLevelNumericValue(level) >= GetLogLevelNumericValue(r.Min) && 
		   GetLogLevelNumericValue(level) <= GetLogLevelNumericValue(r.Max)
}

// NewLogLevelRange creates a new log level range
func NewLogLevelRange(min, max LogLevel) LogLevelRange {
	return LogLevelRange{Min: min, Max: max}
}

// String utility functions inspired by loglevels.py ustr functionality

// SafeString converts any value to a string safely (similar to ustr)
func SafeString(value interface{}) string {
	if value == nil {
		return "<nil>"
	}
	
	switch v := value.(type) {
	case string:
		if utf8.ValidString(v) {
			return v
		}
		// Try to fix invalid UTF-8
		return strings.ToValidUTF8(v, "�")
	case []byte:
		if utf8.Valid(v) {
			return string(v)
		}
		// Try to fix invalid UTF-8
		return string([]rune(string(v)))
	case error:
		return ExceptionToString(v)
	case fmt.Stringer:
		return SafeString(v.String())
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ExceptionToString converts an error to a string (similar to exception_to_unicode)
func ExceptionToString(err error) string {
	if err == nil {
		return ""
	}
	
	// Handle wrapped errors
	if unwrapped := Unwrap(err); unwrapped != nil {
		return fmt.Sprintf("%v: %v", err, ExceptionToString(unwrapped))
	}
	
	return err.Error()
}

// Unwrap safely unwraps an error if it supports unwrapping
func Unwrap(err error) error {
	if unwrapper, ok := err.(interface{ Unwrap() error }); ok {
		return unwrapper.Unwrap()
	}
	return nil
}

// GetPreferredEncoding returns the preferred encoding (always UTF-8 in Go)
func GetPreferredEncoding() string {
	return "utf-8"
}

// ValidateUTF8 checks if a string is valid UTF-8 and optionally fixes it
func ValidateUTF8(s string, fix bool) (string, bool) {
	if utf8.ValidString(s) {
		return s, true
	}
	
	if fix {
		return strings.ToValidUTF8(s, "�"), false
	}
	
	return s, false
}

// LogLevelMetadata holds additional information about log levels
type LogLevelMetadata struct {
	Name        string
	Description string
	Color       [2]int // [foreground, background] color codes
	Severity    int    // Higher number = more severe
}

// GetLogLevelMetadata returns metadata for a log level
func GetLogLevelMetadata(level LogLevel) LogLevelMetadata {
	switch level {
	case DEBUG:
		return LogLevelMetadata{
			Name:        LOG_DEBUG,
			Description: "Detailed information for debugging",
			Color:       [2]int{Blue, Default},
			Severity:    1,
		}
	case INFO:
		return LogLevelMetadata{
			Name:        LOG_INFO,
			Description: "General information about program execution",
			Color:       [2]int{Green, Default},
			Severity:    2,
		}
	case WARNING:
		return LogLevelMetadata{
			Name:        LOG_WARNING,
			Description: "Warning about potential issues",
			Color:       [2]int{Yellow, Default},
			Severity:    3,
		}
	case ERROR:
		return LogLevelMetadata{
			Name:        LOG_ERROR,
			Description: "Error conditions that need attention",
			Color:       [2]int{Red, Default},
			Severity:    4,
		}
	case CRITICAL:
		return LogLevelMetadata{
			Name:        LOG_CRITICAL,
			Description: "Critical errors that may cause program termination",
			Color:       [2]int{White, Red},
			Severity:    5,
		}
	default:
		return LogLevelMetadata{
			Name:        fmt.Sprintf("LEVEL_%d", GetLogLevelNumericValue(level)),
			Description: "Custom log level",
			Color:       [2]int{Default, Default},
			Severity:    GetLogLevelNumericValue(level) / 10,
		}
	}
}