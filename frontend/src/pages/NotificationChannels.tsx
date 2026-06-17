import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { notificationsApi } from '../api/services';
import { useState } from 'react';
import type { NotificationChannel } from '../types';

export default function NotificationChannels() {
  const queryClient = useQueryClient();
  const [showForm, setShowForm] = useState(false);
  const [editingChannel, setEditingChannel] = useState<NotificationChannel | null>(null);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const [testResults, setTestResults] = useState<{ [key: number]: { type: 'success' | 'error'; text: string } }>({});
  
  const [formData, setFormData] = useState({
    name: '',
    type: 'email' as 'email' | 'slack' | 'webhook' | 'snmp',
    description: '',
    config: {} as Record<string, any>,
  });

  const { data: channels, isLoading } = useQuery({
    queryKey: ['notification-channels'],
    queryFn: notificationsApi.listChannels,
  });

  const createMutation = useMutation({
    mutationFn: notificationsApi.createChannel,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notification-channels'] });
      setShowForm(false);
      resetForm();
      setMessage({ type: 'success', text: 'Channel created successfully' });
      setTimeout(() => setMessage(null), 3000);
    },
    onError: (error: any) => {
      setMessage({ type: 'error', text: error.response?.data?.error || 'Failed to create channel' });
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: any }) => notificationsApi.updateChannel(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notification-channels'] });
      setShowForm(false);
      setEditingChannel(null);
      resetForm();
      setMessage({ type: 'success', text: 'Channel updated successfully' });
      setTimeout(() => setMessage(null), 3000);
    },
    onError: (error: any) => {
      setMessage({ type: 'error', text: error.response?.data?.error || 'Failed to update channel' });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: notificationsApi.deleteChannel,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notification-channels'] });
      setMessage({ type: 'success', text: 'Channel deleted successfully' });
      setTimeout(() => setMessage(null), 3000);
    },
    onError: (error: any) => {
      setMessage({ type: 'error', text: error.response?.data?.error || 'Failed to delete channel' });
    },
  });

  const testMutation = useMutation({
    mutationFn: notificationsApi.testChannel,
    onSuccess: (_, channelId) => {
      setTestResults(prev => ({ ...prev, [channelId]: { type: 'success', text: 'Test notification sent successfully' } }));
      setTimeout(() => {
        setTestResults(prev => {
          const newResults = { ...prev };
          delete newResults[channelId];
          return newResults;
        });
      }, 5000);
    },
    onError: (error: any, channelId) => {
      setTestResults(prev => ({ ...prev, [channelId]: { type: 'error', text: error.response?.data?.error || 'Test failed' } }));
      setTimeout(() => {
        setTestResults(prev => {
          const newResults = { ...prev };
          delete newResults[channelId];
          return newResults;
        });
      }, 5000);
    },
  });

  const resetForm = () => {
    setFormData({
      name: '',
      type: 'email',
      description: '',
      config: {},
    });
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (editingChannel) {
      updateMutation.mutate({
        id: editingChannel.id,
        data: {
          name: formData.name,
          type: formData.type,
          config: formData.config,
        },
      });
    } else {
      createMutation.mutate({
        name: formData.name,
        type: formData.type,
        config: formData.config,
        description: formData.description,
      });
    }
  };

  const handleEdit = (channel: NotificationChannel) => {
    setEditingChannel(channel);
    setFormData({
      name: channel.name,
      type: channel.type,
      description: '',
      config: JSON.parse(channel.config || '{}'),
    });
    setShowForm(true);
  };

  const handleDelete = (id: number) => {
    if (confirm('Are you sure you want to delete this channel?')) {
      deleteMutation.mutate(id);
    }
  };

  const handleTest = (id: number) => {
    testMutation.mutate(id);
  };

  const renderConfigForm = () => {
    switch (formData.type) {
      case 'email':
        return (
          <>
            <div>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>SMTP Host *</label>
              <input
                type="text"
                value={formData.config.smtp_host || ''}
                onChange={(e) => setFormData({ ...formData, config: { ...formData.config, smtp_host: e.target.value } })}
                style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
                required
              />
            </div>
            <div>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>SMTP Port *</label>
              <input
                type="number"
                value={formData.config.smtp_port || 587}
                onChange={(e) => setFormData({ ...formData, config: { ...formData.config, smtp_port: parseInt(e.target.value) } })}
                style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
                required
              />
            </div>
            <div>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>Username *</label>
              <input
                type="text"
                value={formData.config.username || ''}
                onChange={(e) => setFormData({ ...formData, config: { ...formData.config, username: e.target.value } })}
                style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
                required
              />
            </div>
            <div>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>Password *</label>
              <input
                type="password"
                value={formData.config.password_encrypted || ''}
                onChange={(e) => setFormData({ ...formData, config: { ...formData.config, password_encrypted: e.target.value } })}
                style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
                required
              />
            </div>
            <div>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>From Address *</label>
              <input
                type="email"
                value={formData.config.from || ''}
                onChange={(e) => setFormData({ ...formData, config: { ...formData.config, from: e.target.value } })}
                style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
                required
              />
            </div>
            <div>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>To Addresses (comma-separated) *</label>
              <input
                type="text"
                value={Array.isArray(formData.config.to) ? formData.config.to.join(', ') : ''}
                onChange={(e) => setFormData({ ...formData, config: { ...formData.config, to: e.target.value.split(',').map(s => s.trim()) } })}
                style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
                placeholder="email1@example.com, email2@example.com"
                required
              />
            </div>
            <div style={{ display: 'flex', gap: '1rem' }}>
              <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <input
                  type="checkbox"
                  checked={formData.config.use_tls || false}
                  onChange={(e) => setFormData({ ...formData, config: { ...formData.config, use_tls: e.target.checked } })}
                />
                Use TLS
              </label>
              <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <input
                  type="checkbox"
                  checked={formData.config.skip_verify || false}
                  onChange={(e) => setFormData({ ...formData, config: { ...formData.config, skip_verify: e.target.checked } })}
                />
                Skip TLS Verify
              </label>
            </div>
          </>
        );
      
      case 'slack':
        return (
          <>
            <div>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>Webhook URL *</label>
              <input
                type="url"
                value={formData.config.webhook_url_encrypted || ''}
                onChange={(e) => setFormData({ ...formData, config: { ...formData.config, webhook_url_encrypted: e.target.value } })}
                style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
                placeholder="https://hooks.slack.com/services/..."
                required
              />
            </div>
            <div>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>Channel</label>
              <input
                type="text"
                value={formData.config.channel || ''}
                onChange={(e) => setFormData({ ...formData, config: { ...formData.config, channel: e.target.value } })}
                style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
                placeholder="#alerts"
              />
            </div>
            <div>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>Username</label>
              <input
                type="text"
                value={formData.config.username || ''}
                onChange={(e) => setFormData({ ...formData, config: { ...formData.config, username: e.target.value } })}
                style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
                placeholder="Snapshot Manager"
              />
            </div>
          </>
        );
      
      case 'webhook':
        return (
          <>
            <div>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>Webhook URL *</label>
              <input
                type="url"
                value={formData.config.url || ''}
                onChange={(e) => setFormData({ ...formData, config: { ...formData.config, url: e.target.value } })}
                style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
                required
              />
            </div>
            <div>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>Method</label>
              <select
                value={formData.config.method || 'POST'}
                onChange={(e) => setFormData({ ...formData, config: { ...formData.config, method: e.target.value } })}
                style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
              >
                <option value="POST">POST</option>
                <option value="PUT">PUT</option>
                <option value="PATCH">PATCH</option>
              </select>
            </div>
            <div>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>Timeout (seconds)</label>
              <input
                type="number"
                value={formData.config.timeout || 30}
                onChange={(e) => setFormData({ ...formData, config: { ...formData.config, timeout: parseInt(e.target.value) } })}
                style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
              />
            </div>
          </>
        );
      
      case 'snmp':
        return (
          <>
            <div>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>SNMP Host *</label>
              <input
                type="text"
                value={formData.config.host || ''}
                onChange={(e) => setFormData({ ...formData, config: { ...formData.config, host: e.target.value } })}
                style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
                required
              />
            </div>
            <div>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>Port</label>
              <input
                type="number"
                value={formData.config.port || 162}
                onChange={(e) => setFormData({ ...formData, config: { ...formData.config, port: parseInt(e.target.value) } })}
                style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
              />
            </div>
            <div>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>Community String *</label>
              <input
                type="text"
                value={formData.config.community || ''}
                onChange={(e) => setFormData({ ...formData, config: { ...formData.config, community: e.target.value } })}
                style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
                required
              />
            </div>
          </>
        );
    }
  };

  if (isLoading) {
    return <div style={{ padding: '2rem' }}>Loading channels...</div>;
  }

  return (
    <div style={{ padding: '2rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
        <h1>Notification Channels</h1>
        <button
          onClick={() => {
            resetForm();
            setEditingChannel(null);
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
          + Add Channel
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
          <h2>{editingChannel ? 'Edit Channel' : 'New Channel'}</h2>
          <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            <div>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>Channel Name *</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
                required
              />
            </div>

            <div>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>Channel Type *</label>
              <select
                value={formData.type}
                onChange={(e) => setFormData({ ...formData, type: e.target.value as any, config: {} })}
                style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
                disabled={!!editingChannel}
              >
                <option value="email">Email</option>
                <option value="slack">Slack</option>
                <option value="webhook">Webhook</option>
                <option value="snmp">SNMP</option>
              </select>
            </div>

            {renderConfigForm()}

            <div style={{ display: 'flex', gap: '1rem', marginTop: '1rem' }}>
              <button
                type="submit"
                style={{
                  padding: '0.5rem 1rem',
                  backgroundColor: '#28a745',
                  color: 'white',
                  border: 'none',
                  borderRadius: '4px',
                  cursor: 'pointer',
                }}
              >
                {editingChannel ? 'Update' : 'Create'}
              </button>
              <button
                type="button"
                onClick={() => {
                  setShowForm(false);
                  setEditingChannel(null);
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
        {channels?.map((channel) => (
          <div
            key={channel.id}
            style={{
              padding: '1.5rem',
              border: '1px solid #ddd',
              borderRadius: '8px',
              backgroundColor: 'white',
            }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start' }}>
              <div style={{ flex: 1 }}>
                <h3 style={{ margin: '0 0 0.5rem 0' }}>{channel.name}</h3>
                <div style={{ display: 'flex', gap: '1rem', fontSize: '0.9rem', color: '#666' }}>
                  <span>Type: <strong>{channel.type}</strong></span>
                  <span>Status: <strong style={{ color: channel.is_active ? '#28a745' : '#dc3545' }}>
                    {channel.is_active ? 'Active' : 'Inactive'}
                  </strong></span>
                  <span>Created: {new Date(channel.created_at).toLocaleDateString()}</span>
                </div>
                {testResults[channel.id] && (
                  <div style={{
                    marginTop: '0.5rem',
                    padding: '0.5rem',
                    backgroundColor: testResults[channel.id].type === 'success' ? '#d4edda' : '#f8d7da',
                    color: testResults[channel.id].type === 'success' ? '#155724' : '#721c24',
                    borderRadius: '4px',
                    fontSize: '0.9rem',
                  }}>
                    {testResults[channel.id].text}
                  </div>
                )}
              </div>
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                <button
                  onClick={() => handleTest(channel.id)}
                  disabled={testMutation.isPending}
                  style={{
                    padding: '0.5rem 1rem',
                    backgroundColor: '#17a2b8',
                    color: 'white',
                    border: 'none',
                    borderRadius: '4px',
                    cursor: 'pointer',
                  }}
                >
                  Test
                </button>
                <button
                  onClick={() => handleEdit(channel)}
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
                  onClick={() => handleDelete(channel.id)}
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

      {channels?.length === 0 && (
        <div style={{ textAlign: 'center', padding: '3rem', color: '#666' }}>
          <p>No notification channels configured.</p>
          <p>Click "Add Channel" to create your first notification channel.</p>
        </div>
      )}
    </div>
  );
}

// 