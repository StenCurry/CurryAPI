package services

import (
	"fmt"
	"sync"
	"time"

	"Curry2API-go/database"
	"github.com/sirupsen/logrus"
)

// UsageTrackerError represents an error from the usage tracker
type UsageTrackerError struct {
	Message string
}

func (e *UsageTrackerError) Error() string {
	return fmt.Sprintf("usage tracker error: %s", e.Message)
}

// UsageTrackerConfig holds configuration for the usage tracker
type UsageTrackerConfig struct {
	Enabled        bool          // Enable/disable usage tracking
	ChannelSize    int           // Size of the buffered channel
	BatchSize      int           // Number of records to batch before writing
	FlushInterval  time.Duration // How often to flush batches
	MaxRetries     int           // Maximum number of retry attempts
	RetryBackoffMs int           // Initial backoff for retries (ms)
}

// UsageRecord represents a single API usage event
type UsageRecord struct {
	UserID           int64
	Username         string
	APIToken         string
	TokenName        string
	Model            string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	CursorSession    string
	StatusCode       int
	ErrorMessage     string
	RequestTime      time.Time
	ResponseTime     time.Time
	Duration         time.Duration
}

// UsageTracker manages asynchronous usage tracking
type UsageTracker struct {
	config      *UsageTrackerConfig
	recordChan  chan *UsageRecord
	stopChan    chan struct{}
	wg          sync.WaitGroup
	mu          sync.RWMutex
	initialized bool
}

var (
	instance *UsageTracker
	once     sync.Once
)

// NewUsageTracker creates a new UsageTracker instance
func NewUsageTracker(config *UsageTrackerConfig) *UsageTracker {
	if config == nil {
		config = &UsageTrackerConfig{
			Enabled:        true,
			ChannelSize:    1000,
			BatchSize:      100,
			FlushInterval:  5 * time.Second,
			MaxRetries:     3,
			RetryBackoffMs: 100,
		}
	}

	tracker := &UsageTracker{
		config:      config,
		recordChan:  make(chan *UsageRecord, config.ChannelSize),
		stopChan:    make(chan struct{}),
		initialized: true,
	}

	// Start background worker if enabled
	if config.Enabled {
		tracker.wg.Add(1)
		go tracker.processRecords()
		logrus.Info("Usage tracker started")
	} else {
		logrus.Info("Usage tracking is disabled")
	}

	return tracker
}

// GetUsageTracker returns the singleton instance of UsageTracker
func GetUsageTracker() *UsageTracker {
	once.Do(func() {
		// Initialize with default config
		instance = NewUsageTracker(nil)
	})
	return instance
}

// InitUsageTracker initializes the singleton with a specific config
func InitUsageTracker(config *UsageTrackerConfig) {
	once.Do(func() {
		instance = NewUsageTracker(config)
	})
}

// IsEnabled returns whether usage tracking is enabled
func (ut *UsageTracker) IsEnabled() bool {
	ut.mu.RLock()
	defer ut.mu.RUnlock()
	return ut.config.Enabled
}

// TrackUsage records a usage event asynchronously (non-blocking)
func (ut *UsageTracker) TrackUsage(record *UsageRecord) error {
	// Skip if tracking is disabled
	if !ut.IsEnabled() {
		return nil
	}

	// Non-blocking send to channel
	select {
	case ut.recordChan <- record:
		return nil
	default:
		// Channel is full, log and drop the record to prevent blocking
		logrus.Warn("Usage tracking channel full, dropping record")
		return ErrChannelFull
	}
}

// processRecords is the background worker that processes usage records
func (ut *UsageTracker) processRecords() {
	defer ut.wg.Done()

	batch := make([]*UsageRecord, 0, ut.config.BatchSize)
	ticker := time.NewTicker(ut.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case record := <-ut.recordChan:
			// Add record to batch
			batch = append(batch, record)

			// Flush if batch is full
			if len(batch) >= ut.config.BatchSize {
				ut.flushBatch(batch)
				batch = batch[:0] // Reset batch
			}

		case <-ticker.C:
			// Periodic flush
			if len(batch) > 0 {
				ut.flushBatch(batch)
				batch = batch[:0] // Reset batch
			}

		case <-ut.stopChan:
			// Graceful shutdown: flush remaining records
			if len(batch) > 0 {
				logrus.Infof("Flushing %d remaining records before shutdown", len(batch))
				ut.flushBatch(batch)
			}
			return
		}
	}
}

// flushBatch writes a batch of usage records to the database with retry logic
func (ut *UsageTracker) flushBatch(batch []*UsageRecord) {
	if len(batch) == 0 {
		return
	}

	startTime := time.Now()
	
	// Convert service records to database records
	dbRecords := make([]*database.UsageRecord, len(batch))
	for i, record := range batch {
		dbRecords[i] = &database.UsageRecord{
			UserID:           record.UserID,
			Username:         record.Username,
			APIToken:         record.APIToken,
			TokenName:        record.TokenName,
			Model:            record.Model,
			PromptTokens:     record.PromptTokens,
			CompletionTokens: record.CompletionTokens,
			TotalTokens:      record.TotalTokens,
			CursorSession:    record.CursorSession,
			StatusCode:       record.StatusCode,
			ErrorMessage:     record.ErrorMessage,
			RequestTime:      record.RequestTime,
			ResponseTime:     record.ResponseTime,
			DurationMs:       int(record.Duration.Milliseconds()),
		}
	}

	// Retry logic with exponential backoff
	var lastErr error
	for attempt := 0; attempt < ut.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate backoff duration
			backoff := time.Duration(ut.config.RetryBackoffMs) * time.Millisecond * time.Duration(1<<uint(attempt-1))
			logrus.Infof("Retrying batch write (attempt %d/%d) after %v", attempt+1, ut.config.MaxRetries, backoff)
			time.Sleep(backoff)
		}

		err := database.BatchInsertUsageRecords(dbRecords)
		if err == nil {
			// Success
			duration := time.Since(startTime)
			logrus.Infof("Successfully flushed batch of %d records in %v", len(batch), duration)
			return
		}

		lastErr = err
		logrus.Warnf("Failed to flush batch (attempt %d/%d): %v", attempt+1, ut.config.MaxRetries, err)
	}

	// All retries failed
	logrus.Errorf("Failed to flush batch after %d attempts: %v", ut.config.MaxRetries, lastErr)
	logrus.Errorf("Lost %d usage records - manual recovery may be required", len(batch))
}

// Shutdown gracefully shuts down the usage tracker
func (ut *UsageTracker) Shutdown() {
	if !ut.initialized {
		return
	}

	logrus.Info("Shutting down usage tracker...")
	close(ut.stopChan)
	ut.wg.Wait()
	logrus.Info("Usage tracker shut down complete")
}

// ErrChannelFull is returned when the tracking channel is full
var ErrChannelFull = &UsageTrackerError{Message: "tracking channel is full"}
