import { useQuery } from '@tanstack/react-query';
import { notificationsApi } from '../api/services';
import { useState } from 'react';
import type { NotificationHistory, NotificationHistoryFilters } from '../types';

export default function NotificationHistory() {
  const [filters, setFilters] = useState<NotificationHistoryFilters>({
    limit: 50,
    offset: 0,
  });
  
  const [expandedId, setExpandedId] = useState<number | null>(null);

  const { data: history, isLoading } = useQuery({
    queryKey: ['notification-history', filters],
    queryFn: () => notificationsApi.listHistory(filters),
  });

  const handleFilterChange = (key: keyof NotificationHistoryFilters, value: string | number | undefined) => {
    setFilters(prev => ({
      ...prev,
      [key]: value,
      offset: 0, // Reset to first page when filters change
    }));
  };

  const clearFilters = () => {
    setFilters({
      limit: 50,
      offset: 0,
    });
  };

  const nextPage = () => {
    setFilters(prev => ({
      ...prev,
      offset: (prev.offset || 0) + (prev.limit || 50),
    }));
  };

  const prevPage = () => {
    setFilters(prev => ({
      ...prev,
      offset: Math.max(0, (prev.offset || 0) - (prev.limit || 50)),
    }));
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'sent': return '#28a745';
      case 'failed': return '#dc3545';
      case 'pending': return '#ffc107';
      default: return '#6c757d';
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

  const exportToCSV = () => {
    if (!history || history.length === 0) return;

    const headers = ['ID', 'Timestamp', 'Event Type', 'Severity', 'Channel', 'Status', 'Message', 'Error'];
    const rows = history.map(h => [
      h.id,
      new Date(h.created_at).toISOString(),
      h.event_type,
      h.severity,
      h.channel_name || 'N/A',
      h.status,
      h.message.replace(/"/g, '""'),
      h.error_message?.replace(/"/g, '""') || '',
    ]);

    const csv = [
      headers.join(','),
      ...rows.map(row => row.map(cell => `"${cell}"`).join(','))
    ].join('\n');

    const blob = new Blob([csv], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `notification-history-${new Date().toISOString().split('T')[0]}.csv`;
    a.click();
    window.URL.revokeObjectURL(url);
  };

  if (isLoading) {
    return <div style={{ padding: '2rem' }}>Loading notification history...</div>;
  }

  return (
    <div style={{ padding: '2rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
        <h1>Notification History</h1>
        <button
          onClick={exportToCSV}
          disabled={!history || history.length === 0}
          style={{
            padding: '0.5rem 1rem',
            backgroundColor: history && history.length > 0 ? '#28a745' : '#ccc',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: history && history.length > 0 ? 'pointer' : 'not-allowed',
          }}
        >
          Export to CSV
        </button>
      </div>

      {/* Filters */}
      <div style={{
        marginBottom: '2rem',
        padding: '1.5rem',
        border: '1px solid #ddd',
        borderRadius: '8px',
        backgroundColor: '#f8f9fa',
      }}>
        <h3 style={{ marginTop: 0 }}>Filters</h3>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '1rem' }}>
          <div>
            <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>Status</label>
            <select
              value={filters.status || ''}
              onChange={(e) => handleFilterChange('status', e.target.value || undefined)}
              style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
            >
              <option value="">All</option>
              <option value="sent">Sent</option>
              <option value="failed">Failed</option>
              <option value="pending">Pending</option>
            </select>
          </div>

          <div>
            <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>Event Type</label>
            <select
              value={filters.event_type || ''}
              onChange={(e) => handleFilterChange('event_type', e.target.value || undefined)}
              style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
            >
              <option value="">All</option>
              <option value="snapshot_success">Snapshot Success</option>
              <option value="snapshot_failure">Snapshot Failure</option>
              <option value="snapshot_warning">Snapshot Warning</option>
              <option value="system_connection_lost">System Connection Lost</option>
              <option value="scheduler_error">Scheduler Error</option>
              <option value="consecutive_failures">Consecutive Failures</option>
            </select>
          </div>

          <div>
            <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>Severity</label>
            <select
              value={filters.severity || ''}
              onChange={(e) => handleFilterChange('severity', e.target.value || undefined)}
              style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
            >
              <option value="">All</option>
              <option value="info">Info</option>
              <option value="warning">Warning</option>
              <option value="error">Error</option>
              <option value="critical">Critical</option>
            </select>
          </div>

          <div>
            <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>Channel ID</label>
            <input
              type="number"
              value={filters.channel_id || ''}
              onChange={(e) => handleFilterChange('channel_id', e.target.value ? parseInt(e.target.value) : undefined)}
              style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
              placeholder="Filter by channel..."
            />
          </div>

          <div>
            <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>From Date</label>
            <input
              type="datetime-local"
              value={filters.from_date || ''}
              onChange={(e) => handleFilterChange('from_date', e.target.value || undefined)}
              style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
            />
          </div>

          <div>
            <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 500 }}>To Date</label>
            <input
              type="datetime-local"
              value={filters.to_date || ''}
              onChange={(e) => handleFilterChange('to_date', e.target.value || undefined)}
              style={{ width: '100%', padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
            />
          </div>
        </div>

        <div style={{ marginTop: '1rem', display: 'flex', gap: '1rem' }}>
          <button
            onClick={clearFilters}
            style={{
              padding: '0.5rem 1rem',
              backgroundColor: '#6c757d',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            Clear Filters
          </button>
          <div style={{ marginLeft: 'auto', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <label style={{ fontWeight: 500 }}>Per Page:</label>
            <select
              value={filters.limit || 50}
              onChange={(e) => handleFilterChange('limit', parseInt(e.target.value))}
              style={{ padding: '0.5rem', border: '1px solid #ddd', borderRadius: '4px' }}
            >
              <option value="25">25</option>
              <option value="50">50</option>
              <option value="100">100</option>
              <option value="200">200</option>
            </select>
          </div>
        </div>
      </div>

      {/* History List */}
      <div style={{ display: 'grid', gap: '1rem' }}>
        {history?.map((item) => (
          <div
            key={item.id}
            style={{
              padding: '1rem',
              border: '1px solid #ddd',
              borderRadius: '8px',
              backgroundColor: 'white',
            }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start' }}>
              <div style={{ flex: 1 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.5rem' }}>
                  <span style={{
                    padding: '0.25rem 0.5rem',
                    borderRadius: '4px',
                    fontSize: '0.75rem',
                    fontWeight: 'bold',
                    color: 'white',
                    backgroundColor: getStatusColor(item.status),
                    textTransform: 'uppercase',
                  }}>
                    {item.status}
                  </span>
                  <span style={{
                    padding: '0.25rem 0.5rem',
                    borderRadius: '4px',
                    fontSize: '0.75rem',
                    fontWeight: 'bold',
                    color: 'white',
                    backgroundColor: getSeverityColor(item.severity),
                    textTransform: 'uppercase',
                  }}>
                    {item.severity}
                  </span>
                  <span style={{ fontSize: '0.85rem', color: '#666' }}>
                    {new Date(item.created_at).toLocaleString()}
                  </span>
                </div>

                <div style={{ marginBottom: '0.5rem' }}>
                  <strong>{item.event_type.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}</strong>
                </div>

                <p style={{ margin: '0.5rem 0', color: '#333' }}>{item.message}</p>

                {item.error_message && (
                  <div style={{
                    marginTop: '0.5rem',
                    padding: '0.5rem',
                    backgroundColor: '#f8d7da',
                    color: '#721c24',
                    borderRadius: '4px',
                    fontSize: '0.9rem',
                  }}>
                    <strong>Error:</strong> {item.error_message}
                  </div>
                )}

                <div style={{ 
                  display: 'grid', 
                  gridTemplateColumns: 'auto 1fr', 
                  gap: '0.5rem 1rem', 
                  fontSize: '0.85rem', 
                  marginTop: '0.5rem',
                  color: '#666'
                }}>
                  <strong>Channel:</strong>
                  <span>{item.channel_name || 'N/A'}</span>
                  
                  {item.rule_name && (
                    <>
                      <strong>Rule:</strong>
                      <span>{item.rule_name}</span>
                    </>
                  )}
                  
                  {item.sent_at && (
                    <>
                      <strong>Sent At:</strong>
                      <span>{new Date(item.sent_at).toLocaleString()}</span>
                    </>
                  )}
                </div>

                {item.event_details && (
                  <div style={{ marginTop: '0.5rem' }}>
                    <button
                      onClick={() => setExpandedId(expandedId === item.id ? null : item.id)}
                      style={{
                        padding: '0.25rem 0.5rem',
                        backgroundColor: '#007bff',
                        color: 'white',
                        border: 'none',
                        borderRadius: '4px',
                        cursor: 'pointer',
                        fontSize: '0.85rem',
                      }}
                    >
                      {expandedId === item.id ? 'Hide' : 'Show'} Details
                    </button>
                    
                    {expandedId === item.id && (
                      <pre style={{
                        marginTop: '0.5rem',
                        padding: '1rem',
                        backgroundColor: '#f8f9fa',
                        borderRadius: '4px',
                        overflow: 'auto',
                        fontSize: '0.85rem',
                      }}>
                        {JSON.stringify(JSON.parse(item.event_details), null, 2)}
                      </pre>
                    )}
                  </div>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>

      {history?.length === 0 && (
        <div style={{ textAlign: 'center', padding: '3rem', color: '#666' }}>
          <p>No notification history found.</p>
          <p>Notifications will appear here once they are sent.</p>
        </div>
      )}

      {/* Pagination */}
      {history && history.length > 0 && (
        <div style={{
          marginTop: '2rem',
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          gap: '1rem',
        }}>
          <button
            onClick={prevPage}
            disabled={(filters.offset || 0) === 0}
            style={{
              padding: '0.5rem 1rem',
              backgroundColor: (filters.offset || 0) === 0 ? '#ccc' : '#007bff',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: (filters.offset || 0) === 0 ? 'not-allowed' : 'pointer',
            }}
          >
            Previous
          </button>
          
          <span style={{ color: '#666' }}>
            Showing {(filters.offset || 0) + 1} - {(filters.offset || 0) + history.length}
          </span>
          
          <button
            onClick={nextPage}
            disabled={history.length < (filters.limit || 50)}
            style={{
              padding: '0.5rem 1rem',
              backgroundColor: history.length < (filters.limit || 50) ? '#ccc' : '#007bff',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: history.length < (filters.limit || 50) ? 'not-allowed' : 'pointer',
            }}
          >
            Next
          </button>
        </div>
      )}
    </div>
  );
}

// 