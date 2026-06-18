import apiClient from './client';
import type {
  LoginRequest,
  LoginResponse,
  User,
  StorageSystem,
  CreateSystemRequest,
  VolumeGroup,
  Volume,
  ScheduleWithVolumeGroup,
  CreateScheduleRequest,
  ExecutionWithDetails,
  DashboardStats,
  NTPServer,
  SystemTimeInfo,
  Snapshot,
  AuditLogFilters,
  AuditRetentionSettings,
  NotificationChannel,
  AlertRule,
  NotificationHistory,
  CreateChannelRequest,
  CreateAlertRuleRequest,
  NotificationHistoryFilters,
} from '../types';

// Authentication
export const authApi = {
  login: async (data: LoginRequest): Promise<LoginResponse> => {
    const response = await apiClient.post('/auth/login', data);
    return response.data;
  },
  logout: async (): Promise<void> => {
    await apiClient.post('/auth/logout');
  },
  getCurrentUser: async (): Promise<User> => {
    const response = await apiClient.get('/auth/me');
    return response.data;
  },
};

// Users
export const usersApi = {
  list: async (): Promise<User[]> => {
    const response = await apiClient.get('/users');
    return response.data;
  },
  get: async (id: number): Promise<User> => {
    const response = await apiClient.get(`/users/${id}`);
    return response.data;
  },
  create: async (data: { username: string; password: string; email?: string; role: string }): Promise<User> => {
    const response = await apiClient.post('/users', data);
    return response.data;
  },
  update: async (id: number, data: { email?: string; password?: string; role?: string }): Promise<User> => {
    const response = await apiClient.put(`/users/${id}`, data);
    return response.data;
  },
  delete: async (id: number): Promise<{ message: string }> => {
    const response = await apiClient.delete(`/users/${id}`);
    return response.data;
  },
};

// Storage Systems
export const systemsApi = {
  list: async (): Promise<StorageSystem[]> => {
    const response = await apiClient.get('/systems');
    return response.data;
  },
  get: async (id: number): Promise<StorageSystem> => {
    const response = await apiClient.get(`/systems/${id}`);
    return response.data;
  },
  create: async (data: CreateSystemRequest): Promise<{ id: number; message: string }> => {
    const response = await apiClient.post('/systems', data);
    return response.data;
  },
  update: async (id: number, data: Partial<CreateSystemRequest>): Promise<{ message: string }> => {
    const response = await apiClient.put(`/systems/${id}`, data);
    return response.data;
  },
  delete: async (id: number): Promise<{ message: string }> => {
    const response = await apiClient.delete(`/systems/${id}`);
    return response.data;
  },
  test: async (id: number): Promise<{ message: string }> => {
    const response = await apiClient.post(`/systems/${id}/test`);
    return response.data;
  },
  checkHealth: async (): Promise<{ checked_at: string; results: Array<{ system_id: number; system_name: string; status: string; error?: string }> }> => {
    const response = await apiClient.post('/systems/health-check', {}, {
      timeout: 60000, // 60 seconds timeout for health checks
    });
    return response.data;
  },
};

// Volume Groups
export const volumeGroupsApi = {
  listBySystem: async (systemId: number): Promise<VolumeGroup[]> => {
    const response = await apiClient.get(`/systems/${systemId}/volumegroups`);
    return response.data;
  },
  sync: async (systemId: number): Promise<{ message: string; count: number }> => {
    const response = await apiClient.post(`/systems/${systemId}/volumegroups/sync`);
    return response.data;
  },
  get: async (id: number): Promise<VolumeGroup> => {
    const response = await apiClient.get(`/volumegroups/${id}`);
    return response.data;
  },
  listSnapshots: async (id: number): Promise<Snapshot[]> => {
    const response = await apiClient.get(`/volumegroups/${id}/snapshots`);
    return response.data;
  },
  listVolumes: async (id: number): Promise<Volume[]> => {
    const response = await apiClient.get(`/volumegroups/${id}/volumes`);
    return response.data;
  },
};

// Snapshot Schedules
export const schedulesApi = {
  list: async (): Promise<ScheduleWithVolumeGroup[]> => {
    const response = await apiClient.get('/schedules');
    return response.data;
  },
  listByVolumeGroup: async (vgId: number): Promise<ScheduleWithVolumeGroup[]> => {
    const response = await apiClient.get(`/volumegroups/${vgId}/schedules`);
    return response.data;
  },
  get: async (id: number): Promise<ScheduleWithVolumeGroup> => {
    const response = await apiClient.get(`/schedules/${id}`);
    return response.data;
  },
  create: async (vgId: number, data: CreateScheduleRequest): Promise<{ id: number; message: string }> => {
    const response = await apiClient.post(`/volumegroups/${vgId}/schedules`, data);
    return response.data;
  },
  update: async (id: number, data: CreateScheduleRequest): Promise<{ message: string }> => {
    const response = await apiClient.put(`/schedules/${id}`, data);
    return response.data;
  },
  delete: async (id: number): Promise<{ message: string }> => {
    const response = await apiClient.delete(`/schedules/${id}`);
    return response.data;
  },
  execute: async (id: number): Promise<{ message: string }> => {
    const response = await apiClient.post(`/schedules/${id}/execute`);
    return response.data;
  },
};

