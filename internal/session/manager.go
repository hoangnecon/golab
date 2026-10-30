package session

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hoangnecon/golab/internal/bridge"
	"github.com/hoangnecon/golab/internal/domain"
	"github.com/hoangnecon/golab/internal/process"
	"github.com/hoangnecon/golab/internal/service/kernel"
	"github.com/hoangnecon/golab/internal/service/notebook"
	"github.com/hoangnecon/golab/internal/service/proxy"
	"github.com/hoangnecon/golab/internal/service/storage"
	"github.com/hoangnecon/golab/internal/service/telemetry"
	"github.com/hoangnecon/golab/internal/service/terminal"
	"github.com/hoangnecon/golab/internal/store"
)

// DomainServices bundles all enterprise domain services for a specific Colab session.
type DomainServices struct {
	Storage   *storage.Service
	Kernel    *kernel.Service
	Telemetry *telemetry.Service
	Notebook  *notebook.Service
	Terminal  *terminal.Service
	Proxy     *proxy.Service
}

// ActiveSession holds live connections and services for a session.
type ActiveSession struct {
	Session   *store.Session
	Profile   *store.Profile
	CDP       *bridge.CDPClient
	Executor  *bridge.ColabExecutor
	Services  *DomainServices
	CreatedAt time.Time
}

// Manager orchestrates multiple Colab sessions across various Chrome profiles.
type Manager struct {
	store   *store.Store
	procMgr *process.Manager
	mu      sync.RWMutex
	active  map[string]*ActiveSession
}

func NewManager(st *store.Store, procMgr *process.Manager) *Manager {
	return &Manager{
		store:   st,
		procMgr: procMgr,
		active:  make(map[string]*ActiveSession),
	}
}

// Store returns the underlying persistent store.
func (m *Manager) Store() *store.Store {
	return m.store
}

// ProcessManager returns the Chrome process supervisor.
func (m *Manager) ProcessManager() *process.Manager {
	return m.procMgr
}

// ResolveSession returns the active session requested by ID, or defaults to the active/most recent session.
func (m *Manager) ResolveSession(ctx context.Context, sessionID string) (*ActiveSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check Session Lock guardrail
	lockedSess, _ := m.store.GetLockedSession(ctx)
	if lockedSess != nil {
		if sessionID != "" && sessionID != lockedSess.ID {
			return nil, fmt.Errorf("session guardrail active: session %q is locked as exclusive worktree. Operations on session %q are blocked", lockedSess.ID, sessionID)
		}
		sessionID = lockedSess.ID
	}

	// 1. If explicit sessionID requested, look up in active map
	if sessionID != "" {
		if sess, ok := m.active[sessionID]; ok {
			return sess, nil
		}
		// Try to revive from store
		dbSess, err := m.store.GetSession(ctx, sessionID)
		if err == nil && dbSess != nil {
			act, errAttach := m.attachLocked(ctx, dbSess)
			if errAttach == nil {
				return act, nil
			}
		}
		return nil, fmt.Errorf("session %q not found or not connected", sessionID)
	}

	// 2. If no sessionID given, try to find active session in store
	dbSess, err := m.store.GetActiveSession(ctx)
	if err == nil && dbSess != nil {
		if act, ok := m.active[dbSess.ID]; ok {
			return act, nil
		}
		act, errAttach := m.attachLocked(ctx, dbSess)
		if errAttach == nil {
			return act, nil
		}
	}

	// 3. Fallback: If any active session exists in memory, return one
	for _, s := range m.active {
		return s, nil
	}

	// 4. Default fallback: Auto-discover from default CDP port 9222
	defaultCDP := "http://127.0.0.1:9222"
	cdp := bridge.NewCDPClient(defaultCDP)
	if err := cdp.Connect(ctx); err == nil {
		target := cdp.Target()
		targetURL := ""
		title := "Default Colab Session"
		tabID := ""
		if target != nil {
			targetURL = target.URL
			title = target.Title
			tabID = target.ID
		}

		exec := bridge.NewColabExecutor(cdp)
		svcs := createDomainServices(exec)
		sessRecord := &store.Session{
			ID:          "default",
			Name:        title,
			ProfileID:   "default",
			NotebookURL: targetURL,
			TabID:       tabID,
			IsActive:    true,
			CreatedAt:   time.Now().UTC(),
		}
		_ = m.store.SaveSession(ctx, sessRecord)

		act := &ActiveSession{
			Session:   sessRecord,
			CDP:       cdp,
			Executor:  exec,
			Services:  svcs,
			CreatedAt: time.Now(),
		}
		m.active["default"] = act
		return act, nil
	}

	return nil, errors.New("no active colab session found. please open a colab notebook or launch a profile via golab ui")
}

