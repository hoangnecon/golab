export interface Profile {
  id: string;
  name: string;
  account_email?: string;
  tier: 'free' | 'pro' | 'pro_plus';
  user_data_dir: string;
  cdp_port: number;
  is_running: boolean;
  created_at: string;
}

export interface Session {
  id: string;
  name: string;
  profile_id: string;
  notebook_url: string;
  notebook_id?: string;
  tab_id?: string;
  is_active: boolean;
  kernel_state?: string;
  gpu_name?: string;
  gpu_memory_free_mb?: number;
  gpu_memory_total_mb?: number;
  ram_free_gb?: number;
  drive_mounted: boolean;
  created_at: string;
  last_attached_at?: string;
}

export interface ActivityLog {
  id: number;
  session_id?: string;
  tool_name: string;
  latency_ms: number;
  status: 'success' | 'error';
  error_message?: string;
  created_at: string;
}
