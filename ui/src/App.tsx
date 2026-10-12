import React, { useState, useEffect } from 'react';
import {
  Layers,
  Globe,
  Activity,
  Plus,
  Square,
  RotateCcw,
  ExternalLink,
  Trash2,
  Play,
  X
} from 'lucide-react';
import { Profile, Session, ActivityLog } from './types';

export default function App() {
  const [activeTab, setActiveTab] = useState<'sessions' | 'profiles' | 'activity'>('sessions');
  const [sessions, setSessions] = useState<Session[]>([]);
  const [profiles, setProfiles] = useState<Profile[]>([]);
  const [activity, setActivity] = useState<ActivityLog[]>([]);
  const [isConnected, setIsConnected] = useState<boolean>(true);

  const [showAttachModal, setShowAttachModal] = useState<boolean>(false);
  const [showNewProfileModal, setShowNewProfileModal] = useState<boolean>(false);

  const [attachUrl, setAttachUrl] = useState('');
  const [attachName, setAttachName] = useState('');
  const [attachProfileId, setAttachProfileId] = useState('');

  const [profName, setProfName] = useState('');
  const [profEmail, setProfEmail] = useState('');
  const [profTier, setProfTier] = useState('free');

  const [statusMessage, setStatusMessage] = useState<string | null>(null);

  const showFeedback = (msg: string) => {
    setStatusMessage(msg);
    setTimeout(() => setStatusMessage(null), 3000);
  };

  const refreshAll = async () => {
    try {
      const [pRes, sRes, aRes] = await Promise.all([
        fetch('/api/profiles').then(r => r.json()),
        fetch('/api/sessions').then(r => r.json()),
        fetch('/api/activity').then(r => r.json())
      ]);
      if (pRes.profiles) {
        setProfiles(pRes.profiles);
        if (pRes.profiles.length > 0 && !attachProfileId) {
          setAttachProfileId(pRes.profiles[0].id);
        }
      }
      if (sRes.sessions) setSessions(sRes.sessions);
      if (aRes.activity) setActivity(aRes.activity);
    } catch (e) {
      console.error('Fetch failed', e);
    }
  };

  useEffect(() => {
    refreshAll();

    const es = new EventSource('/api/events');
    es.onopen = () => setIsConnected(true);
    es.onerror = () => setIsConnected(false);
    es.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data);
        if (payload.type === 'tick') {
          if (payload.sessions) setSessions(payload.sessions);
          if (payload.activity) setActivity(payload.activity);
        }
      } catch {
        // ignore ping
      }
    };

    return () => es.close();
  }, []);

  const handleInterrupt = async (sessionId: string) => {
    try {
      await fetch(`/api/sessions/${sessionId}/interrupt`, { method: 'POST' });
      showFeedback('Interrupt signal sent (SIGINT)');
    } catch (e) {
      alert('Interrupt failed: ' + e);
    }
  };

  const handleRestart = async (sessionId: string) => {
    try {
      await fetch(`/api/sessions/${sessionId}/restart`, { method: 'POST' });
      showFeedback('Kernel restart requested');
    } catch (e) {
      alert('Restart failed: ' + e);
    }
  };

  const handleFocus = async (sessionId: string) => {
    try {
      await fetch(`/api/sessions/${sessionId}/focus`, { method: 'POST' });
      showFeedback('Tab focused');
    } catch (e) {
      console.error(e);
    }
  };

  const handleDeleteSession = async (sessionId: string) => {
    try {
      await fetch(`/api/sessions/${sessionId}`, { method: 'DELETE' });
      refreshAll();
      showFeedback('Session detached');
    } catch (e) {
      console.error(e);
    }
  };

  const handleLaunchProfile = async (profileId: string) => {
    try {
      await fetch(`/api/profiles/${profileId}/launch`, { method: 'POST' });
      refreshAll();
      showFeedback('Chrome launched');
    } catch (e) {
      alert('Launch failed: ' + e);
    }
  };

  const handleStopProfile = async (profileId: string) => {
    try {
      await fetch(`/api/profiles/${profileId}/stop`, { method: 'POST' });
      refreshAll();
      showFeedback('Chrome stopped');
    } catch (e) {
      console.error(e);
    }
  };

  const handleDeleteProfile = async (profileId: string) => {
    try {
      await fetch(`/api/profiles/${profileId}`, { method: 'DELETE' });
      refreshAll();
      showFeedback('Profile deleted');
    } catch (e) {
      console.error(e);
    }
  };

  const handleAttachSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!attachUrl || !attachProfileId) return;
    try {
      const res = await fetch('/api/sessions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          profile_id: attachProfileId,
          notebook_url: attachUrl,
          name: attachName || undefined
        })
      });
      if (!res.ok) {
        const err = await res.json();
        alert('Failed: ' + err.message);
      } else {
        setShowAttachModal(false);
        setAttachUrl('');
        setAttachName('');
        refreshAll();
        showFeedback('Notebook attached');
      }
    } catch (e: any) {
      alert('Error: ' + e.message);
    }
  };

  const handleProfileSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!profName) return;
    try {
      const res = await fetch('/api/profiles', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: profName,
          account_email: profEmail || undefined,
          tier: profTier
        })
      });
      if (!res.ok) {
        const err = await res.json();
        alert('Failed: ' + err.message);
      } else {
        setShowNewProfileModal(false);
        setProfName('');
        setProfEmail('');
        refreshAll();
        showFeedback('Profile created');
      }
    } catch (e: any) {
      alert('Error: ' + e.message);
    }
  };

  return (
    <div className="flex h-screen bg-[#09090b] text-zinc-100 font-sans antialiased overflow-hidden select-none">
      {/* Left Sidebar */}
      <aside className="w-56 flex-shrink-0 bg-[#09090b] border-r border-zinc-800/80 flex flex-col justify-between">
        <div>
          {/* Header */}
          <div className="h-14 px-5 border-b border-zinc-800/80 flex items-center justify-between">
            <div className="flex items-baseline gap-2">
              <span className="text-sm font-semibold tracking-tight text-zinc-100">GoLab</span>
              <span className="text-[11px] font-mono text-zinc-500">control</span>
            </div>
            <div
              className={`w-1.5 h-1.5 rounded-full ${isConnected ? 'bg-emerald-500' : 'bg-zinc-600'}`}
              title={isConnected ? 'Connected' : 'Offline'}
            />
          </div>

          {/* Navigation */}
          <nav className="p-3 space-y-1">
            <button
              onClick={() => setActiveTab('sessions')}
              className={`w-full flex items-center justify-between px-3 py-2 rounded-md text-xs font-medium transition ${
                activeTab === 'sessions'
                  ? 'bg-zinc-800 text-zinc-100'
                  : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900/60'
              }`}
            >
              <div className="flex items-center gap-2.5">
                <Layers className="w-3.5 h-3.5" />
                <span>Sessions</span>
              </div>
              {sessions.length > 0 && (
                <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-zinc-900 text-zinc-400 border border-zinc-800">
                  {sessions.length}
                </span>
              )}
            </button>

            <button
              onClick={() => setActiveTab('profiles')}
              className={`w-full flex items-center justify-between px-3 py-2 rounded-md text-xs font-medium transition ${
                activeTab === 'profiles'
                  ? 'bg-zinc-800 text-zinc-100'
                  : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900/60'
              }`}
            >
              <div className="flex items-center gap-2.5">
                <Globe className="w-3.5 h-3.5" />
                <span>Profiles</span>
              </div>
              {profiles.length > 0 && (
                <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-zinc-900 text-zinc-400 border border-zinc-800">
                  {profiles.length}
                </span>
              )}
            </button>

            <button
              onClick={() => setActiveTab('activity')}
              className={`w-full flex items-center justify-between px-3 py-2 rounded-md text-xs font-medium transition ${
                activeTab === 'activity'
                  ? 'bg-zinc-800 text-zinc-100'
                  : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900/60'
              }`}
            >
              <div className="flex items-center gap-2.5">
                <Activity className="w-3.5 h-3.5" />
                <span>Activity Log</span>
              </div>
              {activity.length > 0 && (
                <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-zinc-900 text-zinc-400 border border-zinc-800">
                  {activity.length}
                </span>
              )}
            </button>
          </nav>
        </div>

        {/* Footer */}
        <div className="p-4 border-t border-zinc-800/80">
          <div className="text-[11px] font-mono text-zinc-500 flex items-center justify-between">
            <span>Daemon</span>
            <span className="text-zinc-400">127.0.0.1:9120</span>
          </div>
        </div>
      </aside>

      {/* Main Container */}
      <div className="flex-1 flex flex-col min-w-0 bg-[#0c0c0e]">
        {/* Header Bar */}
        <header className="h-14 px-6 border-b border-zinc-800/80 flex items-center justify-between bg-[#09090b]/40 flex-shrink-0">
          <div className="flex items-center gap-3">
            <h1 className="text-xs font-medium text-zinc-200 capitalize">
              {activeTab === 'sessions' && 'Active Notebook Sessions'}
              {activeTab === 'profiles' && 'Google Accounts & CDP Profiles'}
              {activeTab === 'activity' && 'Agent Invocation Stream'}
            </h1>
            {statusMessage && (
              <span className="text-[11px] font-mono text-zinc-400 bg-zinc-800 px-2 py-0.5 rounded border border-zinc-700/60">
                {statusMessage}
              </span>
            )}
          </div>

          <div className="flex items-center gap-3">
            {activeTab === 'sessions' && (
              <button
                onClick={() => setShowAttachModal(true)}
                className="flex items-center gap-1.5 px-3 py-1.5 rounded bg-zinc-100 hover:bg-zinc-200 text-zinc-950 text-xs font-medium transition"
              >
                <Plus className="w-3.5 h-3.5" />
                Attach Notebook
              </button>
            )}
            {activeTab === 'profiles' && (
              <button
                onClick={() => setShowNewProfileModal(true)}
                className="flex items-center gap-1.5 px-3 py-1.5 rounded bg-zinc-100 hover:bg-zinc-200 text-zinc-950 text-xs font-medium transition"
              >
                <Plus className="w-3.5 h-3.5" />
                New Profile
              </button>
            )}
          </div>
        </header>

        {/* Content Pane */}
        <main className="flex-1 overflow-y-auto p-6">
          {/* SESSIONS */}
          {activeTab === 'sessions' && (
            <div>
              {sessions.length === 0 ? (
                <div className="border border-dashed border-zinc-800 rounded-lg p-12 text-center max-w-lg mx-auto mt-12">
                  <p className="text-sm font-medium text-zinc-300">No active sessions</p>
                  <p className="text-xs text-zinc-500 mt-1 mb-4">
                    Attach a Colab notebook URL to route kernel commands and inspect execution.
                  </p>
                  <button
                    onClick={() => setShowAttachModal(true)}
                    className="px-3.5 py-1.5 rounded bg-zinc-100 text-zinc-950 text-xs font-medium hover:bg-zinc-200 transition"
                  >
                    Attach Notebook
                  </button>
                </div>
              ) : (
                <div className="space-y-3">
                  {sessions.map((sess) => {
                    const isBusy = sess.kernel_state === 'busy' || sess.kernel_state === 'executing';
                    const prof = profiles.find((p) => p.id === sess.profile_id);
                    return (
                      <div
                        key={sess.id}
                        className="bg-zinc-900/40 border border-zinc-800/80 rounded-lg p-4 flex flex-col gap-3.5"
                      >
                        {/* Top row */}
                        <div className="flex items-center justify-between gap-4">
                          <div className="min-w-0 flex items-center gap-2.5">
                            <span className="text-xs font-semibold text-zinc-200 truncate">
                              {sess.name}
                            </span>
                            <span className="text-[11px] font-mono text-zinc-500 bg-zinc-800/50 px-2 py-0.5 rounded border border-zinc-800">
                              Profile: {prof?.name || sess.profile_id}
                            </span>
                            <a
                              href={sess.notebook_url}
                              target="_blank"
                              rel="noreferrer"
                              className="text-[11px] font-mono text-zinc-500 hover:text-zinc-300 flex items-center gap-1"
                            >
                              <span>Open in Browser</span>
                              <ExternalLink className="w-3 h-3" />
                            </a>
                          </div>

                          {/* Kernel State indicator */}
                          <div className="flex items-center gap-2">
                            <div className="flex items-center gap-1.5 px-2 py-0.5 rounded bg-zinc-950 border border-zinc-800 text-[11px] font-mono">
                              <span
                                className={`w-1.5 h-1.5 rounded-full ${
                                  isBusy
                                    ? 'bg-amber-400 animate-pulse'
                                    : sess.kernel_state === 'idle'
                                    ? 'bg-emerald-500'
                                    : 'bg-zinc-500'
                                }`}
                              />
                              <span className={isBusy ? 'text-amber-300' : 'text-zinc-400'}>
                                {sess.kernel_state || 'unknown'}
                              </span>
                            </div>
                          </div>
                        </div>

                        {/* Telemetry Row (Render real metrics) */}
                        <div className="flex flex-wrap items-center gap-4 text-xs font-mono text-zinc-400 py-2 px-3 bg-zinc-950/60 rounded border border-zinc-800/50">
                          <div>
                            <span className="text-zinc-500 mr-1.5">GPU:</span>
                            <span>{sess.gpu_name ? sess.gpu_name : 'None'}</span>
                            {sess.gpu_memory_free_mb ? (
                              <span className="text-zinc-500 ml-1">({sess.gpu_memory_free_mb} MB free)</span>
                            ) : null}
                          </div>

                          <div className="border-l border-zinc-800 pl-4">
                            <span className="text-zinc-500 mr-1.5">RAM:</span>
                            <span>{sess.ram_free_gb ? `${sess.ram_free_gb.toFixed(1)} GB free` : '—'}</span>
                          </div>

                          <div className="border-l border-zinc-800 pl-4">
                            <span className="text-zinc-500 mr-1.5">Google Drive:</span>
                            <span className={sess.drive_mounted ? 'text-emerald-400' : 'text-zinc-500'}>
                              {sess.drive_mounted ? 'Mounted' : 'Unmounted'}
                            </span>
                          </div>
                        </div>

                        {/* Action buttons */}
                        <div className="flex items-center justify-between pt-1">
                          <div className="flex items-center gap-2">
                            <button
                              onClick={() => handleInterrupt(sess.id)}
                              className="px-2.5 py-1 rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs font-medium border border-zinc-700/60 flex items-center gap-1.5 transition"
                            >
                              <Square className="w-3 h-3 text-amber-400" />
                              Interrupt
                            </button>

                            <button
                              onClick={() => handleRestart(sess.id)}
                              className="px-2.5 py-1 rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs font-medium border border-zinc-700/60 flex items-center gap-1.5 transition"
                            >
                              <RotateCcw className="w-3 h-3" />
                              Restart Kernel
                            </button>

                            <button
                              onClick={() => handleFocus(sess.id)}
                              className="px-2.5 py-1 rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-xs border border-zinc-700/60 transition"
                            >
                              Focus Tab
                            </button>
                          </div>

                          <button
                            onClick={() => handleDeleteSession(sess.id)}
                            title="Detach session"
                            className="p-1.5 text-zinc-500 hover:text-red-400 transition"
                          >
                            <Trash2 className="w-3.5 h-3.5" />
                          </button>
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          )}

          {/* PROFILES */}
          {activeTab === 'profiles' && (
            <div className="border border-zinc-800/80 rounded-lg overflow-hidden bg-zinc-900/30">
              <table className="w-full text-left text-xs font-mono">
                <thead className="bg-zinc-900/80 border-b border-zinc-800 text-zinc-400">
                  <tr>
                    <th className="py-2.5 px-4 font-normal">Profile Name</th>
                    <th className="py-2.5 px-4 font-normal">Account Email</th>
                    <th className="py-2.5 px-4 font-normal">Tier</th>
                    <th className="py-2.5 px-4 font-normal">CDP Port</th>
                    <th className="py-2.5 px-4 font-normal">Chrome Process</th>
                    <th className="py-2.5 px-4 font-normal text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800/60 text-zinc-300">
                  {profiles.map((p) => (
                    <tr key={p.id} className="hover:bg-zinc-900/50 transition">
                      <td className="py-3 px-4 font-sans font-medium text-zinc-200">
                        {p.name}
                      </td>
                      <td className="py-3 px-4 text-zinc-400">
                        {p.account_email || '—'}
                      </td>
                      <td className="py-3 px-4 uppercase text-zinc-400">
                        {p.tier}
                      </td>
                      <td className="py-3 px-4 text-zinc-400">
                        {p.cdp_port}
                      </td>
                      <td className="py-3 px-4">
                        <div className="flex items-center gap-1.5">
                          <span
                            className={`w-1.5 h-1.5 rounded-full ${
                              p.is_running ? 'bg-emerald-500' : 'bg-zinc-600'
                            }`}
                          />
                          <span className={p.is_running ? 'text-emerald-400' : 'text-zinc-500'}>
                            {p.is_running ? 'Running' : 'Stopped'}
                          </span>
                        </div>
                      </td>
                      <td className="py-3 px-4 text-right">
                        <div className="flex items-center justify-end gap-2">
                          {!p.is_running ? (
                            <button
                              onClick={() => handleLaunchProfile(p.id)}
                              className="px-2 py-1 rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs border border-zinc-700/60 flex items-center gap-1 transition"
                            >
                              <Play className="w-3 h-3 text-emerald-400" />
                              Launch Chrome
                            </button>
                          ) : (
                            <button
                              onClick={() => handleStopProfile(p.id)}
                              className="px-2 py-1 rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-400 hover:text-zinc-200 text-xs border border-zinc-700/60 transition"
                            >
                              Stop
                            </button>
                          )}
                          {p.id !== 'default' && (
                            <button
                              onClick={() => handleDeleteProfile(p.id)}
                              className="p-1 text-zinc-500 hover:text-red-400 transition"
                            >
                              <Trash2 className="w-3.5 h-3.5" />
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          {/* ACTIVITY */}
          {activeTab === 'activity' && (
            <div className="border border-zinc-800/80 rounded-lg overflow-hidden bg-zinc-900/30">
              <table className="w-full text-left text-xs font-mono">
                <thead className="bg-zinc-900/80 border-b border-zinc-800 text-zinc-400">
                  <tr>
                    <th className="py-2.5 px-4 font-normal">Timestamp</th>
                    <th className="py-2.5 px-4 font-normal">Tool</th>
                    <th className="py-2.5 px-4 font-normal">Latency</th>
                    <th className="py-2.5 px-4 font-normal">Status</th>
                    <th className="py-2.5 px-4 font-normal">Details</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800/60 text-zinc-300">
                  {activity.length === 0 ? (
                    <tr>
                      <td colSpan={5} className="py-8 text-center text-zinc-500">
                        No tool invocations recorded yet
                      </td>
                    </tr>
                  ) : (
                    activity.map((log) => (
                      <tr key={log.id} className="hover:bg-zinc-900/50 transition">
                        <td className="py-2.5 px-4 text-zinc-500">
                          {new Date(log.created_at).toLocaleTimeString()}
                        </td>
                        <td className="py-2.5 px-4 font-medium text-zinc-200">
                          {log.tool_name}
                        </td>
                        <td className="py-2.5 px-4 text-zinc-400">
                          {log.latency_ms}ms
                        </td>
                        <td className="py-2.5 px-4">
                          <span
                            className={`text-[11px] ${
                              log.status === 'success' ? 'text-emerald-400' : 'text-red-400'
                            }`}
                          >
                            {log.status}
                          </span>
                        </td>
                        <td className="py-2.5 px-4 text-zinc-500 truncate max-w-xs">
                          {log.error_message || '—'}
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          )}
        </main>
      </div>

      {/* Attach Modal */}
      {showAttachModal && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-[#121215] border border-zinc-800 rounded-lg max-w-md w-full p-5 shadow-2xl relative">
            <button
              onClick={() => setShowAttachModal(false)}
              className="absolute top-4 right-4 text-zinc-500 hover:text-zinc-300"
            >
              <X className="w-4 h-4" />
            </button>
            <h3 className="text-sm font-semibold text-zinc-100">Attach Notebook</h3>
            <p className="text-xs text-zinc-500 mt-0.5 mb-4">
              Connect GoLab directly to an open or new Google Colab notebook.
            </p>

            <form onSubmit={handleAttachSubmit} className="space-y-3.5">
              <div>
                <label className="block text-xs text-zinc-400 mb-1">Notebook URL</label>
                <input
                  type="url"
                  required
                  placeholder="https://colab.research.google.com/drive/..."
                  value={attachUrl}
                  onChange={(e) => setAttachUrl(e.target.value)}
                  className="w-full px-3 py-1.5 bg-[#09090b] border border-zinc-800 rounded text-xs text-zinc-200 focus:outline-none focus:border-zinc-600 font-mono"
                />
              </div>

              <div>
                <label className="block text-xs text-zinc-400 mb-1">Name (Optional)</label>
                <input
                  type="text"
                  placeholder="Experiment Model Training"
                  value={attachName}
                  onChange={(e) => setAttachName(e.target.value)}
                  className="w-full px-3 py-1.5 bg-[#09090b] border border-zinc-800 rounded text-xs text-zinc-200 focus:outline-none focus:border-zinc-600"
                />
              </div>

              <div>
                <label className="block text-xs text-zinc-400 mb-1">Google Profile</label>
                <select
                  value={attachProfileId}
                  onChange={(e) => setAttachProfileId(e.target.value)}
                  className="w-full px-3 py-1.5 bg-[#09090b] border border-zinc-800 rounded text-xs text-zinc-200 focus:outline-none focus:border-zinc-600 font-mono"
                >
                  {profiles.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.name} (Port {p.cdp_port})
                    </option>
                  ))}
                </select>
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowAttachModal(false)}
                  className="px-3 py-1.5 text-xs text-zinc-400 hover:text-zinc-200"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-3 py-1.5 rounded bg-zinc-100 text-zinc-950 text-xs font-medium hover:bg-zinc-200 transition"
                >
                  Connect
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Add Profile Modal */}
      {showNewProfileModal && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-[#121215] border border-zinc-800 rounded-lg max-w-md w-full p-5 shadow-2xl relative">
            <button
              onClick={() => setShowNewProfileModal(false)}
              className="absolute top-4 right-4 text-zinc-500 hover:text-zinc-300"
            >
              <X className="w-4 h-4" />
            </button>
            <h3 className="text-sm font-semibold text-zinc-100">Add Chrome Profile</h3>
            <p className="text-xs text-zinc-500 mt-0.5 mb-4">
              Creates an isolated profile folder and assigns an independent CDP debugging port.
            </p>

            <form onSubmit={handleProfileSubmit} className="space-y-3.5">
              <div>
                <label className="block text-xs text-zinc-400 mb-1">Profile Name</label>
                <input
                  type="text"
                  required
                  placeholder="Colab Pro Account"
                  value={profName}
                  onChange={(e) => setProfName(e.target.value)}
                  className="w-full px-3 py-1.5 bg-[#09090b] border border-zinc-800 rounded text-xs text-zinc-200 focus:outline-none focus:border-zinc-600"
                />
              </div>

              <div>
                <label className="block text-xs text-zinc-400 mb-1">Account Email (Optional)</label>
                <input
                  type="email"
                  placeholder="engineer@example.com"
                  value={profEmail}
                  onChange={(e) => setProfEmail(e.target.value)}
                  className="w-full px-3 py-1.5 bg-[#09090b] border border-zinc-800 rounded text-xs text-zinc-200 focus:outline-none focus:border-zinc-600"
                />
              </div>

              <div>
                <label className="block text-xs text-zinc-400 mb-1">Subscription Tier</label>
                <select
                  value={profTier}
                  onChange={(e) => setProfTier(e.target.value)}
                  className="w-full px-3 py-1.5 bg-[#09090b] border border-zinc-800 rounded text-xs text-zinc-200 focus:outline-none focus:border-zinc-600 font-mono"
                >
                  <option value="free">Free</option>
                  <option value="pro">Pro</option>
                  <option value="pro_plus">Pro+</option>
                </select>
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowNewProfileModal(false)}
                  className="px-3 py-1.5 text-xs text-zinc-400 hover:text-zinc-200"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-3 py-1.5 rounded bg-zinc-100 text-zinc-950 text-xs font-medium hover:bg-zinc-200 transition"
                >
                  Save Profile
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
