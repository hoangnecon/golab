package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hoangnecon/golab/internal/service"
)

// SessionController handles HTTP requests for active notebook Sessions and runtime actions.
type SessionController struct {
	sessionSvc *service.SessionService
}

func NewSessionController(sessionSvc *service.SessionService) *SessionController {
	return &SessionController{sessionSvc: sessionSvc}
}

func (c *SessionController) HandleList(w http.ResponseWriter, r *http.Request) {
	sessions, err := c.sessionSvc.ListSessions(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"sessions": sessions})
}

func (c *SessionController) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProfileID   string `json:"profile_id"`
		NotebookURL string `json:"notebook_url"`
		Name        string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.ProfileID == "" || input.NotebookURL == "" {
		WriteError(w, http.StatusBadRequest, "profile_id and notebook_url are required")
		return
	}
	act, err := c.sessionSvc.OpenSession(r.Context(), input.ProfileID, input.NotebookURL, input.Name)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, fmt.Sprintf("open notebook: %v", err))
		return
	}
	WriteJSON(w, http.StatusCreated, act.Session)
}

func (c *SessionController) HandleActivate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := c.sessionSvc.SetActiveSession(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"activated": true, "session_id": id})
}

func (c *SessionController) HandleFocus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := c.sessionSvc.FocusSession(r.Context(), id); err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"focused": true})
}

func (c *SessionController) HandleInterrupt(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := c.sessionSvc.InterruptKernel(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, fmt.Sprintf("interrupt failed: %v", err))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"interrupted": true, "message": "SIGINT signal dispatched to kernel."})
}

func (c *SessionController) HandleRestart(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	detail, err := c.sessionSvc.RestartKernel(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, fmt.Sprintf("restart failed: %v", err))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"restarted": true, "kernel": detail})
}

func (c *SessionController) HandleDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := c.sessionSvc.CloseSession(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"closed": true})
}

func (c *SessionController) HandleToggleLock(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var input struct {
		Locked bool `json:"locked"`
	}
	_ = json.NewDecoder(r.Body).Decode(&input)
	if err := c.sessionSvc.SetSessionLocked(r.Context(), id, input.Locked); err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"session_id": id, "is_locked": input.Locked})
}

func (c *SessionController) HandleConnect(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := c.sessionSvc.ConnectRuntime(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, fmt.Sprintf("connect failed: %v", err))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"connected": true, "message": "Hosted VM allocated and connected."})
}

func (c *SessionController) HandleDisconnect(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := c.sessionSvc.DisconnectRuntime(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, fmt.Sprintf("disconnect failed: %v", err))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"disconnected": true, "message": "VM runtime powerwashed and released."})
}

func (c *SessionController) HandleSave(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := c.sessionSvc.SaveNotebook(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, fmt.Sprintf("save failed: %v", err))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"saved": true, "message": "Notebook changes committed to Google Drive (Ctrl+S)."})
}

func (c *SessionController) HandleRunAll(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := c.sessionSvc.RunAll(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, fmt.Sprintf("run all failed: %v", err))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"triggered": true, "message": "All cells scheduled for execution."})
}

func (c *SessionController) HandleRunningCells(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	report, err := c.sessionSvc.GetRunningCells(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, report)
}

func (c *SessionController) HandleAccelerators(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	report, err := c.sessionSvc.GetAvailableAccelerators(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, report)
}

func (c *SessionController) HandleSetAccelerator(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var input struct {
		Accelerator string `json:"accelerator"`
		HighRAM     bool   `json:"high_ram"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.sessionSvc.SetAccelerator(r.Context(), id, input.Accelerator, input.HighRAM); err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"success": true, "accelerator": input.Accelerator, "high_ram": input.HighRAM})
}
