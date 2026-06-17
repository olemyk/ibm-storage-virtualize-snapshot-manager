import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { notificationsApi } from '../api/services';
import { useState } from 'react';
import type { AlertRule } from '../types';

export default function AlertRules() {
  const queryClient = useQueryClient();
  const [showForm, setShowForm] = useState(false);
  const [editingRule, setEditingRule] = useState<AlertRule | null>(null);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    is_active: true,
    event_type: 'snapshot_failure',
    severity: 'error' as 'info' | 'warning' | 'error' | 'critical',
    notification_channel_ids: [] as number[],
    throttle_minutes: 60,
  });

  const { data: rules, isLoading: rulesLoading } = useQuery({
    queryKey: ['alert-rules'],
    queryFn: notificationsApi.listRules,
  });

  const { data: channels } = useQuery({
    queryKey: ['notification-channels'],
    queryFn: notificationsApi.listChannels,
  });

  const createMutation = useMutation({
    mutationFn: notificationsApi.createRule,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alert-rules'] });
      setShowForm(false);
      resetForm();
      setMessage({ type: 'success', text: 'Alert rule created successfully' });
      setTimeout(() => setMessage(null), 3000);
    },
    onError: (error: any) => {
      setMessage({ type: 'error', text: error.response?.data?.error || 'Failed to create rule' });
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: any }) => notificationsApi.updateRule(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alert-rules'] });
      setShowForm(false);
      setEditingRule(null);
      resetForm();
      setMessage({ type: 'success', text: 'Alert rule updated successfully' });
      setTimeout(() => setMessage(null), 3000);
    },
    onError: (error: any) => {
      setMessage({ type: 'error', text: error.response?.data?.error || 'Failed to update rule' });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: notificationsApi.deleteRule,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alert-rules'] });
      setMessage({ type: 'success', text: 'Alert rule deleted successfully' });
      setTimeout(() => setMessage(null), 3000);
    },
    onError: (error: any) => {
      setMessage({ type: 'error', text: error.response?.data?.error || 'Failed to delete rule' });
    },
  });

  const resetForm = () => {
    setFormData({
      name: '',
      description: '',
      is_active: true,
      event_type: 'snapshot_failure',
      severity: 'error',
      notification_channel_ids: [],
      throttle_minutes: 60,
    });
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    const ruleData = {
      ...formData,
      notification_channel_ids: JSON.stringify(formData.notification_channel_ids),
    };

    if (editingRule) {
      updateMutation.mutate({ id: editingRule.id, data: ruleData });
    } else {
      createMutation.mutate(ruleData);
    }
  };

  const handleEdit = (rule: AlertRule) => {
    setEditingRule(rule);
    const channelIds = JSON.parse(rule.notification_channel_ids || '[]');
    setFormData({
      name: rule.name,
      description: rule.description || '',
      is_active: rule.is_active,
      event_type: rule.event_type,
      severity: rule.severity,
      notification_channel_ids: channelIds,
      throttle_minutes: rule.throttle_minutes,
    });
    setShowForm(true);
  };

  const handleDelete = (id: number) => {
    if (confirm('Are you sure you want to delete this alert rule?')) {
      deleteMutation.mutate(id);
    }
  };

  const toggleChannelSelection = (channelId: number) => {
    setFormData(prev => ({
      ...prev,
      notification_channel_ids: prev.notification_channel_ids.includes(channelId)
        ? prev.notification_channel_ids.filter(id => id !== channelId)
        : [...prev.notification_channel_ids, channelId]
    }));
  };

  const getChannelNames = (channelIdsJson: string) => {
    try {
      const ids = JSON.parse(channelIdsJson);
      return ids
        .map((id: number) => channels?.find(c => c.id === id)?.name)
        .filter(Boolean)
        .join(', ') || 'None';
    } catch {
      return 'None';
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'info': return '#17a2b8';
      case 'warning': return '#ffc107';
      case 'error': return '#dc3545';
      case 'critical': return '#721c24';
      default: return '#6c757d';
    }
  };

  if (rulesLoading) {
    return <div style={{ padding: '2rem' }}>Loading alert rules...</div>;
  }

  return (
    <div style={{ padding: '2rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
        <h1>Alert Rules</h1>
        <button
          onClick={() => {
            resetForm();
            setEditingRule(null);
            setShowForm(true);
          }}
          style={{
            padding: '0.5rem 1rem',
            backgroundColor: '#007bff',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
          }}
        >
          + Add Rule
        </button>
      </div>

      {message && (
        <div style={{
          padding: '1rem',
          marginBottom: '1rem',
          backgroundColor: message.type === 'success' ? '#d4edda' : '#f8d7da',
          color: message.type === 'success' ? '#155724' : '#721c24',
          borderRadius: '4px',
        }}>
          {message.text}
        </div>
      )}

      {showForm && (
        <div style={{
          marginBottom: '2rem',
          padding: '1.5rem',
          border: '1px solid #ddd',
          borderRadius: '8px',
          backgroundColor: '#f8f9fa',
        }}>
          <h2>{editingRule ? 'Edit Alert Rule' : 'New Alert Rule'}</h2>
          <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            <div>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>Rule Name *</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
                placeholder="e.g., Critical Snapshot Failures"
                required
              />
            </div>

            <div>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>Description</label>
              <textarea
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px', minHeight: '80px' }}
                placeholder="Describe when this rule should trigger..."
              />
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
              <div>
                <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>Event Type *</label>
                <select
                  value={formData.event_type}
                  onChange={(e) => setFormData({ ...formData, event_type: e.target.value })}
                  style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
                  required
                >
                  <option value="snapshot_success">Snapshot Success</option>
                  <option value="snapshot_failure">Snapshot Failure</option>
                  <option value="snapshot_warning">Snapshot Warning</option>
                  <option value="system_connection_lost">System Connection Lost</option>
                  <option value="scheduler_error">Scheduler Error</option>
                  <option value="consecutive_failures">Consecutive Failures</option>
                </select>
              </div>

              <div>
                <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>Minimum Severity *</label>
                <select
                  value={formData.severity}
                  onChange={(e) => setFormData({ ...formData, severity: e.target.value as any })}
                  style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
                  required
                >
                  <option value="info">Info</option>
                  <option value="warning">Warning</option>
                  <option value="error">Error</option>
                  <option value="critical">Critical</option>
                </select>
              </div>
            </div>

            <div>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>Throttle (minutes)</label>
              <input
                type="number"
                value={formData.throttle_minutes}
                onChange={(e) => setFormData({ ...formData, throttle_minutes: parseInt(e.target.value) || 0 })}
                style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
                min="0"
                placeholder="Minimum time between notifications (0 = no throttle)"
              />
              <small style={{ color: '#666' }}>Prevents notification spam by limiting how often this rule can trigger</small>
            </div>

            <div>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>Notification Channels *</label>
              <div style={{ 
                border: '1px solid #ddd', 
                borderRadius: '4px', 
                padding: '1rem',
                backgroundColor: 'white',
                maxHeight: '200px',
                overflowY: 'auto'
              }}>
                {channels && channels.length > 0 ? (
                  channels.map(channel => (
                    <label
                      key={channel.id}
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '0.5rem',
                        padding: '0.5rem',
                        cursor: 'pointer',
                        borderRadius: '4px',
                      }}
                    >
                      <input
                        type="checkbox"
                        checked={formData.notification_channel_ids.includes(channel.id)}
                        onChange={() => toggleChannelSelection(channel.id)}
                      />
                      <span>{channel.name}</span>
                      <span style={{ 
                        marginLeft: 'auto', 
                        fontSize: '0.85rem', 
                        color: '#666',
                        textTransform: 'uppercase'
                      }}>
                        {channel.type}
                      </span>
                    </label>
                  ))
                ) : (
                  <p style={{ color: '#666', margin: 0 }}>
                    No channels available. <a href="/notifications/channels">Create a channel first</a>.
                  </p>
                )}
              </div>
              {formData.notification_channel_ids.length === 0 && (
                <small style={{ color: '#dc3545' }}>Please select at least one channel</small>
              )}
            </div>

            <div>
              <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <input
                  type="checkbox"
                  checked={formData.is_active}
                  onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                />
                <span style={{ fontWeight: 500 }}>Active</span>
              </label>
              <small style={{ color: '#666', marginLeft: '1.5rem' }}>
                Inactive rules will not trigger notifications
              </small>
            </div>

            <div style={{ display: 'flex', gap: '1rem', marginTop: '1rem' }}>
              <button
                type="submit"
                disabled={formData.notification_channel_ids.length === 0}
                style={{
                  padding: '0.5rem 1rem',
                  backgroundColor: formData.notification_channel_ids.length === 0 ? '#ccc' : '#28a745',
                  color: 'white',
                  border: 'none',
                  borderRadius: '4px',
                  cursor: formData.notification_channel_ids.length === 0 ? 'not-allowed' : 'pointer',
                }}
              >
                {editingRule ? 'Update' : 'Create'}
              </button>
              <button
                type="button"
                onClick={() => {
                  setShowForm(false);
                  setEditingRule(null);
                  resetForm();
                }}
                style={{
                  padding: '0.5rem 1rem',
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

      <div style={{ display: 'grid', gap: '1rem' }}>
        {rules?.map((rule) => (
          <div
            key={rule.id}
            style={{
              padding: '1.5rem',
              border: '1px solid #ddd',
              borderRadius: '8px',
              backgroundColor: 'white',
              opacity: rule.is_active ? 1 : 0.6,
            }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start' }}>
              <div style={{ flex: 1 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', marginBottom: '0.5rem' }}>
                  <h3 style={{ margin: 0 }}>{rule.name}</h3>
                  <span style={{
                    padding: '0.25rem 0.5rem',
                    borderRadius: '4px',
                    fontSize: '0.75rem',
                    fontWeight: 'bold',
                    color: 'white',
                    backgroundColor: getSeverityColor(rule.severity),
                    textTransform: 'uppercase',
                  }}>
                    {rule.severity}
                  </span>
                  {!rule.is_active && (
                    <span style={{
                      padding: '0.25rem 0.5rem',
                      borderRadius: '4px',
                      fontSize: '0.75rem',
                      backgroundColor: '#6c757d',
                      color: 'white',
                    }}>
                      INACTIVE
                    </span>
                  )}
                </div>
                
                {rule.description && (
                  <p style={{ margin: '0.5rem 0', color: '#666' }}>{rule.description}</p>
                )}
                
                <div style={{ display: 'grid', gridTemplateColumns: 'auto 1fr', gap: '0.5rem 1rem', fontSize: '0.9rem', marginTop: '1rem' }}>
                  <strong>Event:</strong>
                  <span>{rule.event_type.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}</span>
                  
                  <strong>Channels:</strong>
                  <span>{getChannelNames(rule.notification_channel_ids)}</span>
                  
                  <strong>Throttle:</strong>
                  <span>{rule.throttle_minutes > 0 ? `${rule.throttle_minutes} minutes` : 'None'}</span>
                  
                  {rule.last_triggered_at && (
                    <>
                      <strong>Last Triggered:</strong>
                      <span>{new Date(rule.last_triggered_at).toLocaleString()}</span>
                    </>
                  )}
                </div>
              </div>
              
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                <button
                  onClick={() => handleEdit(rule)}
                  style={{
                    padding: '0.5rem 1rem',
                    backgroundColor: '#ffc107',
                    color: 'black',
                    border: 'none',
                    borderRadius: '4px',
                    cursor: 'pointer',
                  }}
                >
                  Edit
                </button>
                <button
                  onClick={() => handleDelete(rule.id)}
                  style={{
                    padding: '0.5rem 1rem',
                    backgroundColor: '#dc3545',
                    color: 'white',
                    border: 'none',
                    borderRadius: '4px',
                    cursor: 'pointer',
                  }}
                >
                  Delete
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>

      {rules?.length === 0 && (
        <div style={{ textAlign: 'center', padding: '3rem', color: '#666' }}>
          <p>No alert rules configured.</p>
          <p>Click "Add Rule" to create your first alert rule.</p>
        </div>
      )}
    </div>
  );
}

// 