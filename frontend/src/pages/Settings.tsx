import { useState, useEffect, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { usersApi, ntpApi, settingsApi } from '../api/services';
import type { FormEvent } from 'react';
import type { User, NTPServer } from '../types';

// Constants
const MESSAGE_DISPLAY_DURATION_MS = 5000;
const NTP_REFRESH_INTERVAL_MS = 30000;
const NTP_PRIORITY_MIN = 1;
const NTP_PRIORITY_MAX = 10;
const AUDIT_MAX_ENTRIES_MIN = 100;
const AUDIT_MAX_ENTRIES_MAX = 100000;
const AUDIT_RETENTION_DAYS_MIN = 1;
const AUDIT_RETENTION_DAYS_MAX = 3650;

// Password requirements
const PASSWORD_REQUIREMENTS = {
  minLength: 8,
  maxLength: 128,
  requireUppercase: true,
  requireLowercase: true,
  requireDigit: true,
  requireSpecial: true,
};

// Validation functions
const validateNTPAddress = (address: string): string | null => {
  if (!address || address.trim().length === 0) {
    return 'Server address is required';
  }
  // Basic validation for hostname or IP
  const hostnameRegex = /^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/;
  const ipRegex = /^(\d{1,3}\.){3}\d{1,3}$/;
  
  if (!hostnameRegex.test(address) && !ipRegex.test(address)) {
    return 'Invalid server address format';
  }
  return null;
};

const validateNTPPriority = (priority: number): string | null => {
  if (isNaN(priority) || priority < NTP_PRIORITY_MIN || priority > NTP_PRIORITY_MAX) {
    return `Priority must be between ${NTP_PRIORITY_MIN} and ${NTP_PRIORITY_MAX}`;
  }
  return null;
};

const validateAuditMaxEntries = (value: number): string | null => {
  if (isNaN(value) || value < AUDIT_MAX_ENTRIES_MIN || value > AUDIT_MAX_ENTRIES_MAX) {
    return `Maximum entries must be between ${AUDIT_MAX_ENTRIES_MIN} and ${AUDIT_MAX_ENTRIES_MAX}`;
  }
  return null;
};

const validateAuditRetentionDays = (value: number): string | null => {
  if (isNaN(value) || value < AUDIT_RETENTION_DAYS_MIN || value > AUDIT_RETENTION_DAYS_MAX) {
    return `Retention days must be between ${AUDIT_RETENTION_DAYS_MIN} and ${AUDIT_RETENTION_DAYS_MAX}`;
  }
  return null;
};

// Styles as constants to prevent recreation on each render
const styles = {
  card: {
    backgroundColor: 'white',
    padding: '1.5rem',
    borderRadius: '8px',
    boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
  },
  cardLarge: {
    backgroundColor: 'white',
    padding: '2rem',
    borderRadius: '8px',
    boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
  },
  input: {
    width: '100%',
    padding: '0.5rem',
    border: '1px solid #ddd',
    borderRadius: '4px',
    fontSize: '1rem',
  },
  button: {
    padding: '0.5rem 1rem',
    color: 'white',
    border: 'none',
    borderRadius: '4px',
    cursor: 'pointer',
    fontSize: '0.875rem',
  },
  buttonLarge: {
    padding: '0.75rem 1.5rem',
    color: 'white',
    border: 'none',
    borderRadius: '4px',
    cursor: 'pointer',
    fontWeight: 'bold' as const,
  },
};

// Confirmation Dialog Component
function ConfirmDialog({ 
  isOpen, 
  title, 
  message, 
  onConfirm, 
  onCancel 
}: { 
  isOpen: boolean; 
  title: string; 
  message: string; 
  onConfirm: () => void; 
  onCancel: () => void;
}) {
  if (!isOpen) return null;

  return (
    <div style={{
      position: 'fixed',
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      backgroundColor: 'rgba(0,0,0,0.5)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      zIndex: 1000,
    }}>
      <div style={{
        backgroundColor: 'white',
        padding: '2rem',
        borderRadius: '8px',
        maxWidth: '500px',
        width: '90%',
      }}>
        <h3 style={{ marginTop: 0 }}>{title}</h3>
        <p>{message}</p>
        <div style={{ display: 'flex', gap: '1rem', justifyContent: 'flex-end', marginTop: '1.5rem' }}>
          <button
            onClick={onCancel}
            style={{
              ...styles.button,
              backgroundColor: '#6c757d',
            }}
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            style={{
              ...styles.button,
              backgroundColor: '#dc3545',
            }}
          >
            Confirm
          </button>
        </div>
      </div>
    </div>
  );
}

// Clock & NTP Tab Component
function ClockNTPTab() {
  const queryClient = useQueryClient();
  const [showNTPForm, setShowNTPForm] = useState(false);
  const [editingNTP, setEditingNTP] = useState<NTPServer | null>(null);
  const [validationError, setValidationError] = useState<string | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<{ isOpen: boolean; server: NTPServer | null }>({
    isOpen: false,
    server: null,
  });
  const [ntpFormData, setNTPFormData] = useState({
    server_address: '',
    is_active: true,
    priority: 1,
  });

  const { data: timeInfo, isLoading: timeLoading } = useQuery({
    queryKey: ['systemTime'],
    queryFn: ntpApi.getSystemTime,
    refetchInterval: NTP_REFRESH_INTERVAL_MS,
  });

  const { data: ntpServers, isLoading: ntpLoading } = useQuery({
    queryKey: ['ntpServers'],
    queryFn: ntpApi.listServers,
  });

  const createNTPMutation = useMutation({
    mutationFn: (data: typeof ntpFormData) => {
      if (editingNTP) {
        return ntpApi.updateServer(editingNTP.id, data);
      }
      return ntpApi.createServer(data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ntpServers'] });
      setShowNTPForm(false);
      setNTPFormData({ server_address: '', is_active: true, priority: 1 });
      setEditingNTP(null);
      setValidationError(null);
    },
  });

  const deleteNTPMutation = useMutation({
    mutationFn: ntpApi.deleteServer,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ntpServers'] });
      setConfirmDelete({ isOpen: false, server: null });
    },
  });

  const syncNTPMutation = useMutation({
    mutationFn: ntpApi.syncServer,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ntpServers'] });
      queryClient.invalidateQueries({ queryKey: ['systemTime'] });
    },
  });

  const handleNTPSubmit = (e: FormEvent) => {
    e.preventDefault();
    
    // Validate inputs
    const addressError = validateNTPAddress(ntpFormData.server_address);
    if (addressError) {
      setValidationError(addressError);
      return;
    }

    const priorityError = validateNTPPriority(ntpFormData.priority);
    if (priorityError) {
      setValidationError(priorityError);
      return;
    }

    setValidationError(null);
    createNTPMutation.mutate(ntpFormData);
  };

  const handleEditNTP = (server: NTPServer) => {
    setEditingNTP(server);
    setNTPFormData({
      server_address: server.server_address,
      is_active: server.is_active,
      priority: server.priority,
    });
    setShowNTPForm(true);
    setValidationError(null);
  };

  const handleCancelNTPEdit = () => {
    setEditingNTP(null);
    setShowNTPForm(false);
    setNTPFormData({ server_address: '', is_active: true, priority: 1 });
    setValidationError(null);
  };

  const handleDeleteClick = (server: NTPServer) => {
    setConfirmDelete({ isOpen: true, server });
  };

  const handleConfirmDelete = () => {
    if (confirmDelete.server) {
      deleteNTPMutation.mutate(confirmDelete.server.id);
    }
  };

  return (
    <div>
      <h2>Clock & NTP Configuration</h2>

      <div style={{ ...styles.card, marginBottom: '2rem' }}>
        <h3 style={{ marginTop: 0 }}>System Time Information</h3>
        {timeLoading ? (
          <p>Loading time information...</p>
        ) : timeInfo ? (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '1rem' }}>
            <div>
              <strong>Current Time:</strong>
              <div style={{ marginTop: '0.5rem', fontSize: '1.1rem' }}>
                {new Date(timeInfo.current_time).toLocaleString()}
              </div>
            </div>
            <div>
              <strong>Timezone:</strong>
              <div style={{ marginTop: '0.5rem' }}>
                {timeInfo.timezone} ({timeInfo.timezone_offset})
              </div>
            </div>
            <div>
              <strong>NTP Sync Status:</strong>
              <div style={{ marginTop: '0.5rem' }}>
                <span style={{
                  padding: '0.25rem 0.75rem',
                  borderRadius: '12px',
                  backgroundColor: timeInfo.ntp_sync_enabled ? '#d4edda' : '#f8d7da',
                  color: timeInfo.ntp_sync_enabled ? '#155724' : '#721c24',
                  fontSize: '0.875rem',
                }}>
                  {timeInfo.ntp_sync_status}
                </span>
              </div>
            </div>
            {timeInfo.time_drift_ms !== undefined && (
              <div>
                <strong>Time Drift:</strong>
                <div style={{ marginTop: '0.5rem' }}>
                  {timeInfo.time_drift_ms}ms
                </div>
              </div>
            )}
            {timeInfo.last_ntp_sync && (
              <div>
                <strong>Last NTP Sync:</strong>
                <div style={{ marginTop: '0.5rem' }}>
                  {new Date(timeInfo.last_ntp_sync).toLocaleString()}
                </div>
              </div>
            )}
          </div>
        ) : (
          <p>Unable to load time information</p>
        )}
      </div>

      {/* Manual Time & Timezone Setting */}
      <div style={{ ...styles.card, marginBottom: '2rem' }}>
        <h3 style={{ marginTop: 0 }}>Manual Time & Timezone Configuration</h3>
        <p style={{ color: '#666', fontSize: '0.875rem', marginBottom: '1.5rem' }}>
          Note: Time changes are stored for reference but may not affect actual system time in containerized environments.
        </p>
        
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: '2rem' }}>
          {/* Set Time Form */}
          <div>
            <h4 style={{ marginTop: 0, marginBottom: '1rem' }}>Set System Time</h4>
            <form onSubmit={async (e: FormEvent<HTMLFormElement>) => {
              e.preventDefault();
              const formData = new FormData(e.currentTarget);
              const timeValue = formData.get('manual_time') as string;
              
              if (!timeValue) {
                alert('Please enter a time');
                return;
              }
              
              try {
                // Convert to ISO 8601 format
                const isoTime = new Date(timeValue).toISOString();
                const data = await ntpApi.setSystemTime(isoTime);
                alert(data.message || 'Time updated successfully');
                queryClient.invalidateQueries({ queryKey: ['systemTime'] });
              } catch (err) {
                console.error('Error setting time:', err);
                alert('Failed to set time');
              }
            }}>
              <div style={{ marginBottom: '1rem' }}>
                <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>
                  Date & Time:
                </label>
                <input
                  type="datetime-local"
                  name="manual_time"
                  style={styles.input}
                  required
                />
              </div>
              <button
                type="submit"
                style={{
                  ...styles.button,
                  backgroundColor: '#0066cc',
                  width: '100%',
                }}
              >
                Set Time
              </button>
            </form>
          </div>

          {/* Set Timezone Form */}
          <div>
            <h4 style={{ marginTop: 0, marginBottom: '1rem' }}>Set Timezone</h4>
            <form onSubmit={async (e: FormEvent<HTMLFormElement>) => {
              e.preventDefault();
              const formData = new FormData(e.currentTarget);
              const timezone = formData.get('timezone') as string;
              
              if (!timezone) {
                alert('Please select a timezone');
                return;
              }
              
              try {
                const data = await ntpApi.setTimezone(timezone);
                alert(data.message || 'Timezone updated successfully');
                queryClient.invalidateQueries({ queryKey: ['systemTime'] });
              } catch (err) {
                console.error('Error setting timezone:', err);
                alert('Failed to set timezone');
              }
            }}>
              <div style={{ marginBottom: '1rem' }}>
                <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>
                  Timezone:
                </label>
                <select
                  name="timezone"
                  style={styles.input}
                  required
                  defaultValue="UTC"
                >
                  <option value="UTC">UTC</option>
                  <option value="Europe/Oslo">Europe/Oslo (CET/CEST)</option>
                  <option value="Europe/London">Europe/London (GMT/BST)</option>
                  <option value="Europe/Paris">Europe/Paris (CET/CEST)</option>
                  <option value="America/New_York">America/New_York (EST/EDT)</option>
                  <option value="America/Chicago">America/Chicago (CST/CDT)</option>
                  <option value="America/Denver">America/Denver (MST/MDT)</option>
                  <option value="America/Los_Angeles">America/Los_Angeles (PST/PDT)</option>
                  <option value="Asia/Tokyo">Asia/Tokyo (JST)</option>
                  <option value="Asia/Shanghai">Asia/Shanghai (CST)</option>
                  <option value="Asia/Singapore">Asia/Singapore (SGT)</option>
                  <option value="Australia/Sydney">Australia/Sydney (AEDT/AEST)</option>
                </select>
              </div>
              <button
                type="submit"
                style={{
                  ...styles.button,
                  backgroundColor: '#0066cc',
                  width: '100%',
                }}
              >
                Set Timezone
              </button>
            </form>
          </div>
        </div>
      </div>



      <div style={styles.card}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
          <h3 style={{ margin: 0 }}>NTP Servers</h3>
          <button
            onClick={() => setShowNTPForm(!showNTPForm)}
            style={{
              ...styles.button,
              backgroundColor: '#0066cc',
            }}
          >
            {showNTPForm ? 'Cancel' : '+ Add NTP Server'}
          </button>
        </div>

        {showNTPForm && (
          <form onSubmit={handleNTPSubmit} style={{
            padding: '1.5rem',
            backgroundColor: '#f8f9fa',
            borderRadius: '4px',
            marginBottom: '1.5rem',
          }}>
            <h4 style={{ marginTop: 0 }}>{editingNTP ? 'Edit NTP Server' : 'Add NTP Server'}</h4>
            
            {validationError && (
              <div style={{
                padding: '0.75rem',
                marginBottom: '1rem',
                backgroundColor: '#f8d7da',
                color: '#721c24',
                border: '1px solid #f5c6cb',
                borderRadius: '4px',
              }}>
                {validationError}
              </div>
            )}

            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>
                Server Address *
              </label>
              <input
                type="text"
                value={ntpFormData.server_address}
                onChange={(e) => {
                  setNTPFormData({ ...ntpFormData, server_address: e.target.value });
                  setValidationError(null);
                }}
                placeholder="e.g., pool.ntp.org or 192.168.1.1"
                required
                style={styles.input}
              />
            </div>
            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>
                Priority
              </label>
              <input
                type="number"
                value={ntpFormData.priority}
                onChange={(e) => {
                  setNTPFormData({ ...ntpFormData, priority: parseInt(e.target.value, 10) });
                  setValidationError(null);
                }}
                min={NTP_PRIORITY_MIN}
                max={NTP_PRIORITY_MAX}
                required
                style={styles.input}
              />
              <small style={{ color: '#666' }}>Lower numbers have higher priority (1 = highest)</small>
            </div>
            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'flex', alignItems: 'center', cursor: 'pointer' }}>
                <input
                  type="checkbox"
                  checked={ntpFormData.is_active}
                  onChange={(e) => setNTPFormData({ ...ntpFormData, is_active: e.target.checked })}
                  style={{ marginRight: '0.5rem' }}
                />
                <span>Active</span>
              </label>
            </div>
            <div style={{ display: 'flex', gap: '0.5rem' }}>
              <button
                type="submit"
                disabled={createNTPMutation.isPending}
                style={{
                  ...styles.button,
                  backgroundColor: '#28a745',
                  padding: '0.5rem 1.5rem',
                }}
              >
                {createNTPMutation.isPending ? 'Saving...' : editingNTP ? 'Update' : 'Add Server'}
              </button>
              <button
                type="button"
                onClick={handleCancelNTPEdit}
                style={{
                  ...styles.button,
                  backgroundColor: '#6c757d',
                  padding: '0.5rem 1.5rem',
                }}
              >
                Cancel
              </button>
            </div>
          </form>
        )}

        {ntpLoading ? (
          <p>Loading NTP servers...</p>
        ) : ntpServers && ntpServers.length > 0 ? (
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ backgroundColor: '#f8f9fa', borderBottom: '2px solid #dee2e6' }}>
                <th style={{ padding: '0.75rem', textAlign: 'left' }}>Server Address</th>
                <th style={{ padding: '0.75rem', textAlign: 'left' }}>Priority</th>
                <th style={{ padding: '0.75rem', textAlign: 'left' }}>Status</th>
                <th style={{ padding: '0.75rem', textAlign: 'left' }}>Last Sync</th>
                <th style={{ padding: '0.75rem', textAlign: 'left' }}>Offset</th>
                <th style={{ padding: '0.75rem', textAlign: 'left' }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {ntpServers.map((server) => (
                <tr key={server.id} style={{ borderBottom: '1px solid #dee2e6' }}>
                  <td style={{ padding: '0.75rem' }}>{server.server_address}</td>
                  <td style={{ padding: '0.75rem' }}>{server.priority}</td>
                  <td style={{ padding: '0.75rem' }}>
                    <span style={{
                      padding: '0.25rem 0.75rem',
                      borderRadius: '12px',
                      backgroundColor: server.is_active ? '#d4edda' : '#f8d7da',
                      color: server.is_active ? '#155724' : '#721c24',
                      fontSize: '0.875rem',
                    }}>
                      {server.is_active ? 'Active' : 'Inactive'}
                    </span>
                  </td>
                  <td style={{ padding: '0.75rem' }}>
                    {server.last_sync_at ? new Date(server.last_sync_at).toLocaleString() : 'Never'}
                  </td>
                  <td style={{ padding: '0.75rem' }}>
                    {server.time_offset_ms !== undefined ? `${server.time_offset_ms}ms` : '-'}
                  </td>
                  <td style={{ padding: '0.75rem' }}>
                    <button
                      onClick={() => syncNTPMutation.mutate(server.id)}
                      disabled={syncNTPMutation.isPending}
                      style={{
                        ...styles.button,
                        backgroundColor: '#17a2b8',
                        marginRight: '0.5rem',
                      }}
                    >
                      Sync
                    </button>
                    <button
                      onClick={() => handleEditNTP(server)}
                      style={{
                        ...styles.button,
                        backgroundColor: '#0066cc',
                        marginRight: '0.5rem',
                      }}
                    >
                      Edit
                    </button>
                    <button
                      onClick={() => handleDeleteClick(server)}
                      disabled={deleteNTPMutation.isPending}
                      style={{
                        ...styles.button,
                        backgroundColor: '#dc3545',
                      }}
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : (
          <div style={{
            padding: '2rem',
            textAlign: 'center',
            color: '#666',
            backgroundColor: '#f8f9fa',
            borderRadius: '4px',
          }}>
            <p style={{ margin: 0 }}>No NTP servers configured.</p>
            <p style={{ marginTop: '0.5rem', fontSize: '0.9rem' }}>
              Click "Add NTP Server" to configure time synchronization.
            </p>
          </div>
        )}
      </div>

      <ConfirmDialog
        isOpen={confirmDelete.isOpen}
        title="Delete NTP Server"
        message={`Are you sure you want to delete NTP server "${confirmDelete.server?.server_address}"?`}
        onConfirm={handleConfirmDelete}
        onCancel={() => setConfirmDelete({ isOpen: false, server: null })}
      />
    </div>
  );
}

// Audit Logs Tab Component
function AuditLogsTab() {
  const queryClient = useQueryClient();
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const [validationError, setValidationError] = useState<string | null>(null);
  const [formData, setFormData] = useState({
    max_entries: 1000,
    retention_days: 365,
  });

  const { data: settings, isLoading } = useQuery({
    queryKey: ['auditRetentionSettings'],
    queryFn: settingsApi.getAuditRetention,
  });

  useEffect(() => {
    if (settings) {
      setFormData({
        max_entries: settings.max_entries,
        retention_days: settings.retention_days,
      });
    }
  }, [settings]);

  const updateSettingsMutation = useMutation({
    mutationFn: settingsApi.updateAuditRetention,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['auditRetentionSettings'] });
      setMessage({ type: 'success', text: 'Audit retention settings updated successfully!' });
      setTimeout(() => setMessage(null), MESSAGE_DISPLAY_DURATION_MS);
    },
    onError: (error: any) => {
      setMessage({ type: 'error', text: error.message || 'Failed to update settings' });
      setTimeout(() => setMessage(null), MESSAGE_DISPLAY_DURATION_MS);
    },
  });

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    
    // Validate inputs
    const maxEntriesError = validateAuditMaxEntries(formData.max_entries);
    if (maxEntriesError) {
      setValidationError(maxEntriesError);
      return;
    }

    const retentionDaysError = validateAuditRetentionDays(formData.retention_days);
    if (retentionDaysError) {
      setValidationError(retentionDaysError);
      return;
    }

    setValidationError(null);
    updateSettingsMutation.mutate(formData);
  };

  return (
    <div>
      <h2>Audit Log Retention Settings</h2>

      {message && (
        <div style={{
          padding: '1rem',
          marginBottom: '1rem',
          borderRadius: '4px',
          backgroundColor: message.type === 'success' ? '#d4edda' : '#f8d7da',
          color: message.type === 'success' ? '#155724' : '#721c24',
          border: `1px solid ${message.type === 'success' ? '#c3e6cb' : '#f5c6cb'}`,
        }}>
          {message.text}
        </div>
      )}

      <div style={styles.cardLarge}>
        {isLoading ? (
          <p>Loading settings...</p>
        ) : (
          <form onSubmit={handleSubmit}>
            <div style={{ marginBottom: '2rem' }}>
              <p style={{ color: '#666', marginBottom: '1.5rem' }}>
                Configure how long audit logs are retained. The system will automatically clean up old logs based on these settings.
              </p>

              {validationError && (
                <div style={{
                  padding: '0.75rem',
                  marginBottom: '1rem',
                  backgroundColor: '#f8d7da',
                  color: '#721c24',
                  border: '1px solid #f5c6cb',
                  borderRadius: '4px',
                }}>
                  {validationError}
                </div>
              )}

              <div style={{ marginBottom: '1.5rem' }}>
                <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
                  Maximum Entries
                </label>
                <input
                  type="number"
                  value={formData.max_entries}
                  onChange={(e) => {
                    setFormData({ ...formData, max_entries: parseInt(e.target.value, 10) });
                    setValidationError(null);
                  }}
                  min={AUDIT_MAX_ENTRIES_MIN}
                  max={AUDIT_MAX_ENTRIES_MAX}
                  required
                  style={styles.input}
                />
                <small style={{ color: '#666', display: 'block', marginTop: '0.5rem' }}>
                  Maximum number of audit log entries to keep ({AUDIT_MAX_ENTRIES_MIN}-{AUDIT_MAX_ENTRIES_MAX.toLocaleString()})
                </small>
              </div>

              <div style={{ marginBottom: '1.5rem' }}>
                <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
                  Retention Days
                </label>
                <input
                  type="number"
                  value={formData.retention_days}
                  onChange={(e) => {
                    setFormData({ ...formData, retention_days: parseInt(e.target.value, 10) });
                    setValidationError(null);
                  }}
                  min={AUDIT_RETENTION_DAYS_MIN}
                  max={AUDIT_RETENTION_DAYS_MAX}
                  required
                  style={styles.input}
                />
                <small style={{ color: '#666', display: 'block', marginTop: '0.5rem' }}>
                  Number of days to retain audit logs ({AUDIT_RETENTION_DAYS_MIN}-{AUDIT_RETENTION_DAYS_MAX} days / ~10 years)
                </small>
              </div>

              <div style={{
                padding: '1rem',
                backgroundColor: '#fff3cd',
                border: '1px solid #ffc107',
                borderRadius: '4px',
                marginBottom: '1.5rem',
              }}>
                <strong>⚠️ Note:</strong> Logs will be deleted if they exceed EITHER the maximum entries OR the retention days limit, whichever is more restrictive.
              </div>
            </div>

            <button
              type="submit"
              disabled={updateSettingsMutation.isPending}
              style={{
                ...styles.buttonLarge,
                backgroundColor: updateSettingsMutation.isPending ? '#ccc' : '#0066cc',
                cursor: updateSettingsMutation.isPending ? 'not-allowed' : 'pointer',
                padding: '0.75rem 2rem',
                fontSize: '1rem',
              }}
            >
              {updateSettingsMutation.isPending ? 'Saving...' : 'Save Settings'}
            </button>
          </form>
        )}
      </div>
    </div>
  );
}

