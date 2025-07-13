package logging

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// PerfContext holds performance metrics for a request
type PerfContext struct {
	StartTime  time.Time
	QueryCount int
	QueryTime  time.Duration
	mu         sync.Mutex
}

// NewPerfContext creates a new performance context
func NewPerfContext() *PerfContext {
	return &PerfContext{
		StartTime: time.Now(),
	}
}

// AddQuery records a database query
func (pc *PerfContext) AddQuery(duration time.Duration) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.QueryCount++
	pc.QueryTime += duration
}

// GetMetrics returns formatted performance metrics
func (pc *PerfContext) GetMetrics() string {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	elapsed := time.Since(pc.StartTime)
	remainingTime := elapsed - pc.QueryTime

	return fmt.Sprintf("%d %.3f %.3f",
		pc.QueryCount,
		pc.QueryTime.Seconds(),
		remainingTime.Seconds(),
	)
}

// GetColoredMetrics returns colored performance metrics
func (pc *PerfContext) GetColoredMetrics() string {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	elapsed := time.Since(pc.StartTime)
	remainingTime := elapsed - pc.QueryTime

	return fmt.Sprintf("%s %s %s",
		ColorizeTime(float64(pc.QueryCount), "%d", 100, 1000),
		ColorizeTime(pc.QueryTime.Seconds(), "%.3f", 0.1, 3),
		ColorizeTime(remainingTime.Seconds(), "%.3f", 1, 5),
	)
}

// PerfFilter adds performance information to log records (like Python's PerfFilter)
type PerfFilter struct {
	colored bool
}

// NewPerfFilter creates a new performance filter
func NewPerfFilter(colored bool) *PerfFilter {
	return &PerfFilter{colored: colored}
}

// Filter adds performance info to a log record
func (pf *PerfFilter) Filter(record *LogRecord, ctx context.Context) {
	if ctx == nil {
		record.PerfInfo = "- - -"
		return
	}

	perfCtx, ok := ctx.Value("perf_context").(*PerfContext)
	if !ok {
		record.PerfInfo = "- - -"
		return
	}

	if pf.colored {
		record.PerfInfo = perfCtx.GetColoredMetrics()
	} else {
		record.PerfInfo = perfCtx.GetMetrics()
	}
}

// PerformanceMiddleware is an Echo middleware that tracks performance metrics
func PerformanceMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Create performance context
			perfCtx := NewPerfContext()

			// Add to request context
			req := c.Request()
			ctx := context.WithValue(req.Context(), "perf_context", perfCtx)
			c.SetRequest(req.WithContext(ctx))

			// Continue with next handler
			err := next(c)

			// Create a log record for the request
			record := CreateLogRecord(
				INFO,
				"goodoo.http",
				fmt.Sprintf("%s %s - %d", c.Request().Method, c.Request().URL.Path, c.Response().Status),
				"",
				0,
				"",
				ctx,
			)

			// Add performance info
			filter := NewPerfFilter(IsColorTerminal())
			filter.Filter(record, ctx)

			// Log the record (would normally go through logger)
			fmt.Printf("Performance: %s - %s\n", record.Message, record.PerfInfo)

			return err
		}
	}
}

// DatabaseQueryWrapper wraps database queries to track performance
type DatabaseQueryWrapper struct {
	ctx context.Context
}

// NewDatabaseQueryWrapper creates a new query wrapper
func NewDatabaseQueryWrapper(ctx context.Context) *DatabaseQueryWrapper {
	return &DatabaseQueryWrapper{ctx: ctx}
}

// TrackQuery records the execution of a database query
func (dqw *DatabaseQueryWrapper) TrackQuery(fn func() error) error {
	start := time.Now()
	err := fn()
	duration := time.Since(start)

	// Add query to performance context
	if perfCtx, ok := dqw.ctx.Value("perf_context").(*PerfContext); ok {
		perfCtx.AddQuery(duration)
	}

	return err
}
