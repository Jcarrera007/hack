package logging

import (
	"fmt"
	"os"
)

// Color constants based on netsvc.py
const (
	Black = iota
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
	Nothing
	Default
)

// ANSI escape sequences
const (
	ResetSeq = "\033[0m"
	ColorSeq = "\033[1;%dm"
	BoldSeq  = "\033[1m"
)

// ColorPattern creates colored text
var ColorPattern = fmt.Sprintf("%s%s%%s%s", ColorSeq, ColorSeq, ResetSeq)

// LevelColorMapping maps log levels to colors
var LevelColorMapping = map[string][2]int{
	"DEBUG":    {Blue, Default},
	"INFO":     {Green, Default},
	"WARNING":  {Yellow, Default},
	"ERROR":    {Red, Default},
	"CRITICAL": {White, Red},
}

// IsColorTerminal checks if we should output colors
func IsColorTerminal() bool {
	if os.Getenv("GOODOO_COLORS") != "" {
		return true
	}
	
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return true
	}
	
	return false
}

// ColorizeLevel colors a log level string
func ColorizeLevel(level string) string {
	if !IsColorTerminal() {
		return level
	}
	
	colors, exists := LevelColorMapping[level]
	if !exists {
		return level
	}
	
	fg, bg := colors[0], colors[1]
	return fmt.Sprintf(ColorPattern, 30+fg, 40+bg, level)
}

// ColorizeTime colors time values based on thresholds
func ColorizeTime(value float64, format string, low, high float64) string {
	if !IsColorTerminal() {
		return fmt.Sprintf(format, value)
	}
	
	if value > high {
		return fmt.Sprintf(ColorPattern, 30+Red, 40+Default, fmt.Sprintf(format, value))
	}
	if value > low {
		return fmt.Sprintf(ColorPattern, 30+Yellow, 40+Default, fmt.Sprintf(format, value))
	}
	return fmt.Sprintf(format, value)
}