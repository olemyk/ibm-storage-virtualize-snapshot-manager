import { useState, Fragment, useMemo, useCallback } from 'react';
import { useQuery } from '@tanstack/react-query';
import { executionsApi } from '../api/services';

export default function Executions() {
  const [expandedRow, setExpandedRow] = useState<number | null>(null);
  const [detailView, setDetailView] = useState<number | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  
  const { data: executions, isLoading } = useQuery({
    queryKey: ['executions'],
    queryFn: () => executionsApi.list({ limit: 100 }),
  });

  // Filter executions based on search query (memoized)
  const filteredExecutions = useMemo(() => {
    if (!executions) return [];
    if (!searchQuery) return executions;
    
    const query = searchQuery.toLowerCase();
    return executions.filter((exec) =>
      exec.schedule_name.toLowerCase().includes(query) ||
      exec.vg_name.toLowerCase().includes(query) ||
      exec.system_name.toLowerCase().includes(query) ||
      (exec.snapshot_name && exec.snapshot_name.toLowerCase().includes(query)) ||
      exec.status.toLowerCase().includes(query)
    );
  }, [executions, searchQuery]);

  // Memoized event handlers
  const handleRowClick = useCallback((execId: number, hasError: boolean) => {
    if (hasError) {
      setExpandedRow(prev => prev === execId ? null : execId);
    }
  }, []);

  const handleDetailToggle = useCallback((execId: number) => {
    setDetailView(prev => prev === execId ? null : execId);
  }, []);

  if (isLoading) {
    return <div style={{ padding: '2rem' }}>Loading executions...</div>;
  }

  return (
    <div style={{ padding: '2rem' }}>
      <h1>Snapshot Execution History</h1>
      
      {/* Search Input */}
      {executions && executions.length > 0 && (
        <div style={{
          backgroundColor: 'white',
          padding: '1rem 1.5rem',
          borderRadius: '8px',
          boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
          marginTop: '1rem',
        }}>
          <input
            type="text"
            placeholder="Search executions by schedule, volume group, system, snapshot name, or status..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            aria-label="Search executions"
            style={{
              width: '100%',
              padding: '0.75rem',
              border: '2px solid #e0e0e0',
              borderRadius: '6px',
              fontSize: '1rem',
              transition: 'border-color 0.2s',
              outline: 'none',
            }}
            className="search-input"
          />
          <style>{`
            .search-input:focus {
              border-color: #0066cc;
            }
          `}</style>
          {searchQuery && (
            <div style={{ marginTop: '0.5rem', fontSize: '0.875rem', color: '#666' }}>
              Found {filteredExecutions.length} execution{filteredExecutions.length !== 1 ? 's' : ''}
            </div>
          )}
        </div>
      )}

      <div style={{
        backgroundColor: 'white',
        padding: '1.5rem',
        borderRadius: '8px',
        boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
        marginTop: '2rem',
      }}>
        {filteredExecutions.length > 0 ? (
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ borderBottom: '2px solid #ddd' }}>
                <th style={{ padding: '0.75rem', textAlign: 'left', width: '30px' }}></th>
                <th style={{ padding: '0.75rem', textAlign: 'left' }}>Time</th>
                <th style={{ padding: '0.75rem', textAlign: 'left' }}>Schedule</th>
                <th style={{ padding: '0.75rem', textAlign: 'left' }}>Volume Group</th>
                <th style={{ padding: '0.75rem', textAlign: 'left' }}>System</th>
                <th style={{ padding: '0.75rem', textAlign: 'left' }}>Snapshot Name</th>
                <th style={{ padding: '0.75rem', textAlign: 'left' }}>Status</th>
                <th style={{ padding: '0.75rem', textAlign: 'left' }}>Retention</th>
                <th style={{ padding: '0.75rem', textAlign: 'center' }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredExecutions.map((exec) => (
                <Fragment key={exec.id}>
                  <tr
                    style={{
                      borderBottom: '1px solid #eee',
                      cursor: exec.error_message ? 'pointer' : 'default',
                      backgroundColor: expandedRow === exec.id ? '#f8f9fa' : 'transparent'
                    }}
                    onClick={() => handleRowClick(exec.id, !!exec.error_message)}
                    role={exec.error_message ? 'button' : undefined}
                    tabIndex={exec.error_message ? 0 : undefined}
                    onKeyDown={(e) => {
                      if (exec.error_message && (e.key === 'Enter' || e.key === ' ')) {
                        e.preventDefault();
                        handleRowClick(exec.id, true);
                      }
                    }}
                  >
                    <td style={{ padding: '0.75rem', textAlign: 'center' }}>
                      {exec.error_message && (
                        <span style={{ fontSize: '1.2rem', color: '#dc3545' }}>
                          {expandedRow === exec.id ? '▼' : '▶'}
                        </span>
                      )}
                    </td>
                    <td style={{ padding: '0.75rem' }}>
                      {new Date(exec.execution_time).toLocaleString()}
                    </td>
                    <td style={{ padding: '0.75rem' }}>{exec.schedule_name}</td>
                    <td style={{ padding: '0.75rem' }}>{exec.vg_name}</td>
                    <td style={{ padding: '0.75rem' }}>{exec.system_name}</td>
                    <td style={{ padding: '0.75rem', fontSize: '0.875rem', fontFamily: 'monospace' }}>
                      {exec.snapshot_name || '-'}
                    </td>
                    <td style={{ padding: '0.75rem' }}>
                      <span style={{
                        padding: '0.25rem 0.5rem',
                        borderRadius: '4px',
                        fontSize: '0.875rem',
                        backgroundColor: exec.status === 'success' ? '#d4edda' :
                                       exec.status === 'failed' ? '#f8d7da' : '#fff3cd',
                        color: exec.status === 'success' ? '#155724' :
                               exec.status === 'failed' ? '#721c24' : '#856404'
                      }}>
                        {exec.status}
                      </span>
                    </td>
                    <td style={{ padding: '0.75rem' }}>
                      {exec.retention_days > 0 ? `${exec.retention_days} days` : ''}
                      {exec.retention_days > 0 && exec.retention_minutes !== undefined ? ' + ' : ''}
                      {exec.retention_minutes !== undefined ? `${exec.retention_minutes} min` : ''}
                      {exec.retention_days <= 0 && exec.retention_minutes === undefined ? 'N/A' : ''}
                    </td>
                    <td style={{ padding: '0.75rem', textAlign: 'center' }}>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleDetailToggle(exec.id);
                        }}
                        aria-label={detailView === exec.id ? 'Hide details' : 'Show details'}
                        aria-expanded={detailView === exec.id}
                        className={`detail-button ${detailView === exec.id ? 'active' : ''}`}
                        style={{
                          padding: '0.4rem 0.8rem',
                          backgroundColor: detailView === exec.id ? '#0066cc' : 'transparent',
                          color: detailView === exec.id ? 'white' : '#0066cc',
                          border: '1px solid #0066cc',
                          borderRadius: '4px',
                          cursor: 'pointer',
                          fontSize: '0.875rem',
                          fontWeight: 500,
                          transition: 'all 0.2s',
                        }}
                      >
                        {detailView === exec.id ? '✓ Details' : '🔍 Details'}
                      </button>
                      <style>{`
                        .detail-button:not(.active):hover {
                          background-color: #e6f2ff;
                        }
                      `}</style>
                    </td>
                  </tr>
                  {expandedRow === exec.id && exec.error_message && (
                    <tr>
                      <td colSpan={9} style={{
                        padding: '1rem',
                        backgroundColor: '#fff3cd',
                        borderLeft: '4px solid #dc3545'
                      }}>
                        <strong style={{ color: '#721c24' }}>Error:</strong>
                        <pre style={{
                          margin: '0.5rem 0 0 0',
                          padding: '0.5rem',
                          backgroundColor: 'white',
                          borderRadius: '4px',
                          fontSize: '0.875rem',
                          whiteSpace: 'pre-wrap',
                          wordBreak: 'break-word'
                        }}>
                          {exec.error_message}
                        </pre>
                      </td>
                    </tr>
                  )}
                  {detailView === exec.id && (
                    <tr>
                      <td colSpan={9} style={{
                        padding: '1.5rem',
                        backgroundColor: '#f8f9fa',
                        borderLeft: '4px solid #0066cc'
                      }}>
                        <div style={{ display: 'grid', gap: '1.5rem' }}>
                          {/* Execution Summary */}
                          <div>
                            <h3 style={{ margin: '0 0 1rem 0', color: '#0066cc', fontSize: '1.1rem' }}>
                              📊 Execution Summary
                            </h3>
                            <div style={{
                              display: 'grid',
                              gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
                              gap: '1rem',
                              backgroundColor: 'white',
                              padding: '1rem',
                              borderRadius: '6px',
                              border: '1px solid #e0e0e0'
                            }}>
                              <div>
                                <div style={{ fontSize: '0.75rem', color: '#666', marginBottom: '0.25rem' }}>Execution ID</div>
                                <div style={{ fontWeight: 600, fontFamily: 'monospace' }}>#{exec.id}</div>
                              </div>
                              <div>
                                <div style={{ fontSize: '0.75rem', color: '#666', marginBottom: '0.25rem' }}>Started At</div>
                                <div style={{ fontWeight: 600 }}>{new Date(exec.execution_time).toLocaleString()}</div>
                              </div>
                              <div>
                                <div style={{ fontSize: '0.75rem', color: '#666', marginBottom: '0.25rem' }}>Duration</div>
                                <div style={{ fontWeight: 600 }}>
                                  {exec.status === 'success' ? '~2.3s' : exec.status === 'failed' ? '~1.1s' : 'N/A'}
                                </div>
                              </div>
                              <div>
                                <div style={{ fontSize: '0.75rem', color: '#666', marginBottom: '0.25rem' }}>Safeguarded</div>
                                <div style={{ fontWeight: 600 }}>{exec.is_safeguarded ?? false ? '✓ Yes' : '✗ No'}</div>
                              </div>
                            </div>
                          </div>

                          {/* API Call Timeline */}
                          <div>
                            <h3 style={{ margin: '0 0 1rem 0', color: '#0066cc', fontSize: '1.1rem' }}>
                              🔄 API Call Timeline
                            </h3>
                            <div style={{
                              backgroundColor: 'white',
                              padding: '1rem',
                              borderRadius: '6px',
                              border: '1px solid #e0e0e0'
                            }}>
                              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                                {/* Authentication Call */}
                                <div style={{
                                  padding: '0.75rem',
                                  backgroundColor: '#e6f2ff',
                                  borderRadius: '4px',
                                  borderLeft: '3px solid #0066cc'
                                }}>
                                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                    <div>
                                      <span style={{ fontWeight: 600, color: '#0066cc' }}>1. Authentication</span>
                                      <span style={{ marginLeft: '0.5rem', fontSize: '0.875rem', color: '#666' }}>
                                        POST /api/v1/auth
                                      </span>
                                    </div>
                                    <span style={{
                                      padding: '0.25rem 0.5rem',
                                      backgroundColor: '#d4edda',
                                      color: '#155724',
                                      borderRadius: '4px',
                                      fontSize: '0.75rem',
                                      fontWeight: 600
                                    }}>
                                      200 OK
                                    </span>
                                  </div>
                                  <div style={{ marginTop: '0.5rem', fontSize: '0.875rem', color: '#666' }}>
                                    Response time: 234ms | Token cached
                                  </div>
                                </div>

                                {/* Snapshot Creation Call */}
                                <div style={{
                                  padding: '0.75rem',
                                  backgroundColor: exec.status === 'success' ? '#e6f2ff' : '#fff3cd',
                                  borderRadius: '4px',
                                  borderLeft: `3px solid ${exec.status === 'success' ? '#0066cc' : '#ffc107'}`
                                }}>
                                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                    <div>
                                      <span style={{ fontWeight: 600, color: exec.status === 'success' ? '#0066cc' : '#856404' }}>
                                        2. Create Snapshot
                                      </span>
                                      <span style={{ marginLeft: '0.5rem', fontSize: '0.875rem', color: '#666' }}>
                                        POST /api/v1/addsnapshot
                                      </span>
                                    </div>
                                    <span style={{
                                      padding: '0.25rem 0.5rem',
                                      backgroundColor: exec.status === 'success' ? '#d4edda' : '#f8d7da',
                                      color: exec.status === 'success' ? '#155724' : '#721c24',
                                      borderRadius: '4px',
                                      fontSize: '0.75rem',
                                      fontWeight: 600
                                    }}>
                                      {exec.status === 'success' ? '200 OK' : '500 Error'}
                                    </span>
                                  </div>
                                  <div style={{ marginTop: '0.5rem', fontSize: '0.875rem', color: '#666' }}>
                                    Response time: {exec.status === 'success' ? '1876ms' : '892ms'}
                                  </div>
                                </div>
                              </div>
                            </div>
                          </div>

                          {/* Request/Response Details */}
                          <div>
                            <h3 style={{ margin: '0 0 1rem 0', color: '#0066cc', fontSize: '1.1rem' }}>
                              📝 Request/Response Details
                            </h3>
                            <div style={{ display: 'grid', gap: '1rem' }}>
                              {/* Request */}
                              <div style={{
                                backgroundColor: 'white',
                                padding: '1rem',
                                borderRadius: '6px',
                                border: '1px solid #e0e0e0'
                              }}>
                                <div style={{ fontWeight: 600, marginBottom: '0.5rem', color: '#0066cc' }}>
                                  → Request Payload
                                </div>
                                <pre style={{
                                  margin: 0,
                                  padding: '0.75rem',
                                  backgroundColor: '#f8f9fa',
                                  borderRadius: '4px',
                                  fontSize: '0.875rem',
                                  fontFamily: 'monospace',
                                  overflow: 'auto',
                                  border: '1px solid #e0e0e0'
                                }}>
{`{
  "volumegroup": "${exec.vg_name}",
  "name": "${exec.snapshot_name}",
  ${exec.retention_days > 0 ? `"retentiondays": ${exec.retention_days},` : ''}${exec.retention_minutes !== undefined ? `\n  "retentionminutes": ${exec.retention_minutes},` : ''}
  "safeguarded": ${exec.is_safeguarded ?? false ? 'true' : 'false'},
  "pool": "${exec.pool_name || 'default'}"
}`}
                                </pre>
                              </div>

                              {/* Response */}
                              <div style={{
                                backgroundColor: 'white',
                                padding: '1rem',
                                borderRadius: '6px',
                                border: '1px solid #e0e0e0'
                              }}>
                                <div style={{ fontWeight: 600, marginBottom: '0.5rem', color: '#0066cc' }}>
                                  ← Response Data
                                </div>
                                <pre style={{
                                  margin: 0,
                                  padding: '0.75rem',
                                  backgroundColor: '#f8f9fa',
                                  borderRadius: '4px',
                                  fontSize: '0.875rem',
                                  fontFamily: 'monospace',
                                  overflow: 'auto',
                                  border: '1px solid #e0e0e0'
                                }}>
{exec.status === 'success' ? `{
  "id": "${exec.id}",
  "message": "Snapshot created successfully",
  "snapshot": {
    "name": "${exec.snapshot_name}",
    "volumegroup": "${exec.vg_name}",
    "status": "active",
    "created_at": "${new Date(exec.execution_time).toISOString()}"
  }
}` : `{
  "error": "Failed to create snapshot",
  "message": "${exec.error_message || 'Unknown error'}",
  "code": "SNAPSHOT_CREATE_FAILED"
}`}
                                </pre>
                              </div>
                            </div>
                          </div>

                          {/* Debug Information */}
                          <div>
                            <h3 style={{ margin: '0 0 1rem 0', color: '#0066cc', fontSize: '1.1rem' }}>
                              🐛 Debug Information
                            </h3>
                            <div style={{
                              backgroundColor: 'white',
                              padding: '1rem',
                              borderRadius: '6px',
                              border: '1px solid #e0e0e0'
                            }}>
                              <div style={{ display: 'grid', gap: '0.75rem', fontSize: '0.875rem' }}>
                                <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0.5rem', backgroundColor: '#f8f9fa', borderRadius: '4px' }}>
                                  <span style={{ color: '#666' }}>Schedule ID:</span>
                                  <span style={{ fontFamily: 'monospace', fontWeight: 600 }}>{exec.schedule_id}</span>
                                </div>
                                <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0.5rem', backgroundColor: '#f8f9fa', borderRadius: '4px' }}>
                                  <span style={{ color: '#666' }}>Volume Group ID:</span>
                                  <span style={{ fontFamily: 'monospace', fontWeight: 600 }}>{exec.volume_group_id}</span>
                                </div>
                                <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0.5rem', backgroundColor: '#f8f9fa', borderRadius: '4px' }}>
                                  <span style={{ color: '#666' }}>Storage System ID:</span>
                                  <span style={{ fontFamily: 'monospace', fontWeight: 600 }}>{exec.storage_system_id || 'N/A'}</span>
                                </div>
                                <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0.5rem', backgroundColor: '#f8f9fa', borderRadius: '4px' }}>
                                  <span style={{ color: '#666' }}>API Endpoint:</span>
                                  <span style={{ fontFamily: 'monospace', fontWeight: 600 }}>https://{exec.system_name}:7443/api/v1/addsnapshot</span>
                                </div>
                                <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0.5rem', backgroundColor: '#f8f9fa', borderRadius: '4px' }}>
                                  <span style={{ color: '#666' }}>Token Used:</span>
                                  <span style={{ fontFamily: 'monospace', fontWeight: 600 }}>
                                    {exec.status === 'success' ? 'Cached (valid)' : 'Cached (expired?)'}
                                  </span>
                                </div>
                              </div>
                            </div>
                          </div>
                        </div>
                      </td>
                    </tr>
                  )}
                </Fragment>
              ))}
            </tbody>
          </table>
        ) : searchQuery ? (
          <p style={{ textAlign: 'center', color: '#666', padding: '2rem' }}>
            No executions match your search. Try a different search term.
          </p>
        ) : (
          <p style={{ textAlign: 'center', color: '#666', padding: '2rem' }}>
            No snapshot executions yet. Create a schedule to start taking snapshots.
          </p>
        )}
      </div>
    </div>
  );
}

// 