func createDomainServices(exec *bridge.ColabExecutor) *DomainServices {
	return &DomainServices{
		Storage:   storage.NewService(exec),
		Kernel:    kernel.NewService(exec),
		Telemetry: telemetry.NewService(exec),
		Notebook:  notebook.NewService(exec),
		Terminal:  terminal.NewService(exec),
		Proxy:     proxy.NewService(exec),
	}
}

func (m *Manager) attachLocked(ctx context.Context, dbSess *store.Session) (*ActiveSession, error) {
	p, err := m.store.GetProfile(ctx, dbSess.ProfileID)
	if err != nil {
		return nil, fmt.Errorf("profile not found: %w", err)
	}

	// Ensure Chrome is running
	if !m.procMgr.IsProfileRunning(p) {
		if err := m.procMgr.LaunchProfile(ctx, p, dbSess.NotebookURL); err != nil {
			return nil, fmt.Errorf("launch profile: %w", err)
		}
	}

	cdpURL := fmt.Sprintf("http://127.0.0.1:%d", p.CDPPort)
	cdp := bridge.NewCDPClient(cdpURL)
	if err := cdp.ConnectNotebook(ctx, dbSess.NotebookURL); err != nil {
		return nil, fmt.Errorf("connect cdp to notebook: %w", err)
	}

	exec := bridge.NewColabExecutor(cdp)
	svcs := createDomainServices(exec)

	act := &ActiveSession{
		Session:   dbSess,
		Profile:   p,
		CDP:       cdp,
		Executor:  exec,
		Services:  svcs,
		CreatedAt: time.Now(),
	}
	m.active[dbSess.ID] = act
	return act, nil
}

