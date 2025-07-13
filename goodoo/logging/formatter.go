package logging

import (
	"context"
	"fmt"
	"os"
	"time"
)

// LogLevel represents different log levels
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
	CRITICAL
)

// String returns the string representation of log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case CRITICAL:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// LogRecord represents a log entry similar to Python's LogRecord
type LogRecord struct {
	Timestamp time.Time
	Level     LogLevel
	Message   string
	Logger    string
	Pathname  string
	LineNo    int
	FuncName  string
	PID       int
	DBName    string
	PerfInfo  string
	Metadata  map[string]interface{}
}

// Formatter interface for log formatting
type Formatter interface {
	Format(record *LogRecord) string
}

// DBFormatter formats logs with database context (like Python's DBFormatter)
type DBFormatter struct {
	format string
}

// NewDBFormatter creates a new database formatter
func NewDBFormatter() *DBFormatter {
	return &DBFormatter{
		format: "%(asctime)s %(pid)s %(levelname)s %(dbname)s %(name)s: %(message)s %(perf_info)s",
	}
}

// Format formats a log record
func (f *DBFormatter) Format(record *LogRecord) string {
	// Extract database name from context if available
	dbname := record.DBName
	if dbname == "" {
		dbname = "?"
	}

	perfInfo := record.PerfInfo
	if perfInfo == "" {
		perfInfo = "- - -"
	}

	return fmt.Sprintf("%s %d %s %s %s: %s %s",
		record.Timestamp.Format("2006-01-02 15:04:05,000"),
		record.PID,
		record.Level.String(),
		dbname,
		record.Logger,
		record.Message,
		perfInfo,
	)
}

// ColoredFormatter extends DBFormatter with color support
type ColoredFormatter struct {
	*DBFormatter
}

// NewColoredFormatter creates a new colored formatter
func NewColoredFormatter() *ColoredFormatter {
	return &ColoredFormatter{
		DBFormatter: NewDBFormatter(),
	}
}

// Format formats a log record with colors
func (f *ColoredFormatter) Format(record *LogRecord) string {
	// Clone record to avoid modifying original
	coloredRecord := *record
	coloredRecord.Level = LogLevel(int(record.Level)) // Ensure proper type

	// Get level string and colorize it
	levelStr := ColorizeLevel(record.Level.String())

	dbname := record.DBName
	if dbname == "" {
		dbname = "?"
	}

	perfInfo := record.PerfInfo
	if perfInfo == "" {
		perfInfo = "- - -"
	}

	return fmt.Sprintf("%s %d %s %s %s: %s %s",
		record.Timestamp.Format("2006-01-02 15:04:05,000"),
		record.PID,
		levelStr,
		dbname,
		record.Logger,
		record.Message,
		perfInfo,
	)
}

// ContextHelper extracts database name and other context from Go context
func ContextHelper(ctx context.Context) (dbname string, metadata map[string]interface{}) {
	if ctx == nil {
		return "", nil
	}

	metadata = make(map[string]interface{})

	// Extract database name from context
	if db := ctx.Value("dbname"); db != nil {
		if dbStr, ok := db.(string); ok {
			dbname = dbStr
		}
	}

	// Extract other metadata
	if reqID := ctx.Value("request_id"); reqID != nil {
		metadata["request_id"] = reqID
	}

	if userID := ctx.Value("user_id"); userID != nil {
		metadata["user_id"] = userID
	}

	return dbname, metadata
}

// CreateLogRecord creates a new log record
func CreateLogRecord(level LogLevel, logger, message, pathname string, lineno int, funcname string, ctx context.Context) *LogRecord {
	dbname, metadata := ContextHelper(ctx)

	return &LogRecord{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Logger:    logger,
		Pathname:  pathname,
		LineNo:    lineno,
		FuncName:  funcname,
		PID:       os.Getpid(),
		DBName:    dbname,
		PerfInfo:  "", // Will be filled by performance filter
		Metadata:  metadata,
	}
}
