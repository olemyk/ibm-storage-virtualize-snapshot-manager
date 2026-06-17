import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { systemsApi } from '../api/services';
import { useState, useMemo, useEffect } from 'react';
import type { FormEvent } from 'react';
import type { StorageSystem } from '../types';

// Constants
const REFETCH_INTERVAL_MS = 60000; // 60 seconds
const INITIAL_HEALTH_CHECK_DELAY_MS = 5000; // 5 seconds
const HEALTH_CHECK_INTERVAL_MS = 60000; // 60 seconds
const MESSAGE_DISPLAY_DURATION_MS = 5000; // 5 seconds
const ERROR_MESSAGE_DISPLAY_DURATION_MS = 10000; // 10 seconds
const DEFAULT_PORT = 7443;

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

export default function Systems() {
  const queryClient = useQueryClient();
  const [showForm, setShowForm] = useState(false);
  const [editingSystem, setEditingSystem] = useState<StorageSystem | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const [testResults, setTestResults] = useState<{ [key: number]: { type: 'success' | 'error'; text: string } }>({});
  const [confirmDialog, setConfirmDialog] = useState<{ isOpen: boolean; systemId?: number; systemName?: string }>({ isOpen: false });
  const [formData, setFormData] = useState({
    name: '',
    ip_address: '',
    port: 7443,
    username: '',
    password: '',
    skip_tls_verify: true, // Default to true for backward compatibility
  });

  const { data: systems, isLoading, refetch } = useQuery({
    queryKey: ['systems'],
    queryFn: systemsApi.list,
    refetchInterval: REFETCH_INTERVAL_MS,
    staleTime: 0, // Always consider data stale
    refetchOnMount: true,
    refetchOnWindowFocus: true,
  });

  // Automatic health check every 60 seconds
  useEffect(() => {
    console.log('[Health Check] Setting up automatic health check...');
    let isActive = true;
    
    const runHealthCheck = async () => {
      if (!isActive) return;
      console.log('[Health Check] Running health check at', new Date().toLocaleTimeString());
      try {
        const result = await systemsApi.checkHealth();
        console.log('[Health Check] Completed:', result);
        if (isActive) {
          console.log('[Health Check] Refetching systems list...');
          const refetchResult = await refetch();
          console.log('[Health Check] Systems list refetched, new data:', refetchResult.data);
        }
      } catch (error) {
        console.error('[Health Check] Failed:', error);
      }
    };

    // Run initial health check after 5 seconds
    const initialCheck = setTimeout(() => {
      console.log('[Health Check] Running initial check...');
      runHealthCheck();
    }, INITIAL_HEALTH_CHECK_DELAY_MS);

    // Set up interval for recurring checks
    const healthCheckInterval = setInterval(() => {
      console.log('[Health Check] Running scheduled check...');
      runHealthCheck();
    }, HEALTH_CHECK_INTERVAL_MS);

    return () => {
      console.log('[Health Check] Cleaning up timers');
      isActive = false;
      clearInterval(healthCheckInterval);
      clearTimeout(initialCheck);
    };
  }, [queryClient]);

  const createMutation = useMutation({
    mutationFn: systemsApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['systems'] });
      setShowForm(false);
    setFormData({ name: '', ip_address: '', port: DEFAULT_PORT, username: '', password: '', skip_tls_verify: true });
      setMessage({ type: 'success', text: 'Storage system added successfully! Use the Test button to verify the connection.' });
      setTimeout(() => setMessage(null), MESSAGE_DISPLAY_DURATION_MS);
    },
    onError: (error: any) => {
      setMessage({ type: 'error', text: error.response?.data?.error || error.message });
      setTimeout(() => setMessage(null), MESSAGE_DISPLAY_DURATION_MS);
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<typeof formData> }) =>
      systemsApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['systems'] });
      setShowForm(false);
      setEditingSystem(null);
      setFormData({ name: '', ip_address: '', port: 7443, username: '', password: '', skip_tls_verify: true });
      setMessage({ type: 'success', text: 'Storage system updated successfully!' });
      setTimeout(() => setMessage(null), MESSAGE_DISPLAY_DURATION_MS);
    },
    onError: (error: any) => {
      setMessage({ type: 'error', text: error.response?.data?.error || error.message });
      setTimeout(() => setMessage(null), 5000);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: systemsApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['systems'] });
      setMessage({ type: 'success', text: 'Storage system deleted successfully!' });
      setTimeout(() => setMessage(null), MESSAGE_DISPLAY_DURATION_MS);
    },
    onError: (error: any) => {
      setMessage({ type: 'error', text: error.response?.data?.error || error.message });
      setTimeout(() => setMessage(null), 5000);
    },
  });

  const testMutation = useMutation({
    mutationFn: systemsApi.test,
    onSuccess: (_data, systemId) => {
      setTestResults(prev => ({ ...prev, [systemId]: { type: 'success', text: 'Connection successful!' } }));
      setTimeout(() => setTestResults(prev => { const newResults = { ...prev }; delete newResults[systemId]; return newResults; }), MESSAGE_DISPLAY_DURATION_MS);
    },
    onError: (error: any, systemId) => {
      const errorMsg = error.response?.data?.error || error.message;
      setTestResults(prev => ({ ...prev, [systemId]: { type: 'error', text: errorMsg } }));
      setTimeout(() => setTestResults(prev => { const newResults = { ...prev }; delete newResults[systemId]; return newResults; }), ERROR_MESSAGE_DISPLAY_DURATION_MS);
    },
  });

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (editingSystem) {
      // Only send fields that have values (password is optional on update)
      const updateData: Partial<typeof formData> = {
        name: formData.name,
        ip_address: formData.ip_address,
        port: formData.port,
        username: formData.username,
        skip_tls_verify: formData.skip_tls_verify,
      };
      // Only include password if it was changed
      if (formData.password) {
        updateData.password = formData.password;
      }
      updateMutation.mutate({ id: editingSystem.id, data: updateData });
    } else {
      createMutation.mutate(formData);
    }
  };

  const handleEdit = (system: StorageSystem) => {
    setEditingSystem(system);
    setFormData({
      name: system.name,
      ip_address: system.ip_address,
      port: system.port,
      username: system.username,
      password: '', // Don't populate password for security
      skip_tls_verify: system.skip_tls_verify,
    });
    setShowForm(true);
  };

  const handleCancelEdit = () => {
    setShowForm(false);
    setEditingSystem(null);
    setFormData({ name: '', ip_address: '', port: 7443, username: '', password: '', skip_tls_verify: true });
  };

  // Filter systems based on search query (memoized)
  const filteredSystems = useMemo(() =>
    systems?.filter((system) => {
      if (!searchQuery) return true;
      const query = searchQuery.toLowerCase();
      return (
        system.name.toLowerCase().includes(query) ||
        system.ip_address.toLowerCase().includes(query) ||
        system.username.toLowerCase().includes(query)
      );
    }) || [],
    [systems, searchQuery]
  );

  if (isLoading) {
    return <div style={{ padding: '2rem' }}>Loading systems...</div>;
  }

  return (
    <>
      <ConfirmDialog
        isOpen={confirmDialog.isOpen}
        title="Delete Storage System"
        message={`Are you sure you want to delete system "${confirmDialog.systemName}"? This action cannot be undone.`}
        onConfirm={() => {
          if (confirmDialog.systemId) {
            deleteMutation.mutate(confirmDialog.systemId);
          }
          setConfirmDialog({ isOpen: false });
        }}
        onCancel={() => setConfirmDialog({ isOpen: false })}
        confirmText="Delete"
        cancelText="Cancel"
      />
      <style>{`
        .search-input:focus {
          border-color: #0066cc !important;
        }
      `}</style>
      <div style={{ padding: '2rem' }}>
      {/* Info banner about status */}
      <div style={{
        padding: '1rem',
        marginBottom: '1rem',
        borderRadius: '4px',
        backgroundColor: '#d1ecf1',
        color: '#0c5460',
        border: '1px solid #bee5eb',
        fontSize: '0.9rem',
      }}>
        <strong>ℹ️ Automatic Health Monitoring:</strong> Connection status is automatically checked every 60 seconds.
        The status shows: <strong>✓ Connected</strong> (reachable), <strong>✗ Disconnected</strong> (unreachable), or <strong>? Unknown/Not Checked</strong> (pending check).
        You can also manually test connections using the <strong>Test</strong> button.
      </div>

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
      
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
        <h1>Storage Systems</h1>
        <div style={{ display: 'flex', gap: '1rem' }}>
          <button
            onClick={async () => {
              setMessage({ type: 'success', text: 'Running health check on all systems...' });
              try {
                const result = await systemsApi.checkHealth();
                console.log('[Manual Check] Health check completed, refetching...');
                await refetch();
                console.log('[Manual Check] Systems list updated');
                setMessage({
                  type: 'success',
                  text: `Health check completed: ${result.results.filter(r => r.status === 'connected').length} connected, ${result.results.filter(r => r.status === 'disconnected').length} disconnected`
                });
                setTimeout(() => setMessage(null), MESSAGE_DISPLAY_DURATION_MS);
              } catch (error: any) {
                setMessage({ type: 'error', text: 'Health check failed: ' + (error.response?.data?.error || error.message) });
                setTimeout(() => setMessage(null), MESSAGE_DISPLAY_DURATION_MS);
              }
            }}
            style={{
              padding: '0.75rem 1.5rem',
              backgroundColor: '#28a745',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
              fontWeight: 'bold',
            }}
          >
            🔄 Check All Now
          </button>
          <button
            onClick={() => {
              if (showForm) {
                handleCancelEdit();
              } else {
                setShowForm(true);
              }
            }}
            style={{
              padding: '0.75rem 1.5rem',
              backgroundColor: '#0066cc',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
              fontWeight: 'bold',
            }}
          >
            {showForm ? 'Cancel' : '+ Add System'}
          </button>
        </div>
      </div>

      {showForm && (
        <div style={{
          backgroundColor: 'white',
          padding: '2rem',
          borderRadius: '8px',
          boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
          marginBottom: '2rem',
        }}>
          <h2 style={{ marginBottom: '1.5rem' }}>
            {editingSystem ? 'Edit Storage System' : 'Add Storage System'}
          </h2>
          <form onSubmit={handleSubmit}>
            <div style={{ display: 'grid', gap: '1rem', marginBottom: '1.5rem' }}>
              <div>
                <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
                  System Name *
                </label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  required
                  style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
                />
              </div>
              <div>
                <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
                  IP Address *
                </label>
                <input
                  type="text"
                  value={formData.ip_address}
                  onChange={(e) => setFormData({ ...formData, ip_address: e.target.value })}
                  required
                  style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
                />
              </div>
              <div>
                <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
                  Port
                </label>
                <input
                  type="number"
                  value={formData.port}
                  onChange={(e) => setFormData({ ...formData, port: parseInt(e.target.value, 10) })}
                  style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
                />
              </div>
              <div>
                <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
                  Username *
                </label>
                <input
                  type="text"
                  value={formData.username}
                  onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                  required
                  style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
                />
              </div>
              <div>
                <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
                  Password {editingSystem ? '(leave blank to keep current)' : '*'}
                </label>
                <input
                  type="password"
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  required={!editingSystem}
                  placeholder={editingSystem ? 'Leave blank to keep current password' : ''}
                  style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
                />
              </div>
              <div>
                <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', cursor: 'pointer' }}>
                  <input
                    type="checkbox"
                    checked={formData.skip_tls_verify}
                    onChange={(e) => setFormData({ ...formData, skip_tls_verify: e.target.checked })}
                    style={{ width: '18px', height: '18px', cursor: 'pointer' }}
                  />
                  <span style={{ fontWeight: 'bold' }}>
                    Skip TLS/Certificate Verification
                  </span>
                </label>
                <p style={{ marginTop: '0.5rem', fontSize: '0.875rem', color: '#666', marginLeft: '26px' }}>
                  Enable this for systems with self-signed certificates.
                </p>
              </div>
            </div>
            <button
              type="submit"
              disabled={createMutation.isPending || updateMutation.isPending}
              style={{
                padding: '0.75rem 1.5rem',
                backgroundColor: (createMutation.isPending || updateMutation.isPending) ? '#ccc' : '#0066cc',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: (createMutation.isPending || updateMutation.isPending) ? 'not-allowed' : 'pointer',
                fontWeight: 'bold',
              }}
            >
              {editingSystem
                ? (updateMutation.isPending ? 'Updating...' : 'Update System')
                : (createMutation.isPending ? 'Adding...' : 'Add System')
              }
            </button>
          </form>
        </div>
      )}

      {/* Search Input */}
      <div style={{
        backgroundColor: 'white',
        padding: '1rem 1.5rem',
        borderRadius: '8px',
        boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
        marginBottom: '1rem',
      }}>
        <input
          type="text"
          placeholder="Search systems by name, IP address, or username..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="search-input"
          style={{
            width: '100%',
            padding: '0.75rem',
            border: '2px solid #e0e0e0',
            borderRadius: '6px',
            fontSize: '1rem',
            transition: 'border-color 0.2s',
            outline: 'none',
          }}
        />
        {searchQuery && (
          <div style={{ marginTop: '0.5rem', fontSize: '0.875rem', color: '#666' }}>
            Found {filteredSystems.length} system{filteredSystems.length !== 1 ? 's' : ''}
          </div>
        )}
      </div>

      <div style={{
        backgroundColor: 'white',
        padding: '1.5rem',
        borderRadius: '8px',
        boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
      }}>
        {filteredSystems.length > 0 ? (
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ borderBottom: '2px solid #ddd' }}>
                <th style={{ padding: '0.75rem', textAlign: 'left' }}>Name</th>
                <th style={{ padding: '0.75rem', textAlign: 'left' }}>IP Address</th>
                <th style={{ padding: '0.75rem', textAlign: 'left' }}>Port</th>
                <th style={{ padding: '0.75rem', textAlign: 'left' }}>Username</th>
                <th style={{ padding: '0.75rem', textAlign: 'left' }}>Status</th>
                <th style={{ padding: '0.75rem', textAlign: 'left' }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredSystems.map((system) => (
                <tr key={system.id} style={{ borderBottom: '1px solid #eee' }}>
                  <td style={{ padding: '0.75rem' }}>{system.name}</td>
                  <td style={{ padding: '0.75rem' }}>{system.ip_address}</td>
                  <td style={{ padding: '0.75rem' }}>{system.port}</td>
                  <td style={{ padding: '0.75rem' }}>{system.username}</td>
                  <td style={{ padding: '0.75rem' }}>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                      {system.connection_status ? (
                        <>
                          <span style={{
                            padding: '0.25rem 0.5rem',
                            borderRadius: '4px',
                            fontSize: '0.875rem',
                            backgroundColor:
                              system.connection_status === 'connected' ? '#d4edda' :
                              system.connection_status === 'disconnected' ? '#f8d7da' :
                              '#fff3cd',
                            color:
                              system.connection_status === 'connected' ? '#155724' :
                              system.connection_status === 'disconnected' ? '#721c24' :
                              '#856404',
                          }}>
                            {system.connection_status === 'connected' ? '✓ Connected' :
                             system.connection_status === 'disconnected' ? '✗ Disconnected' :
                             '? Unknown'}
                          </span>
                          {system.last_connection_check && (
                            <span style={{ fontSize: '0.7rem', color: '#999' }}>
                              Checked: {new Date(system.last_connection_check).toLocaleTimeString()}
                            </span>
                          )}
                          {system.connection_error && system.connection_status === 'disconnected' && (
                            <span style={{ fontSize: '0.7rem', color: '#dc3545', marginTop: '0.25rem' }}>
                              {system.connection_error.substring(0, 50)}...
                            </span>
                          )}
                        </>
                      ) : (
                        <>
                          <span style={{
                            padding: '0.25rem 0.5rem',
                            borderRadius: '4px',
                            fontSize: '0.875rem',
                            backgroundColor: '#fff3cd',
                            color: '#856404',
                          }}>
                            ? Not Checked
                          </span>
                          <span style={{ fontSize: '0.75rem', color: '#666' }}>
                            Waiting for health check...
                          </span>
                        </>
                      )}
                    </div>
                  </td>
                  <td style={{ padding: '0.75rem' }}>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                      <div>
                        <button
                          onClick={() => testMutation.mutate(system.id)}
                          disabled={testMutation.isPending}
                          style={{
                            padding: '0.5rem 1rem',
                            marginRight: '0.5rem',
                            backgroundColor: '#28a745',
                            color: 'white',
                            border: 'none',
                            borderRadius: '4px',
                            cursor: testMutation.isPending ? 'not-allowed' : 'pointer',
                            fontSize: '0.875rem',
                            opacity: testMutation.isPending ? 0.6 : 1,
                          }}
                        >
                          {testMutation.isPending ? 'Testing...' : 'Test'}
                        </button>
                        <button
                          onClick={() => handleEdit(system)}
                          style={{
                            padding: '0.5rem 1rem',
                            marginRight: '0.5rem',
                            backgroundColor: '#0066cc',
                            color: 'white',
                            border: 'none',
                            borderRadius: '4px',
                            cursor: 'pointer',
                            fontSize: '0.875rem',
                          }}
                        >
                          Edit
                        </button>
                        <button
                          onClick={() => setConfirmDialog({ isOpen: true, systemId: system.id, systemName: system.name })}
                          disabled={deleteMutation.isPending}
                          style={{
                            padding: '0.5rem 1rem',
                            backgroundColor: '#dc3545',
                            color: 'white',
                            border: 'none',
                            borderRadius: '4px',
                            cursor: 'pointer',
                            fontSize: '0.875rem',
                          }}
                        >
                          Delete
                        </button>
                      </div>
                      {testResults[system.id] && (
                        <div style={{
                          padding: '0.5rem',
                          borderRadius: '4px',
                          fontSize: '0.875rem',
                          backgroundColor: testResults[system.id].type === 'success' ? '#d4edda' : '#f8d7da',
                          color: testResults[system.id].type === 'success' ? '#155724' : '#721c24',
                          border: `1px solid ${testResults[system.id].type === 'success' ? '#c3e6cb' : '#f5c6cb'}`,
                        }}>
                          {testResults[system.id].text}
                        </div>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : searchQuery ? (
          <p style={{ textAlign: 'center', color: '#666', padding: '2rem' }}>
            No systems match your search. Try a different search term.
          </p>
        ) : (
          <p style={{ textAlign: 'center', color: '#666', padding: '2rem' }}>
            No storage systems configured. Click "Add System" to get started.
          </p>
        )}
      </div>
      </div>
    </>
  );
}

// 