// Executions
export const executionsApi = {
  list: async (params?: { status?: string; limit?: number }): Promise<ExecutionWithDetails[]> => {
    const response = await apiClient.get('/executions', { params });
    return response.data;
  },
  get: async (id: number): Promise<ExecutionWithDetails> => {
    const response = await apiClient.get(`/executions/${id}`);
    return response.data;
  },
};

// Dashboard
export const dashboardApi = {
  getStats: async (): Promise<DashboardStats> => {
    const response = await apiClient.get('/dashboard/stats');
    return response.data;
  },
};

// 


// NTP Servers
export const ntpApi = {
  listServers: async (): Promise<NTPServer[]> => {
    const response = await apiClient.get('/ntp/servers');
    return response.data;
  },
  createServer: async (data: { server_address: string; is_active: boolean; priority: number }): Promise<{ id: number; message: string }> => {
    const response = await apiClient.post('/ntp/servers', data);
    return response.data;
  },
  updateServer: async (id: number, data: { server_address?: string; is_active?: boolean; priority?: number }): Promise<{ message: string }> => {
    const response = await apiClient.put(`/ntp/servers/${id}`, data);
    return response.data;
  },
  deleteServer: async (id: number): Promise<{ message: string }> => {
    const response = await apiClient.delete(`/ntp/servers/${id}`);
    return response.data;
  },
  syncServer: async (id: number): Promise<{ message: string; sync_status: string; time_offset_ms: number }> => {
    const response = await apiClient.post(`/ntp/servers/${id}/sync`);
    return response.data;
  },
  getSystemTime: async (): Promise<SystemTimeInfo> => {
    const response = await apiClient.get('/ntp/time');
    return response.data;
  },
  setSystemTime: async (time: string): Promise<{ message: string; time: string }> => {
    const response = await apiClient.post('/ntp/time', { time });
    return response.data;
  },
  setTimezone: async (timezone: string): Promise<{ message: string; timezone: string }> => {
    const response = await apiClient.put('/ntp/timezone', { timezone });
    return response.data;
  },

};

// Audit Logs
export const auditLogsApi = {
  list: async (filters?: AuditLogFilters) => {
    const params = new URLSearchParams();
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          params.append(key, value.toString());
        }
      });
    }
    const response = await apiClient.get(`/audit-logs?${params.toString()}`);
    return response.data;
  },
};

// Settings
export const settingsApi = {
  getAuditRetention: async (): Promise<AuditRetentionSettings> => {
    const response = await apiClient.get('/settings/audit-retention');
    return response.data;
  },
  updateAuditRetention: async (settings: AuditRetentionSettings): Promise<AuditRetentionSettings> => {
    const response = await apiClient.put('/settings/audit-retention', settings);
    return response.data;
  },
};

// Notifications
export const notificationsApi = {
  // Channels
  listChannels: async (): Promise<NotificationChannel[]> => {
    const response = await apiClient.get('/notifications/channels');
    return response.data;
  },
  getChannel: async (id: number): Promise<NotificationChannel> => {
    const response = await apiClient.get(`/notifications/channels/${id}`);
    return response.data;
  },
  createChannel: async (data: CreateChannelRequest): Promise<NotificationChannel> => {
    const response = await apiClient.post('/notifications/channels', data);
    return response.data;
  },
  updateChannel: async (id: number, data: Partial<CreateChannelRequest>): Promise<NotificationChannel> => {
    const response = await apiClient.put(`/notifications/channels/${id}`, data);
    return response.data;
  },
  deleteChannel: async (id: number): Promise<void> => {
    await apiClient.delete(`/notifications/channels/${id}`);
  },
  testChannel: async (id: number): Promise<{ message: string }> => {
    const response = await apiClient.post(`/notifications/channels/${id}/test`);
    return response.data;
  },

  // Alert Rules
  listRules: async (): Promise<AlertRule[]> => {
    const response = await apiClient.get('/notifications/rules');
    return response.data;
  },
  getRule: async (id: number): Promise<AlertRule> => {
    const response = await apiClient.get(`/notifications/rules/${id}`);
    return response.data;
  },
  createRule: async (data: CreateAlertRuleRequest): Promise<AlertRule> => {
    const response = await apiClient.post('/notifications/rules', data);
    return response.data;
  },
  updateRule: async (id: number, data: Partial<CreateAlertRuleRequest>): Promise<AlertRule> => {
    const response = await apiClient.put(`/notifications/rules/${id}`, data);
    return response.data;
  },
  deleteRule: async (id: number): Promise<void> => {
    await apiClient.delete(`/notifications/rules/${id}`);
  },

  // History
  listHistory: async (filters?: NotificationHistoryFilters): Promise<NotificationHistory[]> => {
    const params = new URLSearchParams();
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          params.append(key, value.toString());
        }
      });
    }
    const response = await apiClient.get(`/notifications/history?${params.toString()}`);
    return response.data;
  },

  // Test notification
  sendTest: async (channelId: number, message?: string): Promise<{ message: string }> => {
    const response = await apiClient.post('/notifications/test', { channel_id: channelId, message });
    return response.data;
  },
};
