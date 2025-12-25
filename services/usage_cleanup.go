package services

import (
	"fmt"
	"sync"
	"time"

	"Curry2API-go/database"
	"github.com/sirupsen/logrus"
)

// CleanupConfig holds configuration for the usage cleanup service
type CleanupConfig struct {
	Enabled        bool          // Enable/disable cleanup
	RetentionDays  int           // Number of days to retain usage records
	BatchSize      int           // Number of records to delete per batch
	ScheduleHour   int           // Hour of day to run cleanup (0-23, UTC)
	ScheduleMinute int           // Minute of hour to run cleanup (0-59)
}

// DefaultCleanupConfig returns the default cleanup configuration
func DefaultCleanupConfig() *CleanupConfig {
	return &CleanupConfig{
		Enabled:        true,
		RetentionDays:  90,  // Default 90 days retention
		BatchSize:      1000,
		ScheduleHour:   3,   // 3 AM UTC
		ScheduleMinute: 0,
	}
}

// UsageCleanupService manages periodic cleanup of old usage records
type UsageCleanupService struct {
	config      *CleanupConfig
	stopChan    chan struct{}
	wg          sync.WaitGroup
	mu          sync.RWMutex
	running     bool
	lastCleanup time.Time
	lastError   error
}

var (
	cleanupInstance *UsageCleanupService
	cleanupOnce     sync.Once
)

// NewUsageCleanupService creates a new UsageCleanupService instance
func NewUsageCleanupService(config *CleanupConfig) *UsageCleanupService {
	if config == nil {
		config = DefaultCleanupConfig()
	}

	// Validate minimum retention period
	if config.RetentionDays < 7 {
		logrus.Warnf("Retention period %d days is below minimum (7 days), using 7 days", config.RetentionDays)
		config.RetentionDays = 7
	}

	return &UsageCleanupService{
		config:   config,
		stopChan: make(chan struct{}),
	}
}

// GetUsageCleanupService returns the singleton instance
func GetUsageCleanupService() *UsageCleanupService {
	cleanupOnce.Do(func() {
		cleanupInstance = NewUsageCleanupService(nil)
	})
	return cleanupInstance
}

// InitUsageCleanupService initializes the singleton with a specific config
func InitUsageCleanupService(config *CleanupConfig) *UsageCleanupService {
	cleanupOnce.Do(func() {
		cleanupInstance = NewUsageCleanupService(config)
	})
	return cleanupInstance
}

// Start begins the cleanup scheduler
func (s *UsageCleanupService) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		logrus.Warn("Usage cleanup service is already running")
		return
	}
	s.running = true
	s.mu.Unlock()

	if !s.config.Enabled {
		logrus.Info("Usage cleanup service is disabled")
		return
	}

	s.wg.Add(1)
	go s.runScheduler()
	logrus.Infof("Usage cleanup service started (retention: %d days, schedule: %02d:%02d UTC)",
		s.config.RetentionDays, s.config.ScheduleHour, s.config.ScheduleMinute)
}

// Stop gracefully stops the cleanup scheduler
func (s *UsageCleanupService) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopChan)
	s.wg.Wait()
	logrus.Info("Usage cleanup service stopped")
}

// IsRunning returns whether the service is running
func (s *UsageCleanupService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetConfig returns the current configuration
func (s *UsageCleanupService) GetConfig() *CleanupConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// UpdateRetentionDays updates the retention period
func (s *UsageCleanupService) UpdateRetentionDays(days int) error {
	if days < 7 {
		return fmt.Errorf("retention period must be at least 7 days")
	}

	s.mu.Lock()
	s.config.RetentionDays = days
	s.mu.Unlock()

	logrus.Infof("Updated retention period to %d days", days)
	return nil
}

// GetLastCleanup returns the time of the last cleanup
func (s *UsageCleanupService) GetLastCleanup() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastCleanup
}

// GetLastError returns the last error from cleanup
func (s *UsageCleanupService) GetLastError() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastError
}

