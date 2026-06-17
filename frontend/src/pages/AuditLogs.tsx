import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { auditLogsApi } from '../api/services';
import type { AuditLog, AuditLogFilters } from '../types';

export default function AuditLogs() {
  const [filters, setFilters] = useState<AuditLogFilters>({
    limit: 50,
    offset: 0,
  });
  const [search, setSearch] = useState('');

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['auditLogs', filters],
    queryFn: () => auditLogsApi.list(filters),
  });

  const logs: AuditLog[] = data?.logs || [];
  const total: number = data?.total || 0;

  // Filter logs based on search (memoized)
  const filteredLogs = useMemo(() =>
    logs.filter(log =>
      !search ||
      log.username.toLowerCase().includes(search.toLowerCase()) ||
      log.resource_name?.toLowerCase().includes(search.toLowerCase())
    ),
    [logs, search]
  );

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  };

  const getStatusBadge = (status: string) => {
    const isSuccess = status === 'success';
    return (
      <span style={{
        padding: '0.25rem 0.75rem',
        backgroundColor: isSuccess ? '#d4edda' : '#f8d7da',
        color: isSuccess ? '#155724' : '#721c24',
        borderRadius: '12px',
        fontSize: '0.85rem',
        fontWeight: 'bold',
      }}>
        {status}
      </span>
    );
  };

  const getActionBadge = (action: string) => {
    const colors: Record<string, string> = {
      login: '#007bff',
      logout: '#6c757d',
      create_schedule: '#28a745',
      execute_schedule: '#17a2b8',
      delete_schedule: '#dc3545',
      create_system: '#28a745',
      update_system: '#ffc107',
      delete_system: '#dc3545',
    };
    return (
      <span style={{
        padding: '0.25rem 0.75rem',
        backgroundColor: colors[action] || '#6c757d',
        color: 'white',
        borderRadius: '12px',
        fontSize: '0.85rem',
      }}>
        {action.replace(/_/g, ' ')}
      </span>
    );
  };

  return (
    <div style={{ padding: '2rem' }}>
      <h1>Audit Logs</h1>
      
      {/* Filters */}
      <div style={{
        backgroundColor: 'white',
        padding: '1.5rem',
        borderRadius: '8px',
        marginBottom: '1.5rem',
        boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
      }}>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '1rem' }}>
          <div>
            <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
              Action
            </label>
            <select
              value={filters.action || ''}
              onChange={(e) => setFilters({ ...filters, action: e.target.value || undefined, offset: 0 })}
              style={{
                width: '100%',
                padding: '0.5rem',
                border: '1px solid #ddd',
                borderRadius: '4px',
              }}
            >
              <option value="">All Actions</option>
              <option value="login">Login</option>
              <option value="logout">Logout</option>
              <option value="create_schedule">Create Schedule</option>
              <option value="execute_schedule">Execute Schedule</option>
              <option value="delete_schedule">Delete Schedule</option>
              <option value="create_system">Create System</option>
              <option value="update_system">Update System</option>
              <option value="delete_system">Delete System</option>
            </select>
          </div>

          <div>
            <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
              Status
            </label>
            <select
              value={filters.status || ''}
              onChange={(e) => setFilters({ ...filters, status: e.target.value || undefined, offset: 0 })}
              style={{
                width: '100%',
                padding: '0.5rem',
                border: '1px solid #ddd',
                borderRadius: '4px',
              }}
            >
              <option value="">All Statuses</option>
              <option value="success">Success</option>
              <option value="failed">Failed</option>
            </select>
          </div>

          <div>
            <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
              Search
            </label>
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search username, resource..."
              style={{
                width: '100%',
                padding: '0.5rem',
                border: '1px solid #ddd',
                borderRadius: '4px',
              }}
            />
          </div>

          <div style={{ display: 'flex', alignItems: 'flex-end' }}>
            <button
              onClick={() => refetch()}
              style={{
                padding: '0.5rem 1rem',
                backgroundColor: '#17a2b8',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: 'pointer',
              }}
            >
              ↻ Refresh
            </button>
          </div>
        </div>
      </div>

      {/* Table */}
      <div style={{
        backgroundColor: 'white',
        padding: '1.5rem',
        borderRadius: '8px',
        boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
      }}>
        <h2>Logs ({total})</h2>
        
        {isLoading ? (
          <p>Loading...</p>
        ) : logs.length === 0 ? (
          <p style={{ textAlign: 'center', color: '#666', padding: '2rem' }}>
            No audit logs found
          </p>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ backgroundColor: '#f8f9fa', borderBottom: '2px solid #dee2e6' }}>
                  <th style={{ padding: '0.75rem', textAlign: 'left' }}>Timestamp</th>
                  <th style={{ padding: '0.75rem', textAlign: 'left' }}>User</th>
                  <th style={{ padding: '0.75rem', textAlign: 'left' }}>Action</th>
                  <th style={{ padding: '0.75rem', textAlign: 'left' }}>Resource</th>
                  <th style={{ padding: '0.75rem', textAlign: 'center' }}>Status</th>
                  <th style={{ padding: '0.75rem', textAlign: 'left' }}>Details</th>
                </tr>
              </thead>
              <tbody>
                {filteredLogs.map((log) => (
                    <tr key={log.id} style={{ borderBottom: '1px solid #dee2e6' }}>
                      <td style={{ padding: '0.75rem', fontSize: '0.9rem' }}>
                        {formatTimestamp(log.created_at)}
                      </td>
                      <td style={{ padding: '0.75rem' }}>
                        <strong>{log.username}</strong>
                        {log.ip_address && (
                          <div style={{ fontSize: '0.85rem', color: '#666' }}>
                            {log.ip_address}
                          </div>
                        )}
                      </td>
                      <td style={{ padding: '0.75rem' }}>
                        {getActionBadge(log.action)}
                      </td>
                      <td style={{ padding: '0.75rem' }}>
                        <div>{log.resource_type}</div>
                        {log.resource_name && (
                          <div style={{ fontSize: '0.85rem', color: '#666' }}>
                            {log.resource_name}
                          </div>
                        )}
                      </td>
                      <td style={{ padding: '0.75rem', textAlign: 'center' }}>
                        {getStatusBadge(log.status)}
                      </td>
                      <td style={{
                        padding: '0.75rem',
                        fontSize: '0.85rem',
                        maxWidth: '400px',
                        wordBreak: 'break-word',
                        whiteSpace: 'pre-wrap'
                      }}>
                        {log.error_message ? (
                          <span style={{ color: '#dc3545' }}>{log.error_message}</span>
                        ) : log.details ? (
                          <pre style={{
                            margin: 0,
                            fontSize: '0.8rem',
                            fontFamily: 'monospace',
                            whiteSpace: 'pre-wrap',
                            wordBreak: 'break-word'
                          }}>
                            {JSON.stringify(JSON.parse(log.details), null, 2)}
                          </pre>
                        ) : '—'}
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Pagination */}
        {total > (filters.limit || 50) && (
          <div style={{ marginTop: '1rem', display: 'flex', justifyContent: 'center', gap: '0.5rem' }}>
            <button
              onClick={() => setFilters({ ...filters, offset: Math.max(0, (filters.offset || 0) - (filters.limit || 50)) })}
              disabled={(filters.offset || 0) === 0}
              style={{
                padding: '0.5rem 1rem',
                backgroundColor: '#007bff',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: (filters.offset || 0) === 0 ? 'not-allowed' : 'pointer',
                opacity: (filters.offset || 0) === 0 ? 0.5 : 1,
              }}
            >
              Previous
            </button>
            <span style={{ padding: '0.5rem 1rem', alignSelf: 'center' }}>
              Page {Math.floor((filters.offset || 0) / (filters.limit || 50)) + 1}
            </span>
            <button
              onClick={() => setFilters({ ...filters, offset: (filters.offset || 0) + (filters.limit || 50) })}
              disabled={(filters.offset || 0) + (filters.limit || 50) >= total}
              style={{
                padding: '0.5rem 1rem',
                backgroundColor: '#007bff',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: (filters.offset || 0) + (filters.limit || 50) >= total ? 'not-allowed' : 'pointer',
                opacity: (filters.offset || 0) + (filters.limit || 50) >= total ? 0.5 : 1,
              }}
            >
              Next
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

// 