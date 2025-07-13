package logging

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"
)

// Handler interface for log handlers
type Handler interface {
	Emit(record *LogRecord) error
	Close() error
}

// StreamHandler writes logs to an io.Writer (like Python's StreamHandler)
type StreamHandler struct {
	writer    io.Writer
	formatter Formatter
	mu        sync.Mutex
}

// NewStreamHandler creates a new stream handler
func NewStreamHandler(writer io.Writer, formatter Formatter) *StreamHandler {
	if formatter == nil {
		if IsColorTerminal() {
			formatter = NewColoredFormatter()
		} else {
			formatter = NewDBFormatter()
		}
	}

	return &StreamHandler{
		writer:    writer,
		formatter: formatter,
	}
}

// Emit writes a log record to the stream
func (h *StreamHandler) Emit(record *LogRecord) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	formatted := h.formatter.Format(record)
	_, err := fmt.Fprintln(h.writer, formatted)
	return err
}

// Close closes the handler
func (h *StreamHandler) Close() error {
	if closer, ok := h.writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// FileHandler writes logs to a file (like Python's FileHandler)
type FileHandler struct {
	*StreamHandler
	file *os.File
}

// NewFileHandler creates a new file handler
func NewFileHandler(filename string, formatter Formatter) (*FileHandler, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	if formatter == nil {
		formatter = NewDBFormatter() // No colors for file output
	}

	return &FileHandler{
		StreamHandler: NewStreamHandler(file, formatter),
		file:          file,
	}, nil
}

// Close closes the file handler
func (h *FileHandler) Close() error {
	return h.file.Close()
}

// PostgreSQLHandler writes logs to PostgreSQL database (like Python's PostgreSQLHandler)
type PostgreSQLHandler struct {
	db              *sql.DB
	dbName          string
	supportMetadata bool
	mu              sync.Mutex
}

// NewPostgreSQLHandler creates a new PostgreSQL handler
func NewPostgreSQLHandler(dbConnStr, dbName string) (*PostgreSQLHandler, error) {
	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		return nil, err
	}

	handler := &PostgreSQLHandler{
		db:     db,
		dbName: dbName,
	}

	// Check if metadata column exists
	err = handler.checkMetadataSupport()
	if err != nil {
		// Log error but continue without metadata support
		slog.Warn("Failed to check metadata support", "error", err)
	}

	return handler, nil
}

// checkMetadataSupport checks if the ir_logging table supports metadata
func (h *PostgreSQLHandler) checkMetadataSupport() error {
	query := `SELECT 1 FROM information_schema.columns 
			  WHERE table_name='ir_logging' AND column_name='metadata'`

	var exists int
	err := h.db.QueryRow(query).Scan(&exists)
	if err == nil {
		h.supportMetadata = true
	} else if err == sql.ErrNoRows {
		h.supportMetadata = false
		err = nil
	}

	return err
}

// Emit writes a log record to PostgreSQL
func (h *PostgreSQLHandler) Emit(record *LogRecord) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Set statement timeout to prevent deadlocks
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := h.db.ExecContext(ctx, "SET LOCAL statement_timeout = 1000")
	if err != nil {
		return err
	}

	dbname := h.dbName
	if record.DBName != "" {
		dbname = record.DBName
	}

	if h.supportMetadata && len(record.Metadata) > 0 {
		metadataJSON, err := json.Marshal(record.Metadata)
		if err != nil {
			return err
		}

		query := `INSERT INTO ir_logging(create_date, type, dbname, name, level, message, path, line, func, metadata)
				  VALUES (NOW() at time zone 'UTC', $1, $2, $3, $4, $5, $6, $7, $8, $9)`

		_, err = h.db.ExecContext(ctx, query,
			"server", dbname, record.Logger, record.Level.String(),
			record.Message, record.Pathname, record.LineNo, record.FuncName,
			string(metadataJSON))
		return err
	}

	// Insert without metadata
	query := `INSERT INTO ir_logging(create_date, type, dbname, name, level, message, path, line, func)
			  VALUES (NOW() at time zone 'UTC', $1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = h.db.ExecContext(ctx, query,
		"server", dbname, record.Logger, record.Level.String(),
		record.Message, record.Pathname, record.LineNo, record.FuncName)

	return err
}

// Close closes the PostgreSQL handler
func (h *PostgreSQLHandler) Close() error {
	return h.db.Close()
}

// SyslogHandler handles syslog output (simplified version)
type SyslogHandler struct {
	*StreamHandler
}

// NewSyslogHandler creates a new syslog handler
func NewSyslogHandler() *SyslogHandler {
	// For simplicity, we'll write to stderr
	// In a real implementation, you'd use the syslog package
	formatter := &DBFormatter{} // Simple format for syslog

	return &SyslogHandler{
		StreamHandler: NewStreamHandler(os.Stderr, formatter),
	}
}