// OpenNotebook creates or attaches to a notebook under a specific profile.
func (m *Manager) OpenNotebook(ctx context.Context, profileID, notebookURL, name string) (*ActiveSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, err := m.store.GetProfile(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("profile %q: %w", profileID, err)
	}

	// 1. Ensure Chrome is running
	if !m.procMgr.IsProfileRunning(p) {
		if err := m.procMgr.LaunchProfile(ctx, p, notebookURL); err != nil {
			return nil, fmt.Errorf("launch chrome: %w", err)
		}
	}

	cdpURL := fmt.Sprintf("http://127.0.0.1:%d", p.CDPPort)
	cdp := bridge.NewCDPClient(cdpURL)

	// 2. Open tab or connect
	var target *bridge.TargetInfo
	targets, _ := cdp.ListTargets(ctx)
	for i := range targets {
		if strings.Contains(targets[i].URL, notebookURL) {
			target = &targets[i]
			break
		}
	}

	if target == nil {
		newTarget, errOpen := cdp.OpenTab(ctx, notebookURL)
		if errOpen != nil {
			return nil, fmt.Errorf("open tab: %w", errOpen)
		}
		target = newTarget
		time.Sleep(1 * time.Second) // allow tab to register
	}

	if errConn := cdp.ConnectNotebook(ctx, notebookURL); errConn != nil {
		return nil, fmt.Errorf("connect cdp: %w", errConn)
	}

	sessID := uuid.New().String()
	if name == "" {
		name = extractNotebookName(notebookURL)
	}

	projectID := ""
	ownerProfileID := profileID
	if proj, errProj := m.store.GetProjectByURL(ctx, notebookURL); errProj == nil && proj != nil {
		projectID = proj.ID
		ownerProfileID = proj.OwnerProfileID
	}

	permission := domain.PermissionEdit
	titleLower := strings.ToLower(target.Title)
	if strings.Contains(titleLower, "từ chối") || strings.Contains(titleLower, "denied") {
		permission = domain.PermissionDenied
	} else if strings.Contains(titleLower, "playground") || strings.Contains(titleLower, "view only") {
		permission = domain.PermissionView
	}

	sessRecord := &store.Session{
		ID:             sessID,
		Name:           name,
		ProfileID:      profileID,
		ProjectID:      projectID,
		OwnerProfileID: ownerProfileID,
		NotebookURL:    notebookURL,
		TabID:          target.ID,
		IsActive:       true,
		Permission:     permission,
		CreatedAt:      time.Now().UTC(),
	}

	_ = m.store.SetActiveSession(ctx, sessID)
	_ = m.store.SaveSession(ctx, sessRecord)

	exec := bridge.NewColabExecutor(cdp)
	svcs := createDomainServices(exec)

	act := &ActiveSession{
		Session:   sessRecord,
		Profile:   p,
		CDP:       cdp,
		Executor:  exec,
		Services:  svcs,
		CreatedAt: time.Now(),
	}

	m.active[sessID] = act
	log.Printf("[SessionManager] Successfully attached session %q for profile %q", sessID, p.Name)
	return act, nil
}

func extractNotebookName(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "Colab Notebook"
	}
	parts := strings.Split(u.Path, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" && strings.HasSuffix(parts[i], ".ipynb") {
			return parts[i]
		}
	}
	return "Colab Notebook"
}

// CloseSession closes tab and disconnects CDP.
func (m *Manager) CloseSession(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	act, ok := m.active[sessionID]
	if ok {
		if act.Session.TabID != "" {
			_ = act.CDP.CloseTab(ctx, act.Session.TabID)
		}
		_ = act.CDP.Close()
		delete(m.active, sessionID)
	}

	return m.store.DeleteSession(ctx, sessionID)
}

// ListLiveSessions returns all registered sessions with real-time hardware and kernel status.
func (m *Manager) ListLiveSessions(ctx context.Context) ([]*store.Session, error) {
	sessions, err := m.store.ListSessions(ctx)
	if err != nil {
		return nil, err
	}

	// Enhance each session with telemetry if connected
	for _, s := range sessions {
		m.mu.RLock()
		act, ok := m.active[s.ID]
		m.mu.RUnlock()

		if ok && act.Services != nil {
			subCtx, cancel := context.WithTimeout(ctx, 800*time.Millisecond)
			kStat, errK := act.Services.Kernel.GetStatusReport(subCtx)
			if errK == nil && kStat != nil && kStat.Kernel != nil {
				s.KernelState = kStat.Kernel.ExecutionState
			}

			telem, errT := act.Services.Telemetry.GetHardwareTelemetry(subCtx)
			if errT == nil && telem != nil {
				s.RAMFreeGB = telem.RAM.FreeGB
				if len(telem.GPUs) > 0 {
					s.GPUName = telem.GPUs[0].Name
					s.GPUMemoryFree = int64(telem.GPUs[0].MemoryFreeMB)
					s.GPUMemoryTotal = int64(telem.GPUs[0].MemoryTotalMB)
				}
			}

			drv, errD := act.Services.Telemetry.CheckDriveStatus(subCtx)
			if errD == nil && drv != nil {
				s.DriveMounted = drv.Mounted
			}
			cancel()
		}
	}

	return sessions, nil
}
