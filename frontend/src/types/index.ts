export interface User {
  id: number;
  username: string;
  email: string;
  role: string;
  created_at: string;
  updated_at: string;
}

export interface StorageSystem {
  id: number;
  name: string;
  ip_address: string;
  port: number;
  username: string;
  skip_tls_verify: boolean;
  is_active: boolean;
  connection_status?: 'connected' | 'disconnected' | 'unknown';
  last_connection_check?: string;
  connection_error?: string;
  created_at: string;
  updated_at: string;
}

export interface VolumeGroup {
  id: number;
  storage_system_id: number;
  vg_id: string;
  vg_name: string;
  partition_id?: string;
  partition_name?: string;
  schedule_count: number;
  last_synced_at?: string;
  created_at: string;
}

export interface VolumeGroupWithSystem extends VolumeGroup {
  system_name: string;
  system_ip: string;
  snapshot_count: number;
}

export interface SnapshotSchedule {
  id: number;
  volume_group_id: number;
  name: string;
  cron_expression: string;
  retention_days: number;
  retention_minutes?: number;
  safeguarded: boolean;
  pool_name?: string;
  snapshot_name_pattern: string;
  is_active: boolean;
  last_executed_at?: string;
  next_execution_at?: string;
  created_at: string;
  updated_at: string;
}

export interface ScheduleWithVolumeGroup extends SnapshotSchedule {
  vg_name: string;
  system_name: string;
}

export interface SnapshotExecution {
  id: number;
  schedule_id: number;
  volume_group_id: number;
  snapshot_name?: string;
  execution_time: string;
  status: 'success' | 'failed' | 'pending';
  error_message?: string;
  snapshot_id?: string;
  retention_days: number;
  retention_minutes?: number;
}

export interface ExecutionWithDetails extends SnapshotExecution {
  schedule_name: string;
  vg_name: string;
  system_name: string;
  storage_system_id?: number;
  is_safeguarded?: boolean;
  pool_name?: string;
}

export interface DashboardStats {
  total_systems: number;
  total_volume_groups: number;
  active_schedules: number;
  recent_executions: number;
  successful_executions: number;
  failed_executions: number;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  csrf_token: string;
  user: User;
}

export interface CreateSystemRequest {
  name: string;
  ip_address: string;
  port: number;
  username: string;
  password: string;
  skip_tls_verify: boolean;
}

export interface CreateScheduleRequest {
  name: string;
  cron_expression: string;
  retention_days: number;
  retention_minutes?: number;
  safeguarded: boolean;
  pool_name?: string;
  snapshot_name_pattern?: string;
  is_active: boolean;
}

export interface NTPServer {
  id: number;
  server_address: string;
  is_active: boolean;
  priority: number;
  last_sync_at?: string;
  sync_status?: string;
  time_offset_ms?: number;
  created_at: string;
  updated_at: string;
}

export interface SystemTimeInfo {
  current_time: string;
  timezone: string;
  timezone_offset: string;
  ntp_sync_enabled: boolean;
  ntp_sync_status: string;
  last_ntp_sync?: string;
  system_uptime: string;
  time_drift_ms?: number;
}

export interface Snapshot {
  id: string;
  name: string;
  volume_group_id?: string;
  volume_group_name?: string;
  snapshot_time?: string;
  expiration_time?: string;
  safeguarded?: boolean;
  capacity?: string;
  status?: string;
  // Additional fields from IBM SVC API
  [key: string]: any;
}

export interface AuditLog {
  id: number;
  user_id?: number;
  username: string;
  action: string;
  resource_type: string;
  resource_id?: string;
  resource_name?: string;
  details?: string;
  ip_address?: string;
  user_agent?: string;
  status: 'success' | 'failed';
  error_message?: string;
  created_at: string;
}

export interface AuditLogFilters {
  user_id?: number;
  action?: string;
  resource_type?: string;
  status?: string;
  start_date?: string;
  end_date?: string;
  limit?: number;
  offset?: number;
}

export interface AuditRetentionSettings {
  max_entries: number;
  retention_days: number;
}

// Notification Types
export interface NotificationChannel {
  id: number;
  name: string;
  type: 'email' | 'slack' | 'webhook' | 'snmp';
  is_active: boolean;
  config: string; // JSON string
  created_at: string;
  updated_at: string;
}

export interface AlertRule {
  id: number;
  name: string;
  description?: string;
  is_active: boolean;
  event_type: string;
  conditions?: string; // JSON string
  severity: 'info' | 'warning' | 'error' | 'critical';
  notification_channel_ids: string; // JSON array
  throttle_minutes: number;
  last_triggered_at?: string;
  created_at: string;
  updated_at: string;
}

export interface NotificationHistory {
  id: number;
  alert_rule_id?: number;
  rule_name?: string;
  notification_channel_id: number;
  channel_name?: string;
  event_type: string;
  severity: string;
  message: string;
  event_details?: string; // JSON string
  status: 'sent' | 'failed' | 'throttled' | 'pending';
  error_message?: string;
  sent_at?: string;
  created_at: string;
}

export interface CreateChannelRequest {
  name: string;
  type: 'email' | 'slack' | 'webhook' | 'snmp';
  config: Record<string, any>;
  description?: string;
}

export interface CreateAlertRuleRequest {
  name: string;
  description?: string;
  is_active: boolean;
  event_type: string;
  conditions?: Record<string, any>;
  severity: 'info' | 'warning' | 'error' | 'critical';
  notification_channel_ids: string;
  throttle_minutes: number;
}

export interface NotificationHistoryFilters {
  channel_id?: number;
  status?: string;
  event_type?: string;
  severity?: string;
  from_date?: string;
  to_date?: string;
  start_time?: string;
  end_time?: string;
  limit?: number;
  offset?: number;
}

// 
