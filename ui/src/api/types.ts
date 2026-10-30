export type Tier = 'free' | 'pro' | 'pro_plus';

export interface Profile {
  id: string;
  name: string;
  account_email?: string;
  tier: Tier;
  user_data_dir: string;
  cdp_port: number;
  is_running: boolean;
  created_at: string;
}

export interface Session {
  id: string;
  name: string;
  profile_id: string;
  project_id?: string;
  owner_profile_id?: string;
  notebook_url: string;
  notebook_id?: string;
  tab_id?: string;
  is_active: boolean;
  is_locked?: boolean;
  permission?: 'edit' | 'view' | 'denied' | 'unknown';
  kernel_state?: string;
  gpu_name?: string;
  gpu_memory_free_mb?: number;
  gpu_memory_total_mb?: number;
  ram_free_gb?: number;
  drive_mounted: boolean;
  created_at: string;
  last_attached_at?: string;
}

export interface Project {
  id: string;
  name: string;
  notebook_url: string;
  owner_profile_id: string;
  created_at: string;
  updated_at: string;
}

export interface NewProject {
  name: string;
  notebook_url: string;
  owner_profile_id: string;
}

export interface RunningCellInfo {
  index: number;
  cell_id: string;
  is_running: boolean;
  elapsed_seconds: number;
  execution_count: string;
  stdout: string;
}

export interface RunningCellsReport {
  has_running_cells: boolean;
  total_running: number;
  running_cells: RunningCellInfo[];
}

export interface AcceleratorOption {
  name: string;
  value: string;
  checked: boolean;
  disabled: boolean;
}

export interface RuntimeConfigReport {
  accelerators: AcceleratorOption[];
  high_ram: boolean;
}

export interface ActivityLog {
  id: number;
  session_id?: string;
  tool_name: string;
  latency_ms: number;
  status: 'success' | 'error';
  error_message?: string;
  created_at: string;
  /** False for calls recorded before payload capture existed. */
  has_payload?: boolean;
}

/** A captured tool request or response body. */
export interface Payload {
  /** Parsed JSON. When `truncated`, a string holding the first bytes of the document. */
  body: unknown;
  /** Size of the full JSON document in bytes. */
  bytes: number;
  truncated: boolean;
}

export interface ActivityDetail extends ActivityLog {
  request: Payload | null;
  response: Payload | null;
}

export interface NewProfile {
  name: string;
  account_email?: string;
  tier: Tier;
}

export interface NewSession {
  profile_id: string;
  notebook_url: string;
  name?: string;
}
