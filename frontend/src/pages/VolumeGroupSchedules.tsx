import { useState, useEffect, useMemo } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation } from '@tanstack/react-query';
import { volumeGroupsApi, schedulesApi, systemsApi } from '../api/services';
import type { CreateScheduleRequest, Snapshot } from '../types';

// Common cron schedule presets
const CRON_PRESETS = [
  { label: 'Every hour', value: '0 * * * *', description: 'At minute 0 of every hour' },
  { label: 'Every 2 hours', value: '0 */2 * * *', description: 'At minute 0 of every 2nd hour' },
  { label: 'Every 4 hours', value: '0 */4 * * *', description: 'At minute 0 of every 4th hour' },
  { label: 'Every 6 hours', value: '0 */6 * * *', description: 'At minute 0 of every 6th hour' },
  { label: 'Every 12 hours', value: '0 */12 * * *', description: 'At minute 0 of every 12th hour' },
  { label: 'Daily at 00:00', value: '0 0 * * *', description: 'At 00:00 every day' },
  { label: 'Daily at 02:00', value: '0 2 * * *', description: 'At 02:00 every day' },
  { label: 'Daily at 06:00', value: '0 6 * * *', description: 'At 06:00 every day' },
  { label: 'Daily at 12:00', value: '0 12 * * *', description: 'At 12:00 every day' },
  { label: 'Daily at 18:00', value: '0 18 * * *', description: 'At 18:00 every day' },
  { label: 'Weekly (Sunday 00:00)', value: '0 0 * * 0', description: 'At 00:00 on Sunday' },
  { label: 'Weekly (Monday 00:00)', value: '0 0 * * 1', description: 'At 00:00 on Monday' },
  { label: 'Monthly (1st at 00:00)', value: '0 0 1 * *', description: 'At 00:00 on day 1 of every month' },
  { label: 'Custom', value: 'custom', description: 'Enter your own cron expression' },
];

// Format cron expression to human-readable description
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

