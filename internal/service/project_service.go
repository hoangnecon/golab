package service

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hoangnecon/golab/internal/domain"
)

// ProjectService handles business use-cases for Notebook Projects.
type ProjectService struct {
	repo domain.ProjectRepository
}

func NewProjectService(repo domain.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) CreateProject(ctx context.Context, name, notebookURL, ownerProfileID string) (*domain.Project, error) {
	name = strings.TrimSpace(name)
	notebookURL = strings.TrimSpace(notebookURL)
	ownerProfileID = strings.TrimSpace(ownerProfileID)

	if name == "" {
		return nil, errors.New("project name cannot be empty")
	}
	if ownerProfileID == "" {
		return nil, errors.New("owner_profile_id is required")
	}

	u, err := url.Parse(notebookURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, errors.New("invalid notebook URL: must be a valid http or https URL")
	}

	now := time.Now().UTC()
	p := &domain.Project{
		ID:             uuid.New().String(),
		Name:           name,
		NotebookURL:    notebookURL,
		OwnerProfileID: ownerProfileID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.CreateProject(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProjectService) GetProject(ctx context.Context, id string) (*domain.Project, error) {
	return s.repo.GetProject(ctx, id)
}

func (s *ProjectService) ListProjects(ctx context.Context) ([]*domain.Project, error) {
	return s.repo.ListProjects(ctx)
}

func (s *ProjectService) UpdateProject(ctx context.Context, p *domain.Project) error {
	if p.ID == "" {
		return errors.New("project ID cannot be empty")
	}
	return s.repo.UpdateProject(ctx, p)
}

func (s *ProjectService) DeleteProject(ctx context.Context, id string) error {
	return s.repo.DeleteProject(ctx, id)
}
