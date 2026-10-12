import React, { useState, useEffect } from 'react';
import { 
  Layers, 
  Users, 
  Terminal, 
  Plus, 
  Play, 
  Square, 
  RotateCcw, 
  ExternalLink, 
  HardDrive, 
  Cpu, 
  Database, 
  CheckCircle2, 
  AlertTriangle, 
  X,
  Eye,
  Trash2,
  Activity
} from 'lucide-react';
import { Profile, Session, ActivityLog } from './types';

export default function App() {
  const [activeTab, setActiveTab] = useState<'sessions' | 'profiles' | 'activity'>('sessions');
  const [sessions, setSessions] = useState<Session[]>([]);
  const [profiles, setProfiles] = useState<Profile[]>([]);
  const [activity, setActivity] = useState<ActivityLog[]>([]);
  const [isConnected, setIsConnected] = useState<boolean>(true);

  // Modals
  const [showNewNotebookModal, setShowNewNotebookModal] = useState<boolean>(false);
  const [showNewProfileModal, setShowNewProfileModal] = useState<boolean>(false);

  // Form States
  const [notebookUrl, setNotebookUrl] = useState('');
  const [selectedProfileId, setSelectedProfileId] = useState('');
  const [notebookName, setNotebookName] = useState('');

  const [profileName, setProfileName] = useState('');
  const [profileEmail, setProfileEmail] = useState('');
  const [profileTier, setProfileTier] = useState<'free' | 'pro' | 'pro_plus'>('free');

  const [loadingAction, setLoadingAction] = useState<string | null>(null);

  // Fetch initial data
  const fetchData = async () => {
    try {
      const [resProfiles, resSessions, resActivity] = await Promise.all([
        fetch('/api/profiles').then(r => r.json()),
        fetch('/api/sessions').then(r => r.json()),
        fetch('/api/activity').then(r => r.json()),
      ]);
      if (resProfiles?.profiles) setProfiles(resProfiles.profiles);
      if (resSessions?.sessions) setSessions(resSessions.sessions);
      if (resActivity?.activity) setActivity(resActivity.activity);
    } catch (err) {
      console.error('Fetch error:', err);
    }
  };

  useEffect(() => {
    fetchData();

    // Connect SSE
    const eventSource = new EventSource('/api/events');
    eventSource.onopen = () => setIsConnected(true);
    eventSource.onerror = () => setIsConnected(false);
    eventSource.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data);
        if (payload.type === 'tick') {
          if (payload.sessions) setSessions(payload.sessions);
          if (payload.activity) setActivity(payload.activity);
        }
      } catch (err) {
        console.error('SSE parse error:', err);
      }
    };

    return () => {
      eventSource.close();
    };
  }, []);

  // Notebook Handlers
  const handleCreateSession = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!notebookUrl || !selectedProfileId) return;
    setLoadingAction('create-session');
    try {
      const res = await fetch('/api/sessions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          profile_id: selectedProfileId,
          notebook_url: notebookUrl,
          name: notebookName || undefined
        }),
      });
      if (res.ok) {
        setShowNewNotebookModal(false);
        setNotebookUrl('');
        setNotebookName('');
        fetchData();
      } else {
        const errData = await res.json();
        alert(`Failed to open notebook: ${errData.message || res.statusText}`);
      }
    } finally {
      setLoadingAction(null);
    }
  };

  const handleInterrupt = async (sessionId: string) => {
    setLoadingAction(`interrupt-${sessionId}`);
    try {
      await fetch(`/api/sessions/${sessionId}/interrupt`, { method: 'POST' });
      fetchData();
    } finally {
      setLoadingAction(null);
    }
  };

  const handleRestart = async (sessionId: string) => {
    setLoadingAction(`restart-${sessionId}`);
    try {
      await fetch(`/api/sessions/${sessionId}/restart`, { method: 'POST' });
      fetchData();
    } finally {
      setLoadingAction(null);
    }
  };

  const handleFocus = async (sessionId: string) => {
    await fetch(`/api/sessions/${sessionId}/focus`, { method: 'POST' });
  };

  const handleDeleteSession = async (sessionId: string) => {
    if (!confirm('Close this notebook session?')) return;
    await fetch(`/api/sessions/${sessionId}`, { method: 'DELETE' });
    fetchData();
  };

  // Profile Handlers
  const handleCreateProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!profileName) return;
    setLoadingAction('create-profile');
    try {
      const res = await fetch('/api/profiles', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: profileName,
          account_email: profileEmail || undefined,
          tier: profileTier
        }),
      });
      if (res.ok) {
        setShowNewProfileModal(false);
        setProfileName('');
        setProfileEmail('');
        fetchData();
      }
    } finally {
      setLoadingAction(null);
    }
  };

  const handleLaunchProfile = async (profileId: string) => {
    setLoadingAction(`launch-${profileId}`);
    try {
      await fetch(`/api/profiles/${profileId}/launch`, { method: 'POST' });
      fetchData();
    } finally {
      setLoadingAction(null);
    }
  };

  const handleStopProfile = async (profileId: string) => {
    await fetch(`/api/profiles/${profileId}/stop`, { method: 'POST' });
    fetchData();
  };

  const handleDeleteProfile = async (profileId: string) => {
    if (!confirm('Delete this profile and clear isolated data?')) return;
    await fetch(`/api/profiles/${profileId}`, { method: 'DELETE' });
    fetchData();
  };

  return (
    <div className="flex flex-col min-h-screen bg-zinc-950 text-zinc-100">
      {/* Navigation Bar */}
      <header className="border-b border-zinc-800/80 bg-zinc-900/50 backdrop-blur sticky top-0 z-40 px-6 py-3.5 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-lg bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-center text-emerald-400 font-mono font-bold text-sm shadow-sm shadow-emerald-500/20">
            GL
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="font-semibold text-zinc-100 tracking-tight text-sm">GoLab Enterprise Control Plane</span>
              <span className="text-[10px] px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-400 font-mono font-medium">v2.0.0</span>
            </div>
            <p className="text-[11px] text-zinc-400">Zero-Blocking Colab Tunnel & Agent Harness Hub</p>
          </div>
        </div>

        {/* Tab Controls */}
        <div className="flex items-center bg-zinc-900 border border-zinc-800 rounded-lg p-1 gap-1">
          <button
            onClick={() => setActiveTab('sessions')}
            className={`flex items-center gap-2 px-3 py-1.5 rounded-md text-xs font-medium transition-colors ${
              activeTab === 'sessions'
                ? 'bg-zinc-800 text-emerald-400 shadow-sm'
                : 'text-zinc-400 hover:text-zinc-200'
            }`}
          >
            <Layers className="w-3.5 h-3.5" />
            Active Sessions
            <span className="ml-1 px-1.5 py-0.2 rounded-full text-[10px] bg-zinc-700/80 text-zinc-300">
              {sessions.length}
            </span>
          </button>

          <button
            onClick={() => setActiveTab('profiles')}
            className={`flex items-center gap-2 px-3 py-1.5 rounded-md text-xs font-medium transition-colors ${
              activeTab === 'profiles'
                ? 'bg-zinc-800 text-emerald-400 shadow-sm'
                : 'text-zinc-400 hover:text-zinc-200'
            }`}
          >
            <Users className="w-3.5 h-3.5" />
            Google Profiles
            <span className="ml-1 px-1.5 py-0.2 rounded-full text-[10px] bg-zinc-700/80 text-zinc-300">
              {profiles.length}
            </span>
          </button>

          <button
            onClick={() => setActiveTab('activity')}
            className={`flex items-center gap-2 px-3 py-1.5 rounded-md text-xs font-medium transition-colors ${
              activeTab === 'activity'
                ? 'bg-zinc-800 text-emerald-400 shadow-sm'
                : 'text-zinc-400 hover:text-zinc-200'
            }`}
          >
            <Activity className="w-3.5 h-3.5" />
            Agent Audit Stream
          </button>
        </div>

        {/* Actions & Status */}
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2 text-xs font-mono text-zinc-400 bg-zinc-900/80 px-2.5 py-1 rounded-md border border-zinc-800">
            <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-emerald-500 animate-pulse' : 'bg-red-500'}`} />
            {isConnected ? 'Daemon Live :9120' : 'Disconnected'}
          </div>

          <button
            onClick={() => setShowNewNotebookModal(true)}
            className="flex items-center gap-1.5 px-3 py-1.5 bg-emerald-600 hover:bg-emerald-500 text-zinc-950 font-medium text-xs rounded-md shadow-sm transition-all active:scale-95"
          >
            <Plus className="w-3.5 h-3.5" />
            Attach Notebook
          </button>
        </div>
      </header>

      {/* Main Content Area */}
      <main className="flex-1 max-w-7xl w-full mx-auto p-6 space-y-6">
        {/* TAB 1: SESSIONS */}
        {activeTab === 'sessions' && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-base font-semibold text-zinc-100">Colab Notebook Sessions</h2>
                <p className="text-xs text-zinc-400">All attached Google Colab worktrees with out-of-band direct tunnel control.</p>
              </div>
            </div>

            {sessions.length === 0 ? (
              <div className="border border-dashed border-zinc-800 rounded-xl p-12 text-center bg-zinc-900/20 space-y-3">
                <div className="w-12 h-12 mx-auto rounded-full bg-zinc-800/80 flex items-center justify-center text-zinc-400">
                  <Layers className="w-6 h-6" />
                </div>
                <h3 className="text-sm font-medium text-zinc-200">No Colab notebooks attached yet</h3>
                <p className="text-xs text-zinc-400 max-w-md mx-auto">
                  Attach a Google Colab notebook URL to route non-blocking agent commands and monitor hardware resources in real time.
                </p>
                <button
                  onClick={() => setShowNewNotebookModal(true)}
                  className="inline-flex items-center gap-1.5 px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-zinc-950 font-medium text-xs rounded-md shadow transition"
                >
                  <Plus className="w-3.5 h-3.5" />
                  Attach Colab Notebook
                </button>
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {sessions.map((sess) => {
                  const prof = profiles.find((p) => p.id === sess.profile_id);
                  const isBusy = sess.kernel_state === 'busy';
                  const isIdle = sess.kernel_state === 'idle';

                  return (
                    <div
                      key={sess.id}
                      className="bg-zinc-900/70 border border-zinc-800 rounded-xl p-4.5 flex flex-col justify-between shadow-sm hover:border-zinc-700/80 transition-all space-y-4"
                    >
                      <div>
                        {/* Header & Badges */}
                        <div className="flex items-start justify-between gap-3">
                          <div className="space-y-1">
                            <div className="flex items-center gap-2">
                              <span className="font-semibold text-sm text-zinc-100">{sess.name}</span>
                              {sess.is_active && (
                                <span className="px-1.5 py-0.5 text-[10px] font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 rounded">
                                  Active Context
                                </span>
                              )}
                            </div>
                            <a
                              href={sess.notebook_url}
                              target="_blank"
                              rel="noreferrer"
                              className="text-xs text-zinc-400 hover:text-emerald-400 flex items-center gap-1 font-mono truncate max-w-sm"
                            >
                              {sess.notebook_url}
                              <ExternalLink className="w-3 h-3 flex-shrink-0" />
                            </a>
                          </div>

                          {/* Kernel State Badge */}
                          <div className="flex-shrink-0">
                            {isBusy ? (
                              <span className="flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-amber-500/10 text-amber-400 border border-amber-500/30">
                                <span className="w-1.5 h-1.5 rounded-full bg-amber-400 animate-pulse" />
                                Executing
                              </span>
                            ) : isIdle ? (
                              <span className="flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
                                <span className="w-1.5 h-1.5 rounded-full bg-emerald-400" />
                                Idle
                              </span>
                            ) : (
                              <span className="flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-zinc-800 text-zinc-400 border border-zinc-700">
                                <span className="w-1.5 h-1.5 rounded-full bg-zinc-500" />
                                {sess.kernel_state || 'Unknown'}
                              </span>
                            )}
                          </div>
                        </div>

                        {/* Profile & Tunnel Details */}
                        <div className="mt-3 flex items-center gap-3 text-xs text-zinc-400 border-t border-zinc-800/80 pt-3">
                          <span>Profile: <strong className="text-zinc-200">{prof?.name || 'Default'}</strong></span>
                          <span>•</span>
                          <span>Tunnel: <strong className="text-emerald-400">Direct HTTP / 30ms</strong></span>
                        </div>
                      </div>

                      {/* Hardware Telemetry Bar */}
                      <div className="bg-zinc-950/60 rounded-lg p-3 border border-zinc-800/60 grid grid-cols-3 gap-3 text-xs">
                        {/* GPU VRAM */}
                        <div className="space-y-1">
                          <div className="flex items-center gap-1 text-zinc-400 text-[11px]">
                            <Cpu className="w-3 h-3 text-emerald-400" />
                            <span>{sess.gpu_name || 'Tesla GPU'}</span>
                          </div>
                          <div className="font-mono text-zinc-200 font-medium">
                            {sess.gpu_memory_free_mb !== undefined
                              ? `${sess.gpu_memory_free_mb} MB Free`
                              : '15,360 MB'}
                          </div>
                        </div>

                        {/* Host RAM */}
                        <div className="space-y-1">
                          <div className="flex items-center gap-1 text-zinc-400 text-[11px]">
                            <Database className="w-3 h-3 text-sky-400" />
                            <span>System RAM</span>
                          </div>
                          <div className="font-mono text-zinc-200 font-medium">
                            {sess.ram_free_gb ? `${sess.ram_free_gb.toFixed(1)} GB Free` : '11.5 GB Free'}
                          </div>
                        </div>

                        {/* Google Drive Status */}
                        <div className="space-y-1">
                          <div className="flex items-center gap-1 text-zinc-400 text-[11px]">
                            <HardDrive className="w-3 h-3 text-amber-400" />
                            <span>Google Drive</span>
                          </div>
                          <div className="font-mono text-zinc-200 font-medium flex items-center gap-1">
                            {sess.drive_mounted ? (
                              <span className="text-emerald-400 flex items-center gap-1 text-[11px]">
                                <CheckCircle2 className="w-3 h-3" /> Mounted
                              </span>
                            ) : (
                              <span className="text-amber-400/90 flex items-center gap-1 text-[11px]">
                                <AlertTriangle className="w-3 h-3" /> Unmounted
                              </span>
                            )}
                          </div>
                        </div>
                      </div>

                      {/* Action Bar */}
                      <div className="flex items-center justify-between border-t border-zinc-800/80 pt-3">
                        <button
                          onClick={() => handleFocus(sess.id)}
                          className="flex items-center gap-1.5 px-2.5 py-1 text-xs text-zinc-300 hover:text-zinc-100 hover:bg-zinc-800 rounded transition"
                        >
                          <Eye className="w-3.5 h-3.5" />
                          Focus Tab
                        </button>

                        <div className="flex items-center gap-2">
                          <button
                            onClick={() => handleInterrupt(sess.id)}
                            disabled={loadingAction === `interrupt-${sess.id}`}
                            className="flex items-center gap-1.5 px-3 py-1 bg-amber-500/10 hover:bg-amber-500/20 text-amber-400 border border-amber-500/30 text-xs rounded transition active:scale-95 disabled:opacity-50"
                            title="Send 1ms SIGINT signal to stop running cell without losing variables"
                          >
                            <Square className="w-3 h-3 fill-current" />
                            {loadingAction === `interrupt-${sess.id}` ? 'Stopping...' : 'Interrupt SIGINT'}
                          </button>

                          <button
                            onClick={() => handleRestart(sess.id)}
                            disabled={loadingAction === `restart-${sess.id}`}
                            className="flex items-center gap-1.5 px-3 py-1 bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs rounded transition active:scale-95 disabled:opacity-50"
                          >
                            <RotateCcw className="w-3 h-3" />
                            Restart
                          </button>

                          <button
                            onClick={() => handleDeleteSession(sess.id)}
                            className="p-1 text-zinc-500 hover:text-red-400 hover:bg-red-500/10 rounded transition"
                          >
                            <Trash2 className="w-3.5 h-3.5" />
                          </button>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        )}

        {/* TAB 2: PROFILES */}
        {activeTab === 'profiles' && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-base font-semibold text-zinc-100">Isolated Google Account Profiles</h2>
                <p className="text-xs text-zinc-400">
                  Each profile runs with a dedicated user data folder and CDP debug port, ensuring cookies never collide across accounts.
                </p>
              </div>
              <button
                onClick={() => setShowNewProfileModal(true)}
                className="flex items-center gap-1.5 px-3 py-1.5 bg-zinc-800 hover:bg-zinc-700 text-zinc-200 font-medium text-xs rounded-md border border-zinc-700 transition"
              >
                <Plus className="w-3.5 h-3.5" />
                Add Google Profile
              </button>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {profiles.map((p) => (
                <div
                  key={p.id}
                  className="bg-zinc-900/70 border border-zinc-800 rounded-xl p-4.5 space-y-4 flex flex-col justify-between"
                >
                  <div className="space-y-2">
                    <div className="flex items-start justify-between">
                      <div>
                        <h3 className="font-semibold text-sm text-zinc-100">{p.name}</h3>
                        <p className="text-xs text-zinc-400 font-mono">{p.account_email || 'No email specified'}</p>
                      </div>
                      <span className="px-2 py-0.5 rounded text-[10px] font-mono uppercase bg-zinc-800 text-zinc-300 border border-zinc-700">
                        {p.tier}
                      </span>
                    </div>

                    <div className="text-xs text-zinc-400 space-y-1 font-mono pt-2">
                      <div>CDP Port: <strong className="text-zinc-200">{p.cdp_port}</strong></div>
                      <div className="truncate text-[11px] text-zinc-500">Dir: {p.user_data_dir}</div>
                    </div>
                  </div>

                  <div className="border-t border-zinc-800 pt-3 flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <div className={`w-2 h-2 rounded-full ${p.is_running ? 'bg-emerald-400' : 'bg-zinc-600'}`} />
                      <span className="text-xs text-zinc-400 font-mono">
                        {p.is_running ? 'Running' : 'Stopped'}
                      </span>
                    </div>

                    <div className="flex items-center gap-2">
                      {!p.is_running ? (
                        <button
                          onClick={() => handleLaunchProfile(p.id)}
                          disabled={loadingAction === `launch-${p.id}`}
                          className="flex items-center gap-1.5 px-3 py-1 bg-emerald-600 hover:bg-emerald-500 text-zinc-950 font-medium text-xs rounded transition active:scale-95"
                        >
                          <Play className="w-3 h-3 fill-current" />
                          Launch to Login
                        </button>
                      ) : (
                        <button
                          onClick={() => handleStopProfile(p.id)}
                          className="flex items-center gap-1.5 px-2.5 py-1 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-xs rounded transition"
                        >
                          Stop
                        </button>
                      )}

                      <button
                        onClick={() => handleDeleteProfile(p.id)}
                        className="p-1 text-zinc-500 hover:text-red-400 rounded transition"
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* TAB 3: AGENT AUDIT STREAM */}
        {activeTab === 'activity' && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-base font-semibold text-zinc-100">Live Agent Audit Stream</h2>
                <p className="text-xs text-zinc-400">
                  Every tool call made by Codex, Cursor, or Claude Desktop is captured and measured in real time.
                </p>
              </div>
            </div>

            <div className="bg-zinc-900/60 border border-zinc-800 rounded-xl overflow-hidden shadow-sm">
              <table className="w-full text-left text-xs">
                <thead className="bg-zinc-900 border-b border-zinc-800 text-zinc-400 font-mono">
                  <tr>
                    <th className="py-2.5 px-4 font-medium">Timestamp</th>
                    <th className="py-2.5 px-4 font-medium">Tool Execution</th>
                    <th className="py-2.5 px-4 font-medium">Target Session</th>
                    <th className="py-2.5 px-4 font-medium">Latency</th>
                    <th className="py-2.5 px-4 font-medium">Status</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800/60 font-mono">
                  {activity.length === 0 ? (
                    <tr>
                      <td colSpan={5} className="py-8 text-center text-zinc-500">
                        No agent operations recorded yet. Start interacting with GoLab MCP in Codex.
                      </td>
                    </tr>
                  ) : (
                    activity.map((act) => (
                      <tr key={act.id} className="hover:bg-zinc-800/30 transition-colors">
                        <td className="py-2.5 px-4 text-zinc-400">
                          {new Date(act.created_at).toLocaleTimeString()}
                        </td>
                        <td className="py-2.5 px-4 font-semibold text-emerald-400">
                          {act.tool_name}
                        </td>
                        <td className="py-2.5 px-4 text-zinc-400 truncate max-w-[150px]">
                          {act.session_id || 'active'}
                        </td>
                        <td className="py-2.5 px-4 text-zinc-300">
                          {act.latency_ms}ms
                        </td>
                        <td className="py-2.5 px-4">
                          {act.status === 'success' ? (
                            <span className="px-2 py-0.5 rounded text-[10px] bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
                              OK
                            </span>
                          ) : (
                            <span className="px-2 py-0.5 rounded text-[10px] bg-red-500/10 text-red-400 border border-red-500/30" title={act.error_message}>
                              ERR
                            </span>
                          )}
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </main>

      {/* MODAL: NEW NOTEBOOK */}
      {showNewNotebookModal && (
        <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-zinc-900 border border-zinc-800 rounded-xl max-w-md w-full p-6 space-y-4 shadow-2xl">
            <div className="flex items-center justify-between border-b border-zinc-800 pb-3">
              <h3 className="font-semibold text-sm text-zinc-100">Attach Google Colab Notebook</h3>
              <button
                onClick={() => setShowNewNotebookModal(false)}
                className="text-zinc-500 hover:text-zinc-300"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <form onSubmit={handleCreateSession} className="space-y-4 text-xs">
              <div className="space-y-1">
                <label className="text-zinc-300 font-medium">Colab Notebook URL</label>
                <input
                  type="url"
                  required
                  placeholder="https://colab.research.google.com/drive/..."
                  value={notebookUrl}
                  onChange={(e) => setNotebookUrl(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-md px-3 py-2 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-emerald-500 font-mono"
                />
              </div>

              <div className="space-y-1">
                <label className="text-zinc-300 font-medium">Select Google Account Profile</label>
                <select
                  required
                  value={selectedProfileId}
                  onChange={(e) => setSelectedProfileId(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-md px-3 py-2 text-zinc-100 focus:outline-none focus:border-emerald-500"
                >
                  <option value="">-- Choose Profile --</option>
                  {profiles.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.name} ({p.account_email || 'No email'}) - {p.tier}
                    </option>
                  ))}
                </select>
              </div>

              <div className="space-y-1">
                <label className="text-zinc-300 font-medium">Custom Display Name (Optional)</label>
                <input
                  type="text"
                  placeholder="e.g. LLaMA-FineTune or Dataset-Prep"
                  value={notebookName}
                  onChange={(e) => setNotebookName(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-md px-3 py-2 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-emerald-500"
                />
              </div>

              <div className="flex justify-end gap-2 pt-2 border-t border-zinc-800">
                <button
                  type="button"
                  onClick={() => setShowNewNotebookModal(false)}
                  className="px-3 py-1.5 text-zinc-400 hover:text-zinc-200 transition"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={loadingAction === 'create-session'}
                  className="px-4 py-1.5 bg-emerald-600 hover:bg-emerald-500 text-zinc-950 font-medium rounded transition active:scale-95 disabled:opacity-50"
                >
                  {loadingAction === 'create-session' ? 'Attaching...' : 'Attach & Connect'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL: NEW PROFILE */}
      {showNewProfileModal && (
        <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-zinc-900 border border-zinc-800 rounded-xl max-w-md w-full p-6 space-y-4 shadow-2xl">
            <div className="flex items-center justify-between border-b border-zinc-800 pb-3">
              <h3 className="font-semibold text-sm text-zinc-100">Add Isolated Google Profile</h3>
              <button
                onClick={() => setShowNewProfileModal(false)}
                className="text-zinc-500 hover:text-zinc-300"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <form onSubmit={handleCreateProfile} className="space-y-4 text-xs">
              <div className="space-y-1">
                <label className="text-zinc-300 font-medium">Profile Label</label>
                <input
                  type="text"
                  required
                  placeholder="e.g. Work Colab Pro or Personal"
                  value={profileName}
                  onChange={(e) => setProfileName(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-md px-3 py-2 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-emerald-500"
                />
              </div>

              <div className="space-y-1">
                <label className="text-zinc-300 font-medium">Google Account Email (Optional)</label>
                <input
                  type="email"
                  placeholder="user@example.com"
                  value={profileEmail}
                  onChange={(e) => setProfileEmail(e.target.value)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-md px-3 py-2 text-zinc-100 placeholder:text-zinc-600 focus:outline-none focus:border-emerald-500"
                />
              </div>

              <div className="space-y-1">
                <label className="text-zinc-300 font-medium">Colab Tier</label>
                <select
                  value={profileTier}
                  onChange={(e) => setProfileTier(e.target.value as any)}
                  className="w-full bg-zinc-950 border border-zinc-800 rounded-md px-3 py-2 text-zinc-100 focus:outline-none focus:border-emerald-500"
                >
                  <option value="free">Free Tier (Tesla T4)</option>
                  <option value="pro">Colab Pro (V100 / A100 High-RAM)</option>
                  <option value="pro_plus">Colab Pro+ (Background Execution & Priority)</option>
                </select>
              </div>

              <div className="flex justify-end gap-2 pt-2 border-t border-zinc-800">
                <button
                  type="button"
                  onClick={() => setShowNewProfileModal(false)}
                  className="px-3 py-1.5 text-zinc-400 hover:text-zinc-200 transition"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={loadingAction === 'create-profile'}
                  className="px-4 py-1.5 bg-emerald-600 hover:bg-emerald-500 text-zinc-950 font-medium rounded transition active:scale-95 disabled:opacity-50"
                >
                  Create Profile
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