export default function VolumeGroupSchedules() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const vgId = parseInt(id || '0', 10);

  const [activeTab, setActiveTab] = useState<'schedules' | 'snapshots'>('schedules');
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [editingSchedule, setEditingSchedule] = useState<number | null>(null);
  const [cronPreset, setCronPreset] = useState<string>('0 2 * * *');
  const [showCustomCron, setShowCustomCron] = useState(false);
  const [snapshotSearch, setSnapshotSearch] = useState('');
  const [confirmExecute, setConfirmExecute] = useState<{ scheduleId: number; scheduleName: string } | null>(null);
  const [formData, setFormData] = useState<CreateScheduleRequest>({
    name: '',
    cron_expression: '0 2 * * *',
    retention_days: 7,
    retention_minutes: undefined,
    safeguarded: false,
    pool_name: '',
    snapshot_name_pattern: '{schedule_name}_{timestamp}',
    is_active: true,
  });

  // Update cron expression when preset changes
  useEffect(() => {
    if (cronPreset !== 'custom') {
      setFormData(prev => ({ ...prev, cron_expression: cronPreset }));
      setShowCustomCron(false);
    } else {
      setShowCustomCron(true);
    }
  }, [cronPreset]);

  // Fetch volume group details
  const { data: volumeGroup, isLoading: vgLoading } = useQuery({
    queryKey: ['volumeGroup', vgId],
    queryFn: () => volumeGroupsApi.get(vgId),
    enabled: vgId > 0,
  });

  // Fetch storage system details
  const { data: storageSystem } = useQuery({
    queryKey: ['storageSystem', volumeGroup?.storage_system_id],
    queryFn: () => systemsApi.get(volumeGroup!.storage_system_id),
    enabled: !!volumeGroup?.storage_system_id,
  });

  // Fetch volumes in this volume group
  const { data: volumes = [] } = useQuery({
    queryKey: ['volumes', vgId],
    queryFn: () => volumeGroupsApi.listVolumes(vgId),
    enabled: vgId > 0,
  });

  // Fetch schedules for this volume group
  const { data: schedules = [], isLoading: schedulesLoading, refetch } = useQuery({
    queryKey: ['schedules', vgId],
    queryFn: () => schedulesApi.listByVolumeGroup(vgId),
    enabled: vgId > 0,
  });

  // Fetch snapshots for this volume group (fetch immediately to show correct count in tab)
  const { data: snapshots = [], isLoading: snapshotsLoading, refetch: refetchSnapshots } = useQuery({
    queryKey: ['snapshots', vgId],
    queryFn: () => volumeGroupsApi.listSnapshots(vgId),
    enabled: vgId > 0,
  });

  // Create schedule mutation
  const createMutation = useMutation({
    mutationFn: (data: CreateScheduleRequest) => schedulesApi.create(vgId, data),
    onSuccess: () => {
      setMessage({ type: 'success', text: 'Schedule created successfully' });
      setTimeout(() => setMessage(null), 5000);
      setShowForm(false);
      resetForm();
      refetch();
    },
    onError: (error: any) => {
      setMessage({ 
        type: 'error', 
        text: error.response?.data?.error || 'Failed to create schedule' 
      });
      setTimeout(() => setMessage(null), 10000);
    },
  });

  // Update schedule mutation
  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: CreateScheduleRequest }) => 
      schedulesApi.update(id, data),
    onSuccess: () => {
      setMessage({ type: 'success', text: 'Schedule updated successfully' });
      setTimeout(() => setMessage(null), 5000);
      setShowForm(false);
      setEditingSchedule(null);
      resetForm();
      refetch();
    },
    onError: (error: any) => {
      setMessage({ 
        type: 'error', 
        text: error.response?.data?.error || 'Failed to update schedule' 
      });
      setTimeout(() => setMessage(null), 10000);
    },
  });

  // Delete schedule mutation
  const deleteMutation = useMutation({
    mutationFn: (scheduleId: number) => schedulesApi.delete(scheduleId),
    onSuccess: () => {
      setMessage({ type: 'success', text: 'Schedule deleted successfully' });
      setTimeout(() => setMessage(null), 5000);
      refetch();
    },
    onError: (error: any) => {
      setMessage({ 
        type: 'error', 
        text: error.response?.data?.error || 'Failed to delete schedule' 
      });
      setTimeout(() => setMessage(null), 10000);
    },
  });

  // Execute schedule mutation
  const executeMutation = useMutation({
    mutationFn: (scheduleId: number) => schedulesApi.execute(scheduleId),
    onSuccess: () => {
      setMessage({ type: 'success', text: 'Snapshot execution started' });
      setTimeout(() => setMessage(null), 5000);
    },
    onError: (error: any) => {
      setMessage({ 
        type: 'error', 
        text: error.response?.data?.error || 'Failed to execute schedule' 
      });
      setTimeout(() => setMessage(null), 10000);
    },
  });

  const resetForm = () => {
    setCronPreset('0 2 * * *');
    setShowCustomCron(false);
    setFormData({
      name: '',
      cron_expression: '0 2 * * *',
      retention_days: 7,
      retention_minutes: undefined,
      safeguarded: false,
      pool_name: '',
      snapshot_name_pattern: '{schedule_name}_{timestamp}',
      is_active: true,
    });
  };

  const handleEdit = (schedule: any) => {
    const matchingPreset = CRON_PRESETS.find((preset) => preset.value === schedule.cron_expression);

    setEditingSchedule(schedule.id);
    setCronPreset(matchingPreset ? matchingPreset.value : 'custom');
    setShowCustomCron(!matchingPreset);
    setFormData({
      name: schedule.name,
      cron_expression: schedule.cron_expression,
      retention_days: schedule.retention_days,
      retention_minutes: schedule.retention_minutes,
      safeguarded: schedule.safeguarded,
      pool_name: schedule.pool_name || '',
      snapshot_name_pattern: schedule.snapshot_name_pattern || '{schedule_name}_{timestamp}',
      is_active: schedule.is_active,
    });
    setShowForm(true);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    const hasValidRetentionDays = Number.isInteger(formData.retention_days) && formData.retention_days >= 1;
    const hasValidRetentionMinutes =
      formData.retention_minutes !== undefined &&
      Number.isInteger(formData.retention_minutes) &&
      formData.retention_minutes >= 0;

    if (!hasValidRetentionDays && !hasValidRetentionMinutes) {
      setMessage({
        type: 'error',
        text: 'Enter either Retention Days (minimum 1) or Retention Minutes.',
      });
      setTimeout(() => setMessage(null), 5000);
      return;
    }
    
    const submitData = {
      ...formData,
      retention_days: hasValidRetentionDays ? formData.retention_days : 0,
      pool_name: formData.pool_name?.trim() || undefined,
      retention_minutes: hasValidRetentionMinutes ? formData.retention_minutes : undefined,
      snapshot_name_pattern: formData.snapshot_name_pattern?.trim() || '{schedule_name}_{timestamp}',
    };

    if (editingSchedule) {
      updateMutation.mutate({ id: editingSchedule, data: submitData });
    } else {
      createMutation.mutate(submitData);
    }
  };

  const handleCancel = () => {
    setShowForm(false);
    setEditingSchedule(null);
    resetForm();
  };

  const handleDelete = (scheduleId: number, scheduleName: string) => {
    if (window.confirm(`Are you sure you want to delete schedule "${scheduleName}"?`)) {
      deleteMutation.mutate(scheduleId);
    }
  };

  const handleExecute = (scheduleId: number, scheduleName: string) => {
    setConfirmExecute({ scheduleId, scheduleName });
  };

  const confirmExecuteSnapshot = () => {
    if (confirmExecute) {
      executeMutation.mutate(confirmExecute.scheduleId);
      setConfirmExecute(null);
    }
  };

  const cancelExecuteSnapshot = () => {
    setConfirmExecute(null);
  };

  // Format capacity for display (IBM SVC returns capacity like "3.00MB" or "40.00GB")
  const formatCapacity = (capacity: string | undefined): string => {
    if (!capacity) return 'N/A';
    return capacity; // IBM SVC already formats it nicely
  };

  // Format timestamp for display (IBM SVC returns time in YYMMDDHHMMSS format like "260526122403")
  const formatTimestamp = (timestamp: string | undefined): string => {
    if (!timestamp || timestamp === '') return 'N/A';
    
    // Parse IBM SVC timestamp format: YYMMDDHHMMSS
    if (timestamp.length === 12) {
      try {
        const year = 2000 + parseInt(timestamp.substring(0, 2));
        const month = parseInt(timestamp.substring(2, 4)) - 1; // JS months are 0-indexed
        const day = parseInt(timestamp.substring(4, 6));
        const hour = parseInt(timestamp.substring(6, 8));
        const minute = parseInt(timestamp.substring(8, 10));
        const second = parseInt(timestamp.substring(10, 12));
        
        const date = new Date(year, month, day, hour, minute, second);
        return date.toLocaleString();
      } catch {
        return timestamp;
      }
    }
    
    // Fallback for other formats
    try {
      return new Date(timestamp).toLocaleString();
    } catch {
      return timestamp;
    }
  };

  // Filter snapshots based on search (memoized)
  const filteredSnapshots = useMemo(() =>
    snapshots.filter((snapshot: Snapshot) => {
      if (!snapshotSearch) return true;
      const query = snapshotSearch.toLowerCase();
      return (
        snapshot.name?.toLowerCase().includes(query) ||
        snapshot.id?.toLowerCase().includes(query)
      );
    }),
    [snapshots, snapshotSearch]
  );

  if (vgLoading) {
    return (
      <div style={{ padding: '2rem' }}>
        <p>Loading volume group...</p>
      </div>
    );
  }

  if (!volumeGroup) {
    return (
      <div style={{ padding: '2rem' }}>
        <h1>Volume Group Not Found</h1>
        <button onClick={() => navigate('/volume-groups')} style={{
          padding: '0.5rem 1rem',
          backgroundColor: '#007bff',
          color: 'white',
          border: 'none',
          borderRadius: '4px',
          cursor: 'pointer',
          marginTop: '1rem',
        }}>
          Back to Volume Groups
        </button>
      </div>
    );
  }

  return (
    <div style={{ padding: '2rem' }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <div>
          <h1 style={{ margin: 0 }}>Volume Group: {volumeGroup.vg_name}</h1>
          {storageSystem && (
            <p style={{ margin: '0.5rem 0 0 0', color: '#666', fontSize: '0.95rem' }}>
              Storage System: <strong>{storageSystem.name}</strong>
              <span style={{ marginLeft: '0.5rem', fontSize: '0.85rem', color: '#999' }}>
                ({storageSystem.ip_address})
              </span>
            </p>
          )}
          <p style={{ margin: '0.25rem 0 0 0', color: '#666' }}>
            VG ID: <strong>{volumeGroup.vg_id || 'N/A'}</strong>
            {volumeGroup.partition_name && (
              <span style={{ marginLeft: '1.5rem' }}>
                Partition: <strong>{volumeGroup.partition_name}</strong>
                {volumeGroup.partition_id && (
                  <span style={{ marginLeft: '0.5rem', fontSize: '0.85rem', color: '#999' }}>
                    (ID: {volumeGroup.partition_id})
                  </span>
                )}
              </span>
            )}
          </p>
          <p style={{ margin: '0.25rem 0 0 0', color: '#666', fontSize: '0.9rem' }}>
            Volumes: <strong>{volumes.length}</strong>
            {volumes.length > 0 && (
              <span style={{ marginLeft: '0.5rem' }}>
                ({volumes.map((v: any) => v.name).join(', ')})
              </span>
            )}
          </p>
        </div>
        <button
          onClick={() => navigate('/volume-groups')}
          style={{
            padding: '0.5rem 1rem',
            backgroundColor: '#6c757d',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
          }}
        >
          ← Back to Volume Groups
        </button>
      </div>

      {/* Tab Navigation */}
      <div style={{
        display: 'flex',
        gap: '0.5rem',
        marginBottom: '1.5rem',
        borderBottom: '2px solid #dee2e6',
      }}>
        <button
          onClick={() => setActiveTab('schedules')}
          style={{
            padding: '0.75rem 1.5rem',
            backgroundColor: activeTab === 'schedules' ? '#007bff' : 'transparent',
            color: activeTab === 'schedules' ? 'white' : '#666',
            border: 'none',
            borderBottom: activeTab === 'schedules' ? '3px solid #007bff' : '3px solid transparent',
            cursor: 'pointer',
            fontSize: '1rem',
            fontWeight: activeTab === 'schedules' ? 'bold' : 'normal',
            transition: 'all 0.2s',
          }}
        >
          Schedules ({schedules?.length || 0})
        </button>
        <button
          onClick={() => setActiveTab('snapshots')}
          style={{
            padding: '0.75rem 1.5rem',
            backgroundColor: activeTab === 'snapshots' ? '#007bff' : 'transparent',
            color: activeTab === 'snapshots' ? 'white' : '#666',
            border: 'none',
            borderBottom: activeTab === 'snapshots' ? '3px solid #007bff' : '3px solid transparent',
            cursor: 'pointer',
            fontSize: '1rem',
            fontWeight: activeTab === 'snapshots' ? 'bold' : 'normal',
            transition: 'all 0.2s',
          }}
        >
          Snapshots ({snapshots?.length || 0})
        </button>
      </div>

      {/* Message Display */}
      {message && (
        <div style={{
          padding: '1rem',
          marginBottom: '1rem',
          borderRadius: '4px',
          backgroundColor: message.type === 'success' ? '#d4edda' : '#f8d7da',
          border: `1px solid ${message.type === 'success' ? '#c3e6cb' : '#f5c6cb'}`,
          color: message.type === 'success' ? '#155724' : '#721c24',
        }}>
          {message.text}
        </div>
      )}

      {/* Execute Confirmation Panel */}
      {confirmExecute && (() => {
        const schedule = schedules.find(s => s.id === confirmExecute.scheduleId);
        if (!schedule) return null;
        
        return (
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
                    <td style={{ padding: '0.5rem' }}>{schedule.name}</td>
                  </tr>
                  <tr>
                    <td style={{ padding: '0.5rem', fontWeight: 'bold' }}>Volume Group:</td>
                    <td style={{ padding: '0.5rem' }}>{volumeGroup?.vg_name || 'Unknown'}</td>
                  </tr>
                  <tr>
                    <td style={{ padding: '0.5rem', fontWeight: 'bold' }}>Retention:</td>
                    <td style={{ padding: '0.5rem' }}>
                      {schedule.retention_days} days
                      {schedule.retention_minutes && ` + ${schedule.retention_minutes} minutes`}
                    </td>
                  </tr>
                  <tr>
                    <td style={{ padding: '0.5rem', fontWeight: 'bold' }}>Safeguarded:</td>
                    <td style={{ padding: '0.5rem' }}>
                      <span style={{
                        padding: '0.25rem 0.75rem',
                        backgroundColor: schedule.safeguarded ? '#d4edda' : '#f8f9fa',
                        color: schedule.safeguarded ? '#155724' : '#666',
                        borderRadius: '12px',
                        fontSize: '0.85rem',
                        fontWeight: 'bold',
                      }}>
                        {schedule.safeguarded ? 'Yes' : 'No'}
                      </span>
                    </td>
                  </tr>
                  {schedule.pool_name && (
                    <tr>
                      <td style={{ padding: '0.5rem', fontWeight: 'bold' }}>Pool:</td>
                      <td style={{ padding: '0.5rem' }}>{schedule.pool_name}</td>
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
        );
      })()}

      {/* Schedules Tab Content */}
      {activeTab === 'schedules' && (
        <>
          {/* Add Schedule Button */}
          {!showForm && (
            <button
              onClick={() => setShowForm(true)}
              style={{
                padding: '0.75rem 1.5rem',
                backgroundColor: '#28a745',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: 'pointer',
                fontSize: '1rem',
                marginBottom: '1.5rem',
              }}
            >
              + Add New Schedule
            </button>
          )}

          {/* Schedule Form */}
          {showForm && (
            <div style={{
              backgroundColor: 'white',
              padding: '1.5rem',
              borderRadius: '8px',
              boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
              marginBottom: '1.5rem',
            }}>
              <h2 style={{ marginTop: 0 }}>
                {editingSchedule ? 'Edit Schedule' : 'Create New Schedule'}
              </h2>
              <form onSubmit={handleSubmit}>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                  <div>
                    <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
                      Schedule Name *
                    </label>
                    <input
                      type="text"
                      value={formData.name}
                      onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                      required
                      placeholder="e.g., Daily Backup"
                      style={{
                        width: '100%',
                        padding: '0.5rem',
                        border: '1px solid #ddd',
                        borderRadius: '4px',
                      }}
                    />
                  </div>

                  <div>
                    <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
                      Schedule Frequency *
                    </label>
                    <select
                      value={cronPreset}
                      onChange={(e) => setCronPreset(e.target.value)}
                      style={{
                        width: '100%',
                        padding: '0.5rem',
                        border: '1px solid #ddd',
                        borderRadius: '4px',
                        backgroundColor: 'white',
                      }}
                    >
                      {CRON_PRESETS.map((preset) => (
                        <option key={preset.value} value={preset.value}>
                          {preset.label}
                        </option>
                      ))}
                    </select>
                    {!showCustomCron && (
                      <p style={{
                        margin: '0.5rem 0 0 0',
                        fontSize: '0.85rem',
                        color: '#666',
                        fontStyle: 'italic'
                      }}>
                        {CRON_PRESETS.find(p => p.value === cronPreset)?.description}
                      </p>
                    )}
                  </div>

                  {showCustomCron && (
                    <div>
                      <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
                        Custom Cron Expression *
                        <span style={{ fontWeight: 'normal', fontSize: '0.85rem', marginLeft: '0.5rem', color: '#666' }}>
                          (minute hour day month weekday)
                        </span>
                      </label>
                      <input
                        type="text"
                        value={formData.cron_expression}
                        onChange={(e) => setFormData({ ...formData, cron_expression: e.target.value })}
                        required
                        placeholder="0 2 * * *"
                        style={{
                          width: '100%',
                          padding: '0.5rem',
                          border: '1px solid #ddd',
                          borderRadius: '4px',
                        }}
                      />
                      <p style={{
                        margin: '0.5rem 0 0 0',
                        fontSize: '0.85rem',
                        color: '#666'
                      }}>
                        Format: minute (0-59) hour (0-23) day (1-31) month (1-12) weekday (0-6, 0=Sunday)
                      </p>
                    </div>
                  )}

                  <div>
                    <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
                      Retention Days
                    </label>
                    <input
                      type="number"
                      value={Number.isNaN(formData.retention_days) ? '' : formData.retention_days}
                      onChange={(e) => setFormData({
                        ...formData,
                        retention_days: e.target.value === '' ? NaN : parseInt(e.target.value, 10),
                      })}
                      min="1"
                      aria-invalid={
                        (!Number.isInteger(formData.retention_days) || formData.retention_days < 1) &&
                        formData.retention_minutes === undefined
                      }
                      style={{
                        width: '100%',
                        padding: '0.5rem',
                        border: `1px solid ${
                          (!Number.isInteger(formData.retention_days) || formData.retention_days < 1) &&
                          formData.retention_minutes === undefined
                            ? '#dc3545'
                            : '#ddd'
                        }`,
                        borderRadius: '4px',
                      }}
                    />
                    <p style={{
                      margin: '0.5rem 0 0 0',
                      fontSize: '0.85rem',
                      color: '#666',
                    }}>
                      Optional if Retention Minutes is provided.
                    </p>
                  </div>

                  <div>
                    <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
                      Retention Minutes
                    </label>
                    <input
                      type="number"
                      value={formData.retention_minutes !== undefined ? formData.retention_minutes : ''}
                      onChange={(e) => setFormData({
                        ...formData,
                        retention_minutes: e.target.value === '' ? undefined : parseInt(e.target.value, 10)
                      })}
                      min="0"
                      placeholder="Leave empty if using days"
                      aria-invalid={
                        formData.retention_minutes === undefined &&
                        (!Number.isInteger(formData.retention_days) || formData.retention_days < 1)
                      }
                      style={{
                        width: '100%',
                        padding: '0.5rem',
                        border: `1px solid ${
                          formData.retention_minutes === undefined &&
                          (!Number.isInteger(formData.retention_days) || formData.retention_days < 1)
                            ? '#dc3545'
                            : '#ddd'
                        }`,
                        borderRadius: '4px',
                      }}
                    />
                    {formData.retention_minutes === undefined &&
                    (!Number.isInteger(formData.retention_days) || formData.retention_days < 1) ? (
                      <p style={{
                        margin: '0.5rem 0 0 0',
                        fontSize: '0.85rem',
                        color: '#dc3545',
                      }}>
                        Enter Retention Days or Retention Minutes.
                      </p>
                    ) : (
                      <p style={{
                        margin: '0.5rem 0 0 0',
                        fontSize: '0.85rem',
                        color: '#666',
                      }}>
                        Optional if Retention Days is provided.
                      </p>
                    )}
                  </div>

                  <div>
                    <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
                      Pool Name (optional)
                    </label>
                    <input
                      type="text"
                      value={formData.pool_name}
                      onChange={(e) => setFormData({ ...formData, pool_name: e.target.value })}
                      placeholder="Leave empty for default"
                      style={{
                        width: '100%',
                        padding: '0.5rem',
                        border: '1px solid #ddd',
                        borderRadius: '4px',
                      }}
                    />
                  </div>

                  <div style={{ gridColumn: '1 / -1' }}>
                    <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
                      Snapshot Name Pattern
                      <span style={{ fontWeight: 'normal', fontSize: '0.85rem', marginLeft: '0.5rem', color: '#666' }}>
                        (Customize how snapshots are named)
                      </span>
                    </label>
                    <input
                      type="text"
                      value={formData.snapshot_name_pattern}
                      onChange={(e) => setFormData({ ...formData, snapshot_name_pattern: e.target.value })}
                      placeholder="{schedule_name}_{timestamp}"
                      style={{
                        width: '100%',
                        padding: '0.5rem',
                        border: '1px solid #ddd',
                        borderRadius: '4px',
                        fontFamily: 'monospace',
                      }}
                    />
                    <p style={{
                      margin: '0.5rem 0 0 0',
                      fontSize: '0.85rem',
                      color: '#666'
                    }}>
                      Available placeholders: <code>{'{schedule_name}'}</code>, <code>{'{vg_name}'}</code>, <code>{'{timestamp}'}</code> (YYYYMMDD_HHMMSS),
                      <code>{'{date}'}</code> (YYYYMMDD), <code>{'{time}'}</code> (HHMMSS), <code>{'{year}'}</code>, <code>{'{month}'}</code>, <code>{'{day}'}</code>
                    </p>
                    <p style={{
                      margin: '0.25rem 0 0 0',
                      fontSize: '0.85rem',
                      color: '#666',
                      fontStyle: 'italic'
                    }}>
                      Example: <code>backup_{'{vg_name}'}_{'{date}'}</code> → backup_vg001_20260531
                    </p>
                  </div>

                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                    <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                      <input
                        type="checkbox"
                        checked={formData.safeguarded}
                        onChange={(e) => setFormData({ ...formData, safeguarded: e.target.checked })}
                      />
                      <span style={{ fontWeight: 'bold' }}>Safeguarded (Immutable)</span>
                    </label>

                    <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                      <input
                        type="checkbox"
                        checked={formData.is_active}
                        onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                      />
                      <span style={{ fontWeight: 'bold' }}>Active</span>
                    </label>
                  </div>
                </div>

                <div style={{ display: 'flex', gap: '1rem', marginTop: '1.5rem' }}>
                  <button
                    type="submit"
                    disabled={createMutation.isPending || updateMutation.isPending}
                    style={{
                      padding: '0.75rem 1.5rem',
                      backgroundColor: '#007bff',
                      color: 'white',
                      border: 'none',
                      borderRadius: '4px',
                      cursor: createMutation.isPending || updateMutation.isPending ? 'not-allowed' : 'pointer',
                    }}
                  >
                    {createMutation.isPending || updateMutation.isPending 
                      ? 'Saving...' 
                      : editingSchedule ? 'Update Schedule' : 'Create Schedule'}
                  </button>
                  <button
                    type="button"
                    onClick={handleCancel}
                    style={{
                      padding: '0.75rem 1.5rem',
                      backgroundColor: '#6c757d',
                      color: 'white',
                      border: 'none',
                      borderRadius: '4px',
                      cursor: 'pointer',
                    }}
                  >
                    Cancel
                  </button>
                </div>
              </form>
            </div>
          )}

          {/* Schedules List */}
          <div style={{
            backgroundColor: 'white',
            padding: '1.5rem',
            borderRadius: '8px',
            boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
          }}>
            <h2 style={{ marginTop: 0 }}>Existing Schedules ({schedules?.length || 0})</h2>

            {schedulesLoading ? (
              <p>Loading schedules...</p>
            ) : !schedules || schedules.length === 0 ? (
              <div style={{
                padding: '2rem',
                textAlign: 'center',
                color: '#666',
                backgroundColor: '#f8f9fa',
                borderRadius: '4px',
              }}>
                <p style={{ margin: 0 }}>No schedules configured for this volume group.</p>
                <p style={{ marginTop: '0.5rem', fontSize: '0.9rem' }}>
                  Click "Add New Schedule" to create one.
                </p>
              </div>
            ) : (
              <div style={{ overflowX: 'auto' }}>
                <table style={{
                  width: '100%',
                  borderCollapse: 'collapse',
                  fontSize: '0.95rem',
                }}>
                  <thead>
                    <tr style={{ backgroundColor: '#f8f9fa', borderBottom: '2px solid #dee2e6' }}>
                      <th style={{ padding: '0.75rem', textAlign: 'left' }}>Name</th>
                      <th style={{ padding: '0.75rem', textAlign: 'left' }}>Frequency</th>
                      <th style={{ padding: '0.75rem', textAlign: 'center' }}>Retention</th>
                      <th style={{ padding: '0.75rem', textAlign: 'center' }}>Safeguarded</th>
                      <th style={{ padding: '0.75rem', textAlign: 'center' }}>Status</th>
                      <th style={{ padding: '0.75rem', textAlign: 'left' }}>Next Run</th>
                      <th style={{ padding: '0.75rem', textAlign: 'center' }}>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {schedules.map((schedule) => (
                      <tr key={schedule.id} style={{ borderBottom: '1px solid #dee2e6' }}>
                        <td style={{ padding: '0.75rem' }}>
                          <strong>{schedule.name}</strong>
                        </td>
                        <td style={{ padding: '0.75rem' }}>
                          <span style={{
                            padding: '0.25rem 0.75rem',
                            backgroundColor: '#d4edda',
                            color: '#155724',
                            borderRadius: '12px',
                            fontSize: '0.85rem',
                            display: 'inline-block',
                          }}>
                            {formatCronDescription(schedule.cron_expression)}
                          </span>
                        </td>
                        <td style={{ padding: '0.75rem', textAlign: 'center' }}>
                          {schedule.retention_days}d
                          {schedule.retention_minutes && ` + ${schedule.retention_minutes}m`}
                        </td>
                        <td style={{ padding: '0.75rem', textAlign: 'center' }}>
                          {schedule.safeguarded ? (
                            <span style={{ color: '#28a745', fontWeight: 'bold' }}>✓</span>
                          ) : (
                            <span style={{ color: '#999' }}>—</span>
                          )}
                        </td>
                        <td style={{ padding: '0.75rem', textAlign: 'center' }}>
                          <span style={{
                            padding: '0.25rem 0.75rem',
                            backgroundColor: schedule.is_active ? '#d4edda' : '#f8d7da',
                            color: schedule.is_active ? '#155724' : '#721c24',
                            borderRadius: '12px',
                            fontSize: '0.85rem',
                          }}>
                            {schedule.is_active ? 'Active' : 'Inactive'}
                          </span>
                        </td>
                        <td style={{ padding: '0.75rem', fontSize: '0.9rem', color: '#666' }}>
                          {schedule.next_execution_at 
                            ? new Date(schedule.next_execution_at).toLocaleString()
                            : 'Not scheduled'}
                        </td>
                        <td style={{ padding: '0.75rem', textAlign: 'center' }}>
                          <div style={{ display: 'flex', gap: '0.5rem', justifyContent: 'center' }}>
                            <button
                              onClick={() => handleExecute(schedule.id, schedule.name)}
                              disabled={executeMutation.isPending}
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
                              onMouseOver={(e) => {
                                if (!executeMutation.isPending) {
                                  e.currentTarget.style.backgroundColor = 'var(--bg-hover)';
                                  e.currentTarget.style.borderColor = 'var(--border-strong)';
                                }
                              }}
                              onMouseOut={(e) => {
                                e.currentTarget.style.backgroundColor = 'var(--bg-secondary)';
                                e.currentTarget.style.borderColor = 'var(--border-subtle)';
                              }}
                              title="Execute snapshot now"
                            >
                              ▶️
                            </button>
                            <button
                              onClick={() => handleEdit(schedule)}
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
                              onMouseOver={(e) => {
                                e.currentTarget.style.backgroundColor = 'var(--bg-hover)';
                                e.currentTarget.style.borderColor = 'var(--border-strong)';
                              }}
                              onMouseOut={(e) => {
                                e.currentTarget.style.backgroundColor = 'var(--bg-secondary)';
                                e.currentTarget.style.borderColor = 'var(--border-subtle)';
                              }}
                              title="Edit schedule"
                            >
                              ✏️
                            </button>
                            <button
                              onClick={() => handleDelete(schedule.id, schedule.name)}
                              disabled={deleteMutation.isPending}
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
                              onMouseOver={(e) => {
                                if (!deleteMutation.isPending) {
                                  e.currentTarget.style.backgroundColor = 'var(--bg-hover)';
                                  e.currentTarget.style.borderColor = 'var(--border-strong)';
                                }
                              }}
                              onMouseOut={(e) => {
                                e.currentTarget.style.backgroundColor = 'var(--bg-secondary)';
                                e.currentTarget.style.borderColor = 'var(--border-subtle)';
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
          </div>
        </>
      )}

      {/* Snapshots Tab Content */}
      {activeTab === 'snapshots' && (
        <div style={{
          backgroundColor: 'white',
          padding: '1.5rem',
          borderRadius: '8px',
          boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
        }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
            <h2 style={{ margin: 0 }}>Existing Snapshots ({filteredSnapshots?.length || 0})</h2>
            <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
              <input
                type="text"
                placeholder="Search by name or ID..."
                value={snapshotSearch}
                onChange={(e) => setSnapshotSearch(e.target.value)}
                style={{
                  padding: '0.5rem',
                  border: '1px solid #ddd',
                  borderRadius: '4px',
                  width: '250px',
                }}
              />
              <button
                onClick={() => refetchSnapshots()}
                disabled={snapshotsLoading}
                style={{
                  padding: '0.5rem 1rem',
                  backgroundColor: '#17a2b8',
                  color: 'white',
                  border: 'none',
                  borderRadius: '4px',
                  cursor: snapshotsLoading ? 'not-allowed' : 'pointer',
                }}
              >
                {snapshotsLoading ? 'Refreshing...' : '↻ Refresh'}
              </button>
            </div>
          </div>

          {snapshotsLoading ? (
            <p>Loading snapshots...</p>
          ) : !filteredSnapshots || filteredSnapshots.length === 0 ? (
            <div style={{
              padding: '2rem',
              textAlign: 'center',
              color: '#666',
              backgroundColor: '#f8f9fa',
              borderRadius: '4px',
            }}>
              <p style={{ margin: 0 }}>
                {snapshotSearch ? 'No snapshots match your search.' : 'No snapshots found for this volume group.'}
              </p>
              <p style={{ marginTop: '0.5rem', fontSize: '0.9rem' }}>
                {!snapshotSearch && 'Snapshots will appear here after schedules are executed.'}
              </p>
            </div>
          ) : (
            <div style={{ overflowX: 'auto' }}>
              <table style={{
                width: '100%',
                borderCollapse: 'collapse',
                fontSize: '0.95rem',
              }}>
                <thead>
                  <tr style={{ backgroundColor: '#f8f9fa', borderBottom: '2px solid #dee2e6' }}>
                    <th style={{ padding: '0.75rem', textAlign: 'left' }}>Name</th>
                    <th style={{ padding: '0.75rem', textAlign: 'left' }}>Partition</th>
                    <th style={{ padding: '0.75rem', textAlign: 'right' }}>Provisioned</th>
                    <th style={{ padding: '0.75rem', textAlign: 'right' }}>Written</th>
                    <th style={{ padding: '0.75rem', textAlign: 'center' }}>Safeguarded</th>
                    <th style={{ padding: '0.75rem', textAlign: 'center' }}>HA Status</th>
                    <th style={{ padding: '0.75rem', textAlign: 'left' }}>Time Created</th>
                    <th style={{ padding: '0.75rem', textAlign: 'left' }}>Expiration Time</th>
                    <th style={{ padding: '0.75rem', textAlign: 'left' }}>ID</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredSnapshots.map((snapshot: any) => (
                    <tr key={snapshot.id} style={{ borderBottom: '1px solid #dee2e6' }}>
                      <td style={{ padding: '0.75rem' }}>
                        <strong>{snapshot.name || 'N/A'}</strong>
                      </td>
                      <td style={{ padding: '0.75rem', fontSize: '0.9rem', color: '#666' }}>
                        {snapshot.partition_name || 'N/A'}
                      </td>
                      <td style={{ padding: '0.75rem', textAlign: 'right', fontFamily: 'monospace' }}>
                        {formatCapacity(snapshot.protection_provisioned_capacity)}
                      </td>
                      <td style={{ padding: '0.75rem', textAlign: 'right', fontFamily: 'monospace' }}>
                        {formatCapacity(snapshot.protection_written_capacity)}
                      </td>
                      <td style={{ padding: '0.75rem', textAlign: 'center' }}>
                        {snapshot.safeguarded === 'yes' ? (
                          <span style={{
                            padding: '0.25rem 0.75rem',
                            backgroundColor: '#d4edda',
                            color: '#155724',
                            borderRadius: '12px',
                            fontSize: '0.85rem',
                            fontWeight: 'bold',
                          }}>
                            Yes
                          </span>
                        ) : (
                          <span style={{
                            padding: '0.25rem 0.75rem',
                            backgroundColor: '#f8f9fa',
                            color: '#666',
                            borderRadius: '12px',
                            fontSize: '0.85rem',
                          }}>
                            No
                          </span>
                        )}
                      </td>
                      <td style={{ padding: '0.75rem', textAlign: 'center' }}>
                        <span style={{
                          padding: '0.25rem 0.75rem',
                          backgroundColor: snapshot.ha_state === 'safeguarded' ? '#d4edda' : '#f8f9fa',
                          color: snapshot.ha_state === 'safeguarded' ? '#155724' : '#666',
                          borderRadius: '12px',
                          fontSize: '0.85rem',
                        }}>
                          {snapshot.ha_state || 'local'}
                        </span>
                      </td>
                      <td style={{ padding: '0.75rem', fontSize: '0.9rem', color: '#666' }}>
                        {formatTimestamp(snapshot.time)}
                      </td>
                      <td style={{ padding: '0.75rem', fontSize: '0.9rem', color: '#666' }}>
                        {formatTimestamp(snapshot.expiration_time)}
                      </td>
                      <td style={{ padding: '0.75rem', fontFamily: 'monospace', fontSize: '0.9rem', color: '#666' }}>
                        {snapshot.id}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

// 