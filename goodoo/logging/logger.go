package logging

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
)

// Logger represents a logger instance (similar to Python's Logger)
type Logger struct {
	name     string
	handlers []Handler
	levels   LoggerLevels
	mu       sync.RWMutex
}

// Global logger registry
var (
	loggers     = make(map[string]*Logger)
	loggersMu   = sync.RWMutex{}
	rootLogger  *Logger
	initialized = false
)

// GetLogger returns a logger with the given name
func GetLogger(name string) *Logger {
	loggersMu.RLock()
	if logger, exists := loggers[name]; exists {
		loggersMu.RUnlock()
		return logger
	}
	loggersMu.RUnlock()
	
	loggersMu.Lock()
	defer loggersMu.Unlock()
	
	// Double-check after acquiring write lock
	if logger, exists := loggers[name]; exists {
		return logger
	}
	
	logger := &Logger{
		name:     name,
		handlers: []Handler{},
		levels:   make(LoggerLevels),
	}
	
	// Inherit handlers and levels from root logger if it exists
	if rootLogger != nil {
		logger.handlers = make([]Handler, len(rootLogger.handlers))
		copy(logger.handlers, rootLogger.handlers)
		logger.levels = rootLogger.levels
	}
	
	loggers[name] = logger
	return logger
}

// InitLogger initializes the logging system (similar to netsvc.init_logger)
func InitLogger() error {
	if initialized {
		return nil
	}
	initialized = true
	
	config := DefaultLogConfig()
	
	// Create root logger
	rootLogger = &Logger{
		name:     "",
		handlers: []Handler{},
		levels:   config.BuildLoggerLevels(),
	}
	
	// Add stream handler (console)
	var streamHandler Handler
	if config.SysLog {
		streamHandler = NewSyslogHandler()
	} else if config.LogFile != "" {
		fileHandler, err := NewFileHandler(config.LogFile, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: couldn't create the logfile. Logging to console: %v\n", err)
			streamHandler = NewStreamHandler(os.Stderr, nil)
		} else {
			streamHandler = fileHandler
		}
	} else {
		streamHandler = NewStreamHandler(os.Stderr, nil)
	}
	
	rootLogger.AddHandler(streamHandler)
	
	// Add PostgreSQL handler if configured
	if config.LogDB != "" {
		// Note: You'll need to provide the connection string
		// This is a placeholder - in real usage, you'd get this from config
		connStr := fmt.Sprintf("host=localhost dbname=%s sslmode=disable", config.LogDB)
		pgHandler, err := NewPostgreSQLHandler(connStr, config.LogDB)
		if err != nil {
			// Log error but continue
			rootLogger.Error("Failed to create PostgreSQL handler: %v", err)
		} else {
			rootLogger.AddHandler(pgHandler)
		}
	}
	
	// Store root logger in registry
	loggers[""] = rootLogger
	
	return nil
}

// AddHandler adds a handler to the logger
func (l *Logger) AddHandler(handler Handler) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.handlers = append(l.handlers, handler)
}

// RemoveHandler removes a handler from the logger
func (l *Logger) RemoveHandler(handler Handler) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	for i, h := range l.handlers {
		if h == handler {
			l.handlers = append(l.handlers[:i], l.handlers[i+1:]...)
			break
		}
	}
}

// SetLevel sets the level for this logger
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.levels == nil {
		l.levels = make(LoggerLevels)
	}
	l.levels[l.name] = level
}

// log is the internal logging method
func (l *Logger) log(level LogLevel, ctx context.Context, format string, args ...interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	// Check if we should log this message
	if !l.levels.ShouldLog(l.name, level) {
		return
	}
	
	// Get caller information
	_, file, line, ok := runtime.Caller(3) // Skip log, Debug/Info/etc, and user function
	funcName := "unknown"
	if ok {
		if pc, _, _, ok := runtime.Caller(3); ok {
			if fn := runtime.FuncForPC(pc); fn != nil {
				funcName = fn.Name()
			}
		}
	} else {
		file = "unknown"
		line = 0
	}
	
	// Format message
	message := fmt.Sprintf(format, args...)
	
	// Create log record
	record := CreateLogRecord(level, l.name, message, file, line, funcName, ctx)
	
	// Add performance info if available
	if ctx != nil {
		filter := NewPerfFilter(IsColorTerminal())
		filter.Filter(record, ctx)
	}
	
	// Emit to all handlers
	for _, handler := range l.handlers {
		if err := handler.Emit(record); err != nil {
			// If we can't log the error, write to stderr as last resort
			fmt.Fprintf(os.Stderr, "Logging error: %v\n", err)
		}
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, nil, format, args...)
}

// DebugCtx logs a debug message with context
func (l *Logger) DebugCtx(ctx context.Context, format string, args ...interface{}) {
	l.log(DEBUG, ctx, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, nil, format, args...)
}

// InfoCtx logs an info message with context
func (l *Logger) InfoCtx(ctx context.Context, format string, args ...interface{}) {
	l.log(INFO, ctx, format, args...)
}

// Warning logs a warning message
func (l *Logger) Warning(format string, args ...interface{}) {
	l.log(WARNING, nil, format, args...)
}

// WarningCtx logs a warning message with context
func (l *Logger) WarningCtx(ctx context.Context, format string, args ...interface{}) {
	l.log(WARNING, ctx, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, nil, format, args...)
}

// ErrorCtx logs an error message with context
func (l *Logger) ErrorCtx(ctx context.Context, format string, args ...interface{}) {
	l.log(ERROR, ctx, format, args...)
}

// Critical logs a critical message
func (l *Logger) Critical(format string, args ...interface{}) {
	l.log(CRITICAL, nil, format, args...)
}

// CriticalCtx logs a critical message with context
func (l *Logger) CriticalCtx(ctx context.Context, format string, args ...interface{}) {
	l.log(CRITICAL, ctx, format, args...)
}

// Close closes all handlers
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	var lastErr error
	for _, handler := range l.handlers {
		if err := handler.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// Convenience functions for package-level logging
var packageLogger = GetLogger("goodoo")

// Debug logs a debug message using the package logger
func Debug(format string, args ...interface{}) {
	packageLogger.Debug(format, args...)
}

// Info logs an info message using the package logger
func Info(format string, args ...interface{}) {
	packageLogger.Info(format, args...)
}

// Warning logs a warning message using the package logger
func Warning(format string, args ...interface{}) {
	packageLogger.Warning(format, args...)
}

// Error logs an error message using the package logger
func Error(format string, args ...interface{}) {
	packageLogger.Error(format, args...)
}

// Critical logs a critical message using the package logger
func Critical(format string, args ...interface{}) {
	packageLogger.Critical(format, args...)
}