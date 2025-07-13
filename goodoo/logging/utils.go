package logging

import (
	"fmt"
	"runtime"
	"strings"
)

// LogMessage safely formats and logs a message with proper string handling
func LogMessage(logger *Logger, level LogLevel, format string, args ...interface{}) {
	// Convert all args to safe strings
	safeArgs := make([]interface{}, len(args))
	for i, arg := range args {
		safeArgs[i] = SafeString(arg)
	}
	
	// Format the message safely
	message := fmt.Sprintf(format, safeArgs...)
	
	// Validate and fix UTF-8 if needed
	if validatedMsg, isValid := ValidateUTF8(message, true); !isValid {
		logger.Warning("Log message contained invalid UTF-8, fixed automatically")
		message = validatedMsg
	}
	
	// Log using the appropriate level
	switch level {
	case DEBUG:
		logger.Debug("%s", message)
	case INFO:
		logger.Info("%s", message)
	case WARNING:
		logger.Warning("%s", message)
	case ERROR:
		logger.Error("%s", message)
	case CRITICAL:
		logger.Critical("%s", message)
	}
}

// GetCallerInfo returns information about the caller (inspired by Python's logging)
func GetCallerInfo(skip int) (file string, line int, funcName string) {
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "unknown", 0, "unknown"
	}
	
	if fn := runtime.FuncForPC(pc); fn != nil {
		funcName = fn.Name()
		// Simplify function name (remove package path)
		if lastSlash := strings.LastIndex(funcName, "/"); lastSlash >= 0 {
			funcName = funcName[lastSlash+1:]
		}
	} else {
		funcName = "unknown"
	}
	
	// Simplify file path
	if lastSlash := strings.LastIndex(file, "/"); lastSlash >= 0 {
		file = file[lastSlash+1:]
	}
	
	return file, line, funcName
}

// ConfigureFromEnvironment configures logging from environment variables
func ConfigureFromEnvironment() error {
	config := DefaultLogConfig()
	
	// Initialize the logger with the configuration
	if err := InitLogger(); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	
	// Apply log level configurations
	rootLogger := GetLogger("")
	levels := config.BuildLoggerLevels()
	rootLogger.levels = levels
	
	// Log the configuration for debugging
	logger := GetLogger("goodoo.logging")
	logger.Info("Logging system initialized")
	logger.Debug("Log level: %s", config.LogLevel)
	logger.Debug("Log file: %s", config.LogFile)
	logger.Debug("Log DB: %s", config.LogDB)
	logger.Debug("Syslog enabled: %v", config.SysLog)
	
	return nil
}

// LevelFromString provides a safe way to convert strings to log levels
func LevelFromString(levelStr string) (LogLevel, error) {
	if !IsValidLogLevel(levelStr) {
		return INFO, fmt.Errorf("invalid log level: %s", levelStr)
	}
	return ParseLogLevelString(levelStr), nil
}

// GetLevelHierarchy returns the log levels in order of severity
func GetLevelHierarchy() []LogLevel {
	return []LogLevel{DEBUG, INFO, WARNING, ERROR, CRITICAL}
}

// GetLevelName returns the string name for a log level
func GetLevelName(level LogLevel) string {
	return GetLogLevelMetadata(level).Name
}

// GetLevelDescription returns a description for a log level
func GetLevelDescription(level LogLevel) string {
	return GetLogLevelMetadata(level).Description
}

// FormatLogLevelForDisplay formats a log level for display with optional colors
func FormatLogLevelForDisplay(level LogLevel, useColor bool) string {
	metadata := GetLogLevelMetadata(level)
	name := metadata.Name
	
	if useColor && IsColorTerminal() {
		fg, bg := metadata.Color[0], metadata.Color[1]
		return fmt.Sprintf(ColorPattern, 30+fg, 40+bg, strings.ToUpper(name))
	}
	
	return strings.ToUpper(name)
}

// CompareLogLevelsByName compares two log levels by their string names
func CompareLogLevelsByName(level1, level2 string) int {
	l1 := ParseLogLevelString(level1)
	l2 := ParseLogLevelString(level2)
	
	n1 := GetLogLevelNumericValue(l1)
	n2 := GetLogLevelNumericValue(l2)
	
	if n1 < n2 {
		return -1
	} else if n1 > n2 {
		return 1
	}
	return 0
}

// LogLevelStats holds statistics about log level usage
type LogLevelStats struct {
	Level LogLevel
	Count int64
	Name  string
}

// LoggerStats tracks statistics for a logger
type LoggerStats struct {
	Name        string
	LevelCounts map[LogLevel]int64
	TotalLogs   int64
}

// NewLoggerStats creates a new logger stats tracker
func NewLoggerStats(name string) *LoggerStats {
	return &LoggerStats{
		Name:        name,
		LevelCounts: make(map[LogLevel]int64),
		TotalLogs:   0,
	}
}

// RecordLog records a log event for statistics
func (ls *LoggerStats) RecordLog(level LogLevel) {
	ls.LevelCounts[level]++
	ls.TotalLogs++
}

// GetStats returns formatted statistics
func (ls *LoggerStats) GetStats() []LogLevelStats {
	var stats []LogLevelStats
	for _, level := range GetLevelHierarchy() {
		count := ls.LevelCounts[level]
		if count > 0 {
			stats = append(stats, LogLevelStats{
				Level: level,
				Count: count,
				Name:  GetLevelName(level),
			})
		}
	}
	return stats
}

// PrintLoggerStats prints formatted statistics for a logger
func PrintLoggerStats(stats *LoggerStats) {
	fmt.Printf("Logger: %s (Total: %d)\n", stats.Name, stats.TotalLogs)
	for _, levelStats := range stats.GetStats() {
		percentage := float64(levelStats.Count) / float64(stats.TotalLogs) * 100
		fmt.Printf("  %s: %d (%.1f%%)\n", 
			FormatLogLevelForDisplay(levelStats.Level, true),
			levelStats.Count, 
			percentage)
	}
}