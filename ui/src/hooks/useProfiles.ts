import { useState, useCallback } from 'react';
import { Profile } from '../types';
import { api } from '../services/api';

export function useProfiles(onFeedback?: (msg: string) => void) {
  const [profiles, setProfiles] = useState<Profile[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(false);

  const loadProfiles = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await api.getProfiles();
      setProfiles(data);
    } catch (err) {
      console.error('Failed to load profiles:', err);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const createProfile = useCallback(
    async (data: { name: string; account_email?: string; tier: string }) => {
      try {
        await api.createProfile(data);
        await loadProfiles();
        onFeedback?.('Profile created successfully');
      } catch (err: any) {
        alert('Failed to create profile: ' + (err.message || err));
      }
    },
    [loadProfiles, onFeedback]
  );

  const launchProfile = useCallback(
    async (id: string) => {
      try {
        await api.launchProfile(id);
        await loadProfiles();
        onFeedback?.('Chrome browser launched');
      } catch (err: any) {
        alert('Launch failed: ' + (err.message || err));
      }
    },
    [loadProfiles, onFeedback]
  );

  const stopProfile = useCallback(
    async (id: string) => {
      try {
        await api.stopProfile(id);
        await loadProfiles();
        onFeedback?.('Chrome process stopped');
      } catch (err: any) {
        console.error('Stop failed:', err);
      }
    },
    [loadProfiles, onFeedback]
  );

  const deleteProfile = useCallback(
    async (id: string) => {
      try {
        await api.deleteProfile(id);
        await loadProfiles();
        onFeedback?.('Profile deleted');
      } catch (err: any) {
        console.error('Delete failed:', err);
      }
    },
    [loadProfiles, onFeedback]
  );

  return {
    profiles,
    setProfiles,
    isLoading,
    loadProfiles,
    createProfile,
    launchProfile,
    stopProfile,
    deleteProfile,
  };
}
