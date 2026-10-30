package controller_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hoangnecon/golab/internal/api/controller"
	"github.com/hoangnecon/golab/internal/service"
	"github.com/hoangnecon/golab/internal/store/sqlite"
)

func TestProjectController(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "ctrl_test.db")
	st, err := sqlite.New(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	projectSvc := service.NewProjectService(st)
	ctrl := controller.NewProjectController(projectSvc)

	// 1. Create Project
	body := `{"name":"Test Project","notebook_url":"https://colab.research.google.com/drive/xyz","owner_profile_id":"default"}`
	req := httptest.NewRequest("POST", "/api/projects", strings.NewReader(body))
	w := httptest.NewRecorder()
	ctrl.HandleCreate(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	// 2. List Projects
	reqList := httptest.NewRequest("GET", "/api/projects", nil)
	wList := httptest.NewRecorder()
	ctrl.HandleList(wList, reqList)

	if wList.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", wList.Code, wList.Body.String())
	}
	if !strings.Contains(wList.Body.String(), "Test Project") {
		t.Fatalf("expected project in list, got: %s", wList.Body.String())
	}
}
