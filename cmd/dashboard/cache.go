package main

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// refreshResourceCache fetches fresh resources from the cloud adapter
func (s *server) refreshResourceCache() {
	s.resourceCache.refreshMu.Lock()
	if s.resourceCache.isRefreshing {
		s.resourceCache.refreshMu.Unlock()
		return
	}
	s.resourceCache.isRefreshing = true
	s.resourceCache.refreshMu.Unlock()

	defer func() {
		s.resourceCache.refreshMu.Lock()
		s.resourceCache.isRefreshing = false
		s.resourceCache.refreshMu.Unlock()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.logger.Info("fetching fresh resources from cloud")
	resources, err := s.adapter.FetchResources(ctx)
	if err != nil {
		s.logger.Error("failed to refresh resource cache", zap.Error(err))
		return
	}

	s.resourceCache.Lock()
	defer s.resourceCache.Unlock()

	s.resourceCache.resources = resources
	s.resourceCache.fetchedAt = time.Now()
	s.logger.Info("resource cache refreshed successfully", zap.Int("count", len(resources)))
}
