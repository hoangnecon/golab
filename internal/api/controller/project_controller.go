package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hoangnecon/golab/internal/service"
)

// ProjectController handles HTTP requests for notebook Projects.
type ProjectController struct {
	projectSvc *service.ProjectService
	sessionSvc *service.SessionService
}

func NewProjectController(projectSvc *service.ProjectService) *ProjectController {
	return &ProjectController{projectSvc: projectSvc}
}

func (c *ProjectController) SetSessionService(sessionSvc *service.SessionService) {
	c.sessionSvc = sessionSvc
}

func (c *ProjectController) HandleList(w http.ResponseWriter, r *http.Request) {
	list, err := c.projectSvc.ListProjects(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"projects": list})
}

func (c *ProjectController) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name           string `json:"name"`
		NotebookURL    string `json:"notebook_url"`
		OwnerProfileID string `json:"owner_profile_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	p, err := c.projectSvc.CreateProject(r.Context(), input.Name, input.NotebookURL, input.OwnerProfileID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, p)
}

func (c *ProjectController) HandleGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := c.projectSvc.GetProject(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, p)
}

func (c *ProjectController) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := c.projectSvc.GetProject(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	var input struct {
		Name           string `json:"name"`
		NotebookURL    string `json:"notebook_url"`
		OwnerProfileID string `json:"owner_profile_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.Name != "" {
		p.Name = input.Name
	}
	if input.NotebookURL != "" {
		p.NotebookURL = input.NotebookURL
	}
	if input.OwnerProfileID != "" {
		p.OwnerProfileID = input.OwnerProfileID
	}
	if err := c.projectSvc.UpdateProject(r.Context(), p); err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, p)
}

func (c *ProjectController) HandleDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := c.projectSvc.DeleteProject(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (c *ProjectController) HandleLaunch(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := c.projectSvc.GetProject(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	var input struct {
		ProfileID string `json:"profile_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&input)
	profileID := input.ProfileID
	if profileID == "" {
		profileID = p.OwnerProfileID
	}

	if c.sessionSvc == nil {
		WriteError(w, http.StatusInternalServerError, "session service unavailable")
		return
	}

	act, err := c.sessionSvc.OpenSession(r.Context(), profileID, p.NotebookURL, p.Name)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, fmt.Sprintf("launch notebook: %v", err))
		return
	}
	WriteJSON(w, http.StatusOK, act.Session)
}
