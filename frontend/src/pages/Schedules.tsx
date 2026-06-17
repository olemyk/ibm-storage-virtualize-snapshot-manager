import { useState, useMemo, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { schedulesApi } from '../api/services';
import type { ScheduleWithVolumeGroup } from '../types';

// Constants
const MESSAGE_DISPLAY_DURATION_MS = 3000; // 3 seconds
const ERROR_MESSAGE_DISPLAY_DURATION_MS = 5000; // 5 seconds

// ConfirmDialog Component
interface ConfirmDialogProps {
  isOpen: boolean;
  title: string;
  message: string;
  onConfirm: () => void;
  onCancel: () => void;
  confirmText?: string;
  cancelText?: string;
}

function ConfirmDialog({ isOpen, title, message, onConfirm, onCancel, confirmText = 'Confirm', cancelText = 'Cancel' }: ConfirmDialogProps) {
  if (!isOpen) return null;

  return (
    <div style={{
      position: 'fixed',
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      backgroundColor: 'rgba(0, 0, 0, 0.5)',
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
        boxShadow: '0 4px 6px rgba(0, 0, 0, 0.1)',
      }}>
        <h3 style={{ marginTop: 0, marginBottom: '1rem' }}>{title}</h3>
        <p style={{ marginBottom: '1.5rem' }}>{message}</p>
        <div style={{ display: 'flex', gap: '1rem', justifyContent: 'flex-end' }}>
          <button
            onClick={onCancel}
            style={{
              padding: '0.5rem 1rem',
              backgroundColor: '#6c757d',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            {cancelText}
          </button>
          <button
            onClick={onConfirm}
            style={{
              padding: '0.5rem 1rem',
              backgroundColor: '#dc3545',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            {confirmText}
          </button>
        </div>
      </div>
    </div>
  );
}

export default function Schedules() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [confirmExecute, setConfirmExecute] = useState<{ schedule: ScheduleWithVolumeGroup } | null>(null);
  const [confirmToggle, setConfirmToggle] = useState<{ schedule: ScheduleWithVolumeGroup } | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<{ schedule: ScheduleWithVolumeGroup } | null>(null);

  // Fetch all schedules
  const { data: schedules = [], isLoading, error } = useQuery({
    queryKey: ['schedules'],
    queryFn: () => schedulesApi.list(),
  });

  // Filter schedules based on search query (memoized)
  const filteredSchedules = useMemo(() =>
    schedules.filter((schedule: ScheduleWithVolumeGroup) => {
      if (!searchQuery) return true;
      const query = searchQuery.toLowerCase();
      return (
        schedule.name.toLowerCase().includes(query) ||
        schedule.vg_name.toLowerCase().includes(query) ||
        schedule.system_name.toLowerCase().includes(query) ||
        schedule.cron_expression.toLowerCase().includes(query)
      );
    }),
    [schedules, searchQuery]
  );

  // Toggle schedule active status
  const toggleMutation = useMutation({
    mutationFn: ({ schedule, isActive }: { schedule: ScheduleWithVolumeGroup; isActive: boolean }) =>
      schedulesApi.update(schedule.id, {
        name: schedule.name,
        cron_expression: schedule.cron_expression,
        retention_days: schedule.retention_days,
        retention_minutes: schedule.retention_minutes,
        safeguarded: schedule.safeguarded,
        pool_name: schedule.pool_name,
        snapshot_name_pattern: schedule.snapshot_name_pattern || '{schedule_name}_{timestamp}',
        is_active: isActive,
      }),
    onSuccess: () => {
      setMessage({ type: 'success', text: 'Schedule status updated' });
      setTimeout(() => setMessage(null), MESSAGE_DISPLAY_DURATION_MS);
      queryClient.invalidateQueries({ queryKey: ['schedules'] });
    },
    onError: () => {
      setMessage({ type: 'error', text: 'Failed to update schedule status' });
      setTimeout(() => setMessage(null), ERROR_MESSAGE_DISPLAY_DURATION_MS);
    },
  });

  // Delete schedule
  const deleteMutation = useMutation({
    mutationFn: (id: number) => schedulesApi.delete(id),
    onSuccess: () => {
      setMessage({ type: 'success', text: 'Schedule deleted successfully' });
      setTimeout(() => setMessage(null), MESSAGE_DISPLAY_DURATION_MS);
      queryClient.invalidateQueries({ queryKey: ['schedules'] });
    },
    onError: () => {
      setMessage({ type: 'error', text: 'Failed to delete schedule' });
      setTimeout(() => setMessage(null), ERROR_MESSAGE_DISPLAY_DURATION_MS);
    },
  });

  // Execute schedule manually
  const executeMutation = useMutation({
    mutationFn: (id: number) => schedulesApi.execute(id),
    onSuccess: () => {
      setMessage({ type: 'success', text: 'Snapshot execution started' });
      setTimeout(() => setMessage(null), MESSAGE_DISPLAY_DURATION_MS);
      setConfirmExecute(null);
    },
    onError: (error: any) => {
      setMessage({
        type: 'error',
        text: error.response?.data?.error || 'Failed to execute snapshot'
      });
      setTimeout(() => setMessage(null), ERROR_MESSAGE_DISPLAY_DURATION_MS);
    },
  });

  const handleToggleActive = useCallback((schedule: ScheduleWithVolumeGroup) => {
    setConfirmToggle({ schedule });
  }, []);

  const handleDelete = useCallback((schedule: ScheduleWithVolumeGroup) => {
    setConfirmDelete({ schedule });
  }, []);

  const handleExecute = useCallback((schedule: ScheduleWithVolumeGroup) => {
    setConfirmExecute({ schedule });
  }, []);

  const confirmExecuteSnapshot = useCallback(() => {
    if (confirmExecute) {
      executeMutation.mutate(confirmExecute.schedule.id);
    }
  }, [confirmExecute, executeMutation]);

  const cancelExecuteSnapshot = useCallback(() => {
    setConfirmExecute(null);
  }, []);

  const formatCronDescription = (cron: string): string => {
    const cronMap: { [key: string]: string } = {
      '0 * * * *': 'Every hour',
      '0 */2 * * *': 'Every 2 hours',
      '0 */4 * * *': 'Every 4 hours',
      '0 */6 * * *': 'Every 6 hours',
      '0 */12 * * *': 'Every 12 hours',
      '0 0 * * *': 'Daily at 00:00',
      '0 2 * * *': 'Daily at 02:00',
      '0 6 * * *': 'Daily at 06:00',
      '0 12 * * *': 'Daily at 12:00',
      '0 18 * * *': 'Daily at 18:00',
      '0 0 * * 0': 'Weekly (Sunday 00:00)',
      '0 0 * * 1': 'Weekly (Monday 00:00)',
      '0 0 1 * *': 'Monthly (1st at 00:00)',
    };
    return cronMap[cron] || cron;
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return 'Never';
    return new Date(dateString).toLocaleString();
  };

  if (isLoading) {
    return (
      <div style={{ padding: '2rem' }}>
        <h1>Snapshot Schedules</h1>
        <p>Loading schedules...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ padding: '2rem' }}>
        <h1>Snapshot Schedules</h1>
        <div style={{
          padding: '1rem',
          backgroundColor: 'var(--danger-light)',
          color: 'var(--danger)',
          borderRadius: 'var(--radius-sm)',
          border: '1px solid var(--danger)',
        }}>
          Error loading schedules. Please try again.
        </div>
      </div>
    );
  }

  return (
    <>
      <ConfirmDialog
        isOpen={confirmToggle !== null}
        title={confirmToggle?.schedule.is_active ? 'Disable Schedule' : 'Enable Schedule'}
        message={`${confirmToggle?.schedule.is_active ? 'Disable' : 'Enable'} schedule "${confirmToggle?.schedule.name}"?`}
        onConfirm={() => {
          if (confirmToggle) {
            toggleMutation.mutate({ schedule: confirmToggle.schedule, isActive: !confirmToggle.schedule.is_active });
          }
          setConfirmToggle(null);
        }}
        onCancel={() => setConfirmToggle(null)}
        confirmText={confirmToggle?.schedule.is_active ? 'Disable' : 'Enable'}
      />
      <ConfirmDialog
        isOpen={confirmDelete !== null}
        title="Delete Schedule"
        message={`Are you sure you want to delete schedule "${confirmDelete?.schedule.name}"? This action cannot be undone.`}
        onConfirm={() => {
          if (confirmDelete) {
            deleteMutation.mutate(confirmDelete.schedule.id);
          }
          setConfirmDelete(null);
        }}
        onCancel={() => setConfirmDelete(null)}
        confirmText="Delete"
      />
      <style>{`
        .action-button:not(:disabled):hover {
          background-color: var(--bg-hover) !important;
          border-color: var(--border-strong) !important;
        }
        .search-input:focus {
          border-color: var(--primary) !important;
          box-shadow: 0 0 0 3px var(--primary-light) !important;
        }
      `}</style>
      <div style={{ padding: 'var(--spacing-xl)' }}>
      {/* Header */}
      <div style={{ marginBottom: 'var(--spacing-xl)' }}>
        <h1 style={{ margin: 0 }}>Snapshot Schedules</h1>
        <p style={{ margin: 'var(--spacing-sm) 0 0 0', color: 'var(--text-secondary)' }}>
          Manage all snapshot schedules across volume groups
        </p>
      </div>

      {/* Search Bar */}
      <div style={{ marginBottom: 'var(--spacing-lg)' }}>
        <input
          type="text"
          placeholder="🔍 Search schedules by name, volume group, system, or cron expression..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="search-input"
          style={{
            width: '100%',
            padding: '0.75rem 1rem',
            fontSize: '0.875rem',
            border: '1px solid var(--border-subtle)',
            borderRadius: 'var(--radius-sm)',
            backgroundColor: 'var(--bg-primary)',
            color: 'var(--text-primary)',
            transition: 'all var(--transition-fast)',
          }}
        />
        {searchQuery && (
          <p style={{
            margin: 'var(--spacing-sm) 0 0 0',
            fontSize: '0.875rem',
            color: 'var(--text-tertiary)'
          }}>
            Found {filteredSchedules.length} of {schedules.length} schedules
          </p>
        )}
      </div>

      {/* Message Display */}
      {message && (
        <div style={{
          padding: 'var(--spacing-md)',
          marginBottom: 'var(--spacing-lg)',
          borderRadius: 'var(--radius-sm)',
          backgroundColor: message.type === 'success' ? 'var(--success-light)' : 'var(--danger-light)',
          border: `1px solid ${message.type === 'success' ? 'var(--success)' : 'var(--danger)'}`,
          color: message.type === 'success' ? 'var(--success)' : 'var(--danger)',
        }}>
          {message.text}
        </div>
      )}

      {/* Execute Confirmation Panel */}
      {confirmExecute && (
        <div style={{
          backgroundColor: '#fff3cd',
          border: '2px solid #ffc107',
          borderRadius: '8px',
          padding: '1.5rem',
          marginBottom: '1.5rem',
          boxShadow: '0 4px 6px rgba(0,0,0,0.1)',
        }}>
          <h3 style={{
            marginTop: 0,
            marginBottom: '1rem',
            color: '#856404',
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
          }}>
            <span style={{ fontSize: '1.5rem' }}>⚠️</span>
            Execute Snapshot Now?
          </h3>
          
          <div style={{
            backgroundColor: 'white',
            padding: '1rem',
            borderRadius: '4px',
            marginBottom: '1rem',
            border: '1px solid #ffc107',
          }}>
            <table style={{ width: '100%', fontSize: '0.95rem' }}>
              <tbody>
                <tr>
                  <td style={{ padding: '0.5rem', fontWeight: 'bold', width: '150px' }}>Schedule:</td>
                  <td style={{ padding: '0.5rem' }}>{confirmExecute.schedule.name}</td>
                </tr>
                <tr>
                  <td style={{ padding: '0.5rem', fontWeight: 'bold' }}>Volume Group:</td>
                  <td style={{ padding: '0.5rem' }}>{confirmExecute.schedule.vg_name}</td>
                </tr>
                <tr>
                  <td style={{ padding: '0.5rem', fontWeight: 'bold' }}>Storage System:</td>
                  <td style={{ padding: '0.5rem' }}>{confirmExecute.schedule.system_name}</td>
                </tr>
                <tr>
                  <td style={{ padding: '0.5rem', fontWeight: 'bold' }}>Retention:</td>
                  <td style={{ padding: '0.5rem' }}>
                    {confirmExecute.schedule.retention_days} days
                    {confirmExecute.schedule.retention_minutes && ` + ${confirmExecute.schedule.retention_minutes} minutes`}
                  </td>
                </tr>
                <tr>
                  <td style={{ padding: '0.5rem', fontWeight: 'bold' }}>Safeguarded:</td>
                  <td style={{ padding: '0.5rem' }}>
                    <span style={{
                      padding: '0.25rem 0.75rem',
                      backgroundColor: confirmExecute.schedule.safeguarded ? '#d4edda' : '#f8f9fa',
                      color: confirmExecute.schedule.safeguarded ? '#155724' : '#666',
                      borderRadius: '12px',
                      fontSize: '0.85rem',
                      fontWeight: 'bold',
                    }}>
                      {confirmExecute.schedule.safeguarded ? 'Yes' : 'No'}
                    </span>
                  </td>
                </tr>
                {confirmExecute.schedule.pool_name && (
                  <tr>
                    <td style={{ padding: '0.5rem', fontWeight: 'bold' }}>Pool:</td>
                    <td style={{ padding: '0.5rem' }}>{confirmExecute.schedule.pool_name}</td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          <p style={{
            margin: '0 0 1rem 0',
            color: '#856404',
            fontSize: '0.95rem',
          }}>
            This will create a snapshot immediately, outside of the regular schedule.
          </p>

          <div style={{ display: 'flex', gap: '1rem' }}>
            <button
              onClick={confirmExecuteSnapshot}
              disabled={executeMutation.isPending}
              style={{
                padding: '0.75rem 1.5rem',
                backgroundColor: '#28a745',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: executeMutation.isPending ? 'not-allowed' : 'pointer',
                fontSize: '1rem',
                fontWeight: 'bold',
              }}
            >
              {executeMutation.isPending ? 'Executing...' : '✓ Yes, Execute Now'}
            </button>
            <button
              onClick={cancelExecuteSnapshot}
              disabled={executeMutation.isPending}
              style={{
                padding: '0.75rem 1.5rem',
                backgroundColor: '#6c757d',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: executeMutation.isPending ? 'not-allowed' : 'pointer',
                fontSize: '1rem',
              }}
            >
              ✕ Cancel
            </button>
          </div>
        </div>
      )}

      {/* Schedules Table */}
      {filteredSchedules.length === 0 ? (
        <div className="card" style={{ textAlign: 'center', padding: 'var(--spacing-2xl)' }}>
          <p style={{ fontSize: '1.125rem', marginBottom: 'var(--spacing-md)' }}>
            {searchQuery ? 'No schedules match your search' : 'No schedules found'}
          </p>
          <p style={{ color: 'var(--text-tertiary)', marginBottom: 'var(--spacing-lg)' }}>
            Create schedules from the Volume Groups page
          </p>
          <button
            onClick={() => navigate('/volume-groups')}
            style={{
              padding: 'var(--spacing-md) var(--spacing-xl)',
              backgroundColor: 'var(--primary)',
              color: 'white',
              border: 'none',
              borderRadius: 'var(--radius-sm)',
              cursor: 'pointer',
              fontWeight: 600,
            }}
          >
            Go to Volume Groups
          </button>
        </div>
      ) : (
        <div style={{ overflowX: 'auto' }}>
          <table>
            <thead>
              <tr>
                <th>Schedule Name</th>
                <th>Volume Group</th>
                <th>Storage System</th>
                <th>Frequency</th>
                <th>Retention</th>
                <th>Status</th>
                <th>Last Executed</th>
                <th>Next Execution</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredSchedules.map((schedule: ScheduleWithVolumeGroup) => (
                <tr key={schedule.id}>
                  <td>
                    <strong>{schedule.name}</strong>
                    {schedule.safeguarded && (
                      <span className="badge badge-info" style={{ marginLeft: 'var(--spacing-sm)' }}>
                        🔒 Safeguarded
                      </span>
                    )}
                  </td>
                  <td>{schedule.vg_name}</td>
                  <td>{schedule.system_name}</td>
                  <td>
                    <code style={{ fontSize: '0.8rem' }}>
                      {formatCronDescription(schedule.cron_expression)}
                    </code>
                  </td>
                  <td>
                    {schedule.retention_days} days
                    {schedule.retention_minutes && ` / ${schedule.retention_minutes} min`}
                  </td>
                  <td>
                    <span className={`badge ${schedule.is_active ? 'badge-success' : 'badge-warning'}`}>
                      {schedule.is_active ? 'Active' : 'Inactive'}
                    </span>
                  </td>
                  <td style={{ fontSize: '0.875rem' }}>
                    {formatDate(schedule.last_executed_at)}
                  </td>
                  <td style={{ fontSize: '0.875rem' }}>
                    {formatDate(schedule.next_execution_at)}
                  </td>
                  <td>
                    <div style={{ display: 'flex', gap: 'var(--spacing-xs)', alignItems: 'center' }}>
                      <button
                        onClick={() => handleExecute(schedule)}
                        disabled={executeMutation.isPending}
                        className="action-button"
                        style={{
                          padding: '0.375rem 0.5rem',
                          backgroundColor: 'var(--bg-secondary)',
                          color: 'var(--text-primary)',
                          border: '1px solid var(--border-subtle)',
                          borderRadius: 'var(--radius-sm)',
                          cursor: executeMutation.isPending ? 'not-allowed' : 'pointer',
                          fontSize: '1rem',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          transition: 'all var(--transition-fast)',
                          opacity: executeMutation.isPending ? 0.5 : 1,
                          minWidth: '32px',
                        }}
                        title="Execute snapshot now"
                      >
                        ▶️
                      </button>
                      <button
                        onClick={() => handleToggleActive(schedule)}
                        disabled={toggleMutation.isPending}
                        className="action-button"
                        style={{
                          padding: '0.375rem 0.5rem',
                          backgroundColor: 'var(--bg-secondary)',
                          color: 'var(--text-primary)',
                          border: '1px solid var(--border-subtle)',
                          borderRadius: 'var(--radius-sm)',
                          cursor: toggleMutation.isPending ? 'not-allowed' : 'pointer',
                          fontSize: '1rem',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          transition: 'all var(--transition-fast)',
                          opacity: toggleMutation.isPending ? 0.5 : 1,
                          minWidth: '32px',
                        }}
                        title={schedule.is_active ? 'Disable schedule' : 'Enable schedule'}
                      >
                        {schedule.is_active ? '⏸️' : '✅'}
                      </button>
                      <button
                        onClick={() => navigate(`/volumegroups/${schedule.volume_group_id}/schedules`)}
                        className="action-button"
                        style={{
                          padding: '0.375rem 0.5rem',
                          backgroundColor: 'var(--bg-secondary)',
                          color: 'var(--text-primary)',
                          border: '1px solid var(--border-subtle)',
                          borderRadius: 'var(--radius-sm)',
                          cursor: 'pointer',
                          fontSize: '1rem',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          transition: 'all var(--transition-fast)',
                          minWidth: '32px',
                        }}
                        title="Edit schedule"
                      >
                        ✏️
                      </button>
                      <button
                        onClick={() => handleDelete(schedule)}
                        disabled={deleteMutation.isPending}
                        className="action-button"
                        style={{
                          padding: '0.375rem 0.5rem',
                          backgroundColor: 'var(--bg-secondary)',
                          color: 'var(--text-primary)',
                          border: '1px solid var(--border-subtle)',
                          borderRadius: 'var(--radius-sm)',
                          cursor: deleteMutation.isPending ? 'not-allowed' : 'pointer',
                          fontSize: '1rem',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          transition: 'all var(--transition-fast)',
                          opacity: deleteMutation.isPending ? 0.5 : 1,
                          minWidth: '32px',
                        }}
                        title="Delete schedule"
                      >
                        🗑️
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Summary */}
      {schedules.length > 0 && (
        <div style={{
          marginTop: 'var(--spacing-xl)',
          padding: 'var(--spacing-md)',
          backgroundColor: 'var(--bg-primary)',
          borderRadius: 'var(--radius-sm)',
          border: '1px solid var(--border-subtle)',
        }}>
          <strong>Summary:</strong> {schedules.length} total schedule{schedules.length !== 1 ? 's' : ''}
          {' • '}
          {schedules.filter((s: ScheduleWithVolumeGroup) => s.is_active).length} active
          {' • '}
          {schedules.filter((s: ScheduleWithVolumeGroup) => !s.is_active).length} inactive
        </div>
      )}
      </div>
    </>
  );
}

// 
