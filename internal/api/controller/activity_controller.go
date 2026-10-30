package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/hoangnecon/golab/internal/domain"
	"github.com/hoangnecon/golab/internal/service"
)

// ActivityController handles audit logs and payload details.
type ActivityController struct {
	activitySvc *service.ActivityService
}

func NewActivityController(activitySvc *service.ActivityService) *ActivityController {
	return &ActivityController{activitySvc: activitySvc}
}

func (c *ActivityController) HandleList(w http.ResponseWriter, r *http.Request) {
	logs, err := c.activitySvc.ListRecentActivity(r.Context(), 50)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"activity": logs})
}

func (c *ActivityController) HandleGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		WriteError(w, http.StatusBadRequest, "invalid activity id")
		return
	}
	detail, err := c.activitySvc.GetActivity(r.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "activity not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, detail)
}
