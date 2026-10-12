import React, { useState, useEffect, useCallback } from 'react';
import { AppLayout } from './components/layout/AppLayout';
import { TabType } from './components/layout/Sidebar';
import { SessionsPage } from './features/sessions/SessionsPage';
import { AttachSessionModal } from './features/sessions/AttachSessionModal';
import { ProfilesPage } from './features/profiles/ProfilesPage';
import { CreateProfileModal } from './features/profiles/CreateProfileModal';
import { ActivityPage } from './features/activity/ActivityPage';
import { useSessions } from './hooks/useSessions';
import { useProfiles } from './hooks/useProfiles';
import { useSSE } from './hooks/useSSE';
import { api } from './services/api';
import { ActivityLog } from './types';

export default function App() {
  const [activeTab, setActiveTab] = useState<TabType>('sessions');
  const [statusMessage, setStatusMessage] = useState<string | null>(null);
  const [activity, setActivity] = useState<ActivityLog[]>([]);

  // Modals
  const [showAttachModal, setShowAttachModal] = useState(false);
  const [showProfileModal, setShowProfileModal] = useState(false);

  const showFeedback = useCallback((msg: string) => {
    setStatusMessage(msg);
    const timer = setTimeout(() => setStatusMessage(null), 3000);
    return () => clearTimeout(timer);
  }, []);

  const {
    sessions,
    setSessions,
    loadSessions,
    attachSession,
    interruptSession,
    restartSession,
    focusSession,
    detachSession,
  } = useSessions(showFeedback);

  const {
    profiles,
    setProfiles,
    loadProfiles,
    createProfile,
    launchProfile,
    stopProfile,
    deleteProfile,
  } = useProfiles(showFeedback);

  // SSE Real-time sync handler
  const handleSSETick = useCallback(
    (payload: { type?: string; sessions?: any[]; activity?: ActivityLog[] }) => {
      if (payload.type === 'tick') {
        if (payload.sessions) setSessions(payload.sessions);
        if (payload.activity) setActivity(payload.activity);
      }
    },
    [setSessions]
  );

  const { isConnected } = useSSE(handleSSETick);

  // Initial data loading
  useEffect(() => {
    loadSessions();
    loadProfiles();
    api.getActivity().then(setActivity).catch(console.error);
  }, [loadSessions, loadProfiles]);

  const handlePrimaryAction = () => {
    if (activeTab === 'sessions') {
      setShowAttachModal(true);
    } else if (activeTab === 'profiles') {
      setShowProfileModal(true);
    }
  };

  return (
    <AppLayout
      activeTab={activeTab}
      onTabChange={setActiveTab}
      onPrimaryAction={handlePrimaryAction}
      statusMessage={statusMessage}
      isConnected={isConnected}
      sessionCount={sessions.length}
      profileCount={profiles.length}
      activityCount={activity.length}
    >
      {activeTab === 'sessions' && (
        <SessionsPage
          sessions={sessions}
          profiles={profiles}
          onAttachClick={() => setShowAttachModal(true)}
          onInterrupt={interruptSession}
          onRestart={restartSession}
          onFocus={focusSession}
          onDetach={detachSession}
        />
      )}

      {activeTab === 'profiles' && (
        <ProfilesPage
          profiles={profiles}
          onLaunch={launchProfile}
          onStop={stopProfile}
          onDelete={deleteProfile}
        />
      )}

      {activeTab === 'activity' && <ActivityPage activity={activity} />}

      {/* Attach Session Modal */}
      <AttachSessionModal
        isOpen={showAttachModal}
        onClose={() => setShowAttachModal(false)}
        profiles={profiles}
        defaultProfileId={profiles[0]?.id ?? ''}
        onSubmit={attachSession}
      />

      {/* Create Profile Modal */}
      <CreateProfileModal
        isOpen={showProfileModal}
        onClose={() => setShowProfileModal(false)}
        onSubmit={createProfile}
      />
    </AppLayout>
  );
}
