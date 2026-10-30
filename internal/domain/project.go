package domain

import (
	"context"
	"time"
)

// Project represents a saved notebook project with its designated owner profile.
type Project struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	NotebookURL    string    `json:"notebook_url"`
	OwnerProfileID string    `json:"owner_profile_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ProjectRepository defines persistence operations for Projects.
type ProjectRepository interface {
	CreateProject(ctx context.Context, p *Project) error
	GetProject(ctx context.Context, id string) (*Project, error)
	GetProjectByURL(ctx context.Context, notebookURL string) (*Project, error)
	ListProjects(ctx context.Context) ([]*Project, error)
	UpdateProject(ctx context.Context, p *Project) error
	DeleteProject(ctx context.Context, id string) error
}
