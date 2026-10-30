package service

import (
	"context"

	"github.com/hoangnecon/golab/internal/domain"
)

// ActivityService handles audit logs and payload captures.
type ActivityService struct {
	repo domain.ActivityRepository
}

func NewActivityService(repo domain.ActivityRepository) *ActivityService {
	return &ActivityService{repo: repo}
}

func (s *ActivityService) LogActivity(ctx context.Context, sessionID, toolName string, latencyMs int64, status, errorMessage string) error {
	return s.repo.LogActivity(ctx, sessionID, toolName, latencyMs, status, errorMessage)
}

func (s *ActivityService) RecordActivity(ctx context.Context, rec *domain.ActivityRecord) (int64, error) {
	return s.repo.RecordActivity(ctx, rec)
}

func (s *ActivityService) ListRecentActivity(ctx context.Context, limit int) ([]*domain.ActivityLog, error) {
	return s.repo.ListRecentActivity(ctx, limit)
}

func (s *ActivityService) GetActivity(ctx context.Context, id int64) (*domain.ActivityDetail, error) {
	return s.repo.GetActivity(ctx, id)
}
