package container

import (
	"context"
	"time"

	"github.com/launchstack/backend/db"
	"github.com/sirupsen/logrus"
)

// StatsCollector collects resource usage stats from containers
type StatsCollector struct {
	manager        Manager
	logger         *logrus.Logger
	collectionRate time.Duration
	stopChan       chan struct{}
	maxPerInstance int
}

// NewStatsCollector creates a new stats collector
func NewStatsCollector(manager Manager, logger *logrus.Logger, collectionRate time.Duration) *StatsCollector {
	return &StatsCollector{
		manager:        manager,
		logger:         logger,
		collectionRate: collectionRate,
		stopChan:       make(chan struct{}),
		maxPerInstance: 100, // Keep last 100 stats per instance
	}
}

// Start starts the stats collector
func (s *StatsCollector) Start() {
	go s.run()
}

// Stop stops the stats collector
func (s *StatsCollector) Stop() {
	close(s.stopChan)
}

// run is the main loop for the stats collector
func (s *StatsCollector) run() {
	ticker := time.NewTicker(s.collectionRate)
	defer ticker.Stop()
	
	// Run cleanup on startup
	if err := s.cleanupOldStats(); err != nil {
		s.logger.WithError(err).Error("Failed to clean up old stats")
	}
	
	for {
		select {
		case <-ticker.C:
			if err := s.collectStats(); err != nil {
				s.logger.WithError(err).Error("Failed to collect stats")
			}
		case <-s.stopChan:
			s.logger.Info("Stats collector stopped")
			return
		}
	}
}

// collectStats collects stats for all running instances
func (s *StatsCollector) collectStats() error {
	// Get all running instances
	instances, err := db.GetRunningInstances()
	if err != nil {
		return err
	}
	
	s.logger.WithField("count", len(instances)).Debug("Collecting stats for running instances")
	
	// Collect stats for each instance
	for _, instance := range instances {
		// Skip if no container ID
		if instance.ContainerID == "" {
			continue
		}
		
		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		
		// Collect stats
		_, err := s.manager.GetInstanceStats(ctx, instance.ID)
		if err != nil {
			s.logger.WithFields(logrus.Fields{
				"instance_id":  instance.ID,
				"container_id": instance.ContainerID,
				"error":        err,
			}).Error("Failed to collect stats for instance")
		}
		
		cancel()
	}
	
	return nil
}

// cleanupOldStats removes old stats records to prevent database growth
func (s *StatsCollector) cleanupOldStats() error {
	return db.PruneResourceUsage(s.maxPerInstance)
} 