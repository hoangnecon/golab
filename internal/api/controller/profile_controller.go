package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hoangnecon/golab/internal/domain"
	"github.com/hoangnecon/golab/internal/service"
)

// ProfileController handles HTTP requests for Chrome profiles.
type ProfileController struct {
	profileSvc *service.ProfileService
}

func NewProfileController(profileSvc *service.ProfileService) *ProfileController {
	return &ProfileController{profileSvc: profileSvc}
}

func (c *ProfileController) HandleList(w http.ResponseWriter, r *http.Request) {
	profiles, err := c.profileSvc.ListProfiles(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"profiles": profiles})
}

func (c *ProfileController) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name         string             `json:"name"`
		AccountEmail string             `json:"account_email"`
		Tier         domain.ProfileTier `json:"tier"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	p, err := c.profileSvc.CreateProfile(r.Context(), input.Name, input.AccountEmail, input.Tier)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, p)
}

func (c *ProfileController) HandleLaunch(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := c.profileSvc.LaunchProfile(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, fmt.Sprintf("launch profile: %v", err))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"launched": true, "profile_id": id})
}

func (c *ProfileController) HandleStop(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := c.profileSvc.StopProfile(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, fmt.Sprintf("stop profile: %v", err))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"stopped": true, "profile_id": id})
}

func (c *ProfileController) HandleDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := c.profileSvc.DeleteProfile(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}