export default function Settings() {
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState<'users' | 'clock' | 'audit'>('users');
  const [showUserForm, setShowUserForm] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<{ isOpen: boolean; user: User | null }>({
    isOpen: false,
    user: null,
  });
  const [formData, setFormData] = useState({
    username: '',
    password: '',
    role: 'viewer',
  });

  const { data: users, isLoading } = useQuery({
    queryKey: ['users'],
    queryFn: usersApi.list,
  });

  const createUserMutation = useMutation({
    mutationFn: (data: typeof formData) => {
      if (editingUser) {
        return usersApi.update(editingUser.id, {
          email: data.username,
          password: data.password || undefined,
          role: data.role,
        });
      }
      return usersApi.create(data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setShowUserForm(false);
      setFormData({ username: '', password: '', role: 'viewer' });
      setMessage({ type: 'success', text: 'User created successfully!' });
      setTimeout(() => setMessage(null), MESSAGE_DISPLAY_DURATION_MS);
    },
    onError: (error: any) => {
      setMessage({ type: 'error', text: error.message });
      setTimeout(() => setMessage(null), MESSAGE_DISPLAY_DURATION_MS);
    },
  });

  const deleteUserMutation = useMutation({
    mutationFn: usersApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setMessage({ type: 'success', text: 'User deleted successfully!' });
      setTimeout(() => setMessage(null), MESSAGE_DISPLAY_DURATION_MS);
      setConfirmDelete({ isOpen: false, user: null });
    },
    onError: (error: any) => {
      setMessage({ type: 'error', text: error.message });
      setTimeout(() => setMessage(null), MESSAGE_DISPLAY_DURATION_MS);
    },
  });

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    createUserMutation.mutate(formData);
  };

  const handleEdit = (user: User) => {
    setEditingUser(user);
    setFormData({
      username: user.username,
      password: '',
      role: user.role,
    });
    setShowUserForm(true);
  };

  const handleCancelEdit = () => {
    setEditingUser(null);
    setShowUserForm(false);
    setFormData({ username: '', password: '', role: 'viewer' });
  };

  const handleDeleteClick = (user: User) => {
    setConfirmDelete({ isOpen: true, user });
  };

  const handleConfirmDelete = () => {
    if (confirmDelete.user) {
      deleteUserMutation.mutate(confirmDelete.user.id);
    }
  };

  const passwordRequirementsText = useMemo(() => {
    const reqs = [];
    reqs.push(`${PASSWORD_REQUIREMENTS.minLength}+ characters`);
    if (PASSWORD_REQUIREMENTS.requireUppercase) reqs.push('uppercase letter');
    if (PASSWORD_REQUIREMENTS.requireLowercase) reqs.push('lowercase letter');
    if (PASSWORD_REQUIREMENTS.requireDigit) reqs.push('digit');
    if (PASSWORD_REQUIREMENTS.requireSpecial) reqs.push('special character');
    return `Must contain: ${reqs.join(', ')}`;
  }, []);

  return (
    <div style={{ padding: '2rem' }}>
      <h1>Settings</h1>

      {message && (
        <div style={{
          padding: '1rem',
          marginBottom: '1rem',
          borderRadius: '4px',
          backgroundColor: message.type === 'success' ? '#d4edda' : '#f8d7da',
          color: message.type === 'success' ? '#155724' : '#721c24',
          border: `1px solid ${message.type === 'success' ? '#c3e6cb' : '#f5c6cb'}`,
        }}>
          {message.text}
        </div>
      )}

      <div style={{
        display: 'flex',
        gap: '0.5rem',
        marginBottom: '2rem',
        borderBottom: '2px solid #e0e0e0',
      }}>
        <button
          onClick={() => setActiveTab('users')}
          style={{
            padding: '0.75rem 1.5rem',
            backgroundColor: 'transparent',
            color: activeTab === 'users' ? '#0066cc' : '#666',
            border: 'none',
            borderBottom: activeTab === 'users' ? '3px solid #0066cc' : '3px solid transparent',
            cursor: 'pointer',
            fontWeight: activeTab === 'users' ? 'bold' : 'normal',
            fontSize: '1rem',
            transition: 'all 0.2s',
          }}
        >
          👥 User Management
        </button>
        <button
          onClick={() => setActiveTab('clock')}
          style={{
            padding: '0.75rem 1.5rem',
            backgroundColor: 'transparent',
            color: activeTab === 'clock' ? '#0066cc' : '#666',
            border: 'none',
            borderBottom: activeTab === 'clock' ? '3px solid #0066cc' : '3px solid transparent',
            cursor: 'pointer',
            fontWeight: activeTab === 'clock' ? 'bold' : 'normal',
            fontSize: '1rem',
            transition: 'all 0.2s',
          }}
        >
          🕐 Clock & NTP
        </button>
        <button
          onClick={() => setActiveTab('audit')}
          style={{
            padding: '0.75rem 1.5rem',
            backgroundColor: 'transparent',
            color: activeTab === 'audit' ? '#0066cc' : '#666',
            border: 'none',
            borderBottom: activeTab === 'audit' ? '3px solid #0066cc' : '3px solid transparent',
            cursor: 'pointer',
            fontWeight: activeTab === 'audit' ? 'bold' : 'normal',
            fontSize: '1rem',
            transition: 'all 0.2s',
          }}
        >
          📋 Audit Logs
        </button>
      </div>

      {activeTab === 'users' && (
        <div>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
            <h2 style={{ margin: 0 }}>Users</h2>
            <button
              onClick={() => {
                setShowUserForm(!showUserForm);
                if (showUserForm) handleCancelEdit();
              }}
              style={{
                ...styles.buttonLarge,
                backgroundColor: '#0066cc',
              }}
            >
              {showUserForm ? 'Cancel' : '+ Add User'}
            </button>
          </div>

          {showUserForm && (
            <div style={{ ...styles.cardLarge, marginBottom: '2rem' }}>
              <h3 style={{ marginTop: 0 }}>{editingUser ? 'Edit User' : 'Add New User'}</h3>
              <form onSubmit={handleSubmit}>
                <div style={{ display: 'grid', gap: '1rem', marginBottom: '1.5rem' }}>
                  <div>
                    <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
                      Username *
                    </label>
                    <input
                      type="text"
                      value={formData.username}
                      onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                      required
                      disabled={!!editingUser}
                      style={{
                        ...styles.input,
                        backgroundColor: editingUser ? '#f5f5f5' : 'white',
                      }}
                    />
                    {editingUser && (
                      <small style={{ color: '#666', fontSize: '0.875rem' }}>
                        Username cannot be changed
                      </small>
                    )}
                  </div>
                  <div>
                    <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
                      Password {editingUser ? '(leave blank to keep current)' : '*'}
                    </label>
                    <input
                      type="password"
                      value={formData.password}
                      onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                      required={!editingUser}
                      style={styles.input}
                    />
                    {!editingUser && (
                      <small style={{ color: '#666', display: 'block', marginTop: '0.5rem' }}>
                        {passwordRequirementsText}
                      </small>
                    )}
                  </div>
                  <div>
                    <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
                      Role *
                    </label>
                    <select
                      value={formData.role}
                      onChange={(e) => setFormData({ ...formData, role: e.target.value })}
                      style={styles.input}
                    >
                      <option value="viewer">Viewer (Read-only)</option>
                      <option value="operator">Operator (Manage schedules)</option>
                      <option value="admin">Admin (Full access)</option>
                    </select>
                  </div>
                </div>
                <div style={{ display: 'flex', gap: '1rem' }}>
                  <button
                    type="submit"
                    disabled={createUserMutation.isPending}
                    style={{
                      ...styles.buttonLarge,
                      backgroundColor: createUserMutation.isPending ? '#ccc' : '#0066cc',
                      cursor: createUserMutation.isPending ? 'not-allowed' : 'pointer',
                    }}
                  >
                    {createUserMutation.isPending ? 'Saving...' : editingUser ? 'Update User' : 'Create User'}
                  </button>
                  {editingUser && (
                    <button
                      type="button"
                      onClick={handleCancelEdit}
                      style={{
                        ...styles.buttonLarge,
                        backgroundColor: '#6c757d',
                      }}
                    >
                      Cancel
                    </button>
                  )}
                </div>
              </form>
            </div>
          )}

          <div style={styles.card}>
            {isLoading ? (
              <p>Loading users...</p>
            ) : users && users.length > 0 ? (
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ borderBottom: '2px solid #ddd' }}>
                    <th style={{ padding: '0.75rem', textAlign: 'left' }}>Username</th>
                    <th style={{ padding: '0.75rem', textAlign: 'left' }}>Role</th>
                    <th style={{ padding: '0.75rem', textAlign: 'left' }}>Created</th>
                    <th style={{ padding: '0.75rem', textAlign: 'left' }}>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {users.map((user) => (
                    <tr key={user.id} style={{ borderBottom: '1px solid #eee' }}>
                      <td style={{ padding: '0.75rem' }}>{user.username}</td>
                      <td style={{ padding: '0.75rem' }}>
                        <span style={{
                          padding: '0.25rem 0.5rem',
                          borderRadius: '4px',
                          fontSize: '0.875rem',
                          backgroundColor: user.role === 'admin' ? '#d4edda' :
                                         user.role === 'operator' ? '#cfe2ff' : '#f8f9fa',
                          color: user.role === 'admin' ? '#155724' :
                                 user.role === 'operator' ? '#084298' : '#666',
                        }}>
                          {user.role}
                        </span>
                      </td>
                      <td style={{ padding: '0.75rem' }}>
                        {new Date(user.created_at).toLocaleDateString()}
                      </td>
                      <td style={{ padding: '0.75rem' }}>
                        <button
                          onClick={() => handleEdit(user)}
                          style={{
                            ...styles.button,
                            backgroundColor: '#0066cc',
                            marginRight: '0.5rem',
                          }}
                        >
                          Edit
                        </button>
                        <button
                          onClick={() => handleDeleteClick(user)}
                          disabled={deleteUserMutation.isPending}
                          style={{
                            ...styles.button,
                            backgroundColor: '#dc3545',
                          }}
                        >
                          Delete
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : (
              <div style={{
                padding: '2rem',
                textAlign: 'center',
                color: '#666',
                backgroundColor: '#f8f9fa',
                borderRadius: '4px',
              }}>
                <p style={{ margin: 0 }}>No users found.</p>
                <p style={{ marginTop: '0.5rem', fontSize: '0.9rem' }}>
                  Click "Add User" to create the first user.
                </p>
              </div>
            )}
          </div>
        </div>
      )}

      {activeTab === 'clock' && <ClockNTPTab />}
      {activeTab === 'audit' && <AuditLogsTab />}

      <ConfirmDialog
        isOpen={confirmDelete.isOpen}
        title="Delete User"
        message={`Are you sure you want to delete user "${confirmDelete.user?.username}"?`}
        onConfirm={handleConfirmDelete}
        onCancel={() => setConfirmDelete({ isOpen: false, user: null })}
      />
    </div>
  );
}

// 