// runScheduler runs the cleanup scheduler
func (s *UsageCleanupService) runScheduler() {
	defer s.wg.Done()

	for {
		// Calculate time until next scheduled cleanup
		nextRun := s.calculateNextRun()
		duration := time.Until(nextRun)

		logrus.Infof("Next usage cleanup scheduled for %s (in %v)", nextRun.Format(time.RFC3339), duration)

		select {
		case <-time.After(duration):
			s.performCleanup()
		case <-s.stopChan:
			logrus.Info("Cleanup scheduler received stop signal")
			return
		}
	}
}

// calculateNextRun calculates the next scheduled cleanup time
func (s *UsageCleanupService) calculateNextRun() time.Time {
	now := time.Now().UTC()
	
	// Create today's scheduled time
	scheduled := time.Date(
		now.Year(), now.Month(), now.Day(),
		s.config.ScheduleHour, s.config.ScheduleMinute, 0, 0,
		time.UTC,
	)

	// If we've already passed today's scheduled time, schedule for tomorrow
	if now.After(scheduled) {
		scheduled = scheduled.Add(24 * time.Hour)
	}

	return scheduled
}

// performCleanup executes the cleanup operation
func (s *UsageCleanupService) performCleanup() {
	startTime := time.Now()
	logrus.Info("Starting usage records cleanup...")

	// Calculate cutoff date
	cutoffDate := time.Now().AddDate(0, 0, -s.config.RetentionDays)

	// First, preserve aggregate statistics before deletion
	if err := s.preserveAggregates(cutoffDate); err != nil {
		logrus.Errorf("Failed to preserve aggregates: %v", err)
		s.mu.Lock()
		s.lastError = err
		s.mu.Unlock()
		// Continue with cleanup even if aggregate preservation fails
	}

	// Perform the cleanup
	deletedCount, err := s.CleanupOldRecords(s.config.RetentionDays)
	
	s.mu.Lock()
	s.lastCleanup = time.Now()
	s.lastError = err
	s.mu.Unlock()

	duration := time.Since(startTime)
	if err != nil {
		logrus.Errorf("Cleanup completed with errors in %v: %v", duration, err)
	} else {
		logrus.Infof("Cleanup completed successfully in %v: deleted %d records", duration, deletedCount)
	}
}

// CleanupOldRecords deletes usage records older than the retention period
func (s *UsageCleanupService) CleanupOldRecords(retentionDays int) (int64, error) {
	if retentionDays < 7 {
		return 0, fmt.Errorf("retention period must be at least 7 days")
	}

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	logrus.Infof("Cleaning up usage records older than %s", cutoffDate.Format("2006-01-02"))

	totalDeleted, err := database.DeleteOldUsageRecords(cutoffDate, s.config.BatchSize)
	if err != nil {
		return totalDeleted, fmt.Errorf("failed to delete old records: %w", err)
	}

	logrus.Infof("Deleted %d usage records older than %d days", totalDeleted, retentionDays)
	return totalDeleted, nil
}

// preserveAggregates saves aggregate statistics before deletion
func (s *UsageCleanupService) preserveAggregates(cutoffDate time.Time) error {
	logrus.Infof("Preserving aggregate statistics for records before %s", cutoffDate.Format("2006-01-02"))
	
	return database.PreserveUsageAggregates(cutoffDate)
}

// RunCleanupNow triggers an immediate cleanup (for admin use)
func (s *UsageCleanupService) RunCleanupNow() (int64, error) {
	logrus.Info("Manual cleanup triggered")
	
	// Preserve aggregates first
	cutoffDate := time.Now().AddDate(0, 0, -s.config.RetentionDays)
	if err := s.preserveAggregates(cutoffDate); err != nil {
		logrus.Warnf("Failed to preserve aggregates during manual cleanup: %v", err)
	}
	
	return s.CleanupOldRecords(s.config.RetentionDays)
}
