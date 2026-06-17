import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { systemsApi, volumeGroupsApi } from '../api/services';
import type { StorageSystem, VolumeGroup } from '../types';

// Constants
const MESSAGE_DISPLAY_DURATION_MS = 5000; // 5 seconds
const ERROR_MESSAGE_DISPLAY_DURATION_MS = 10000; // 10 seconds

export default function VolumeGroups() {
  const navigate = useNavigate();
  const [selectedSystemId, setSelectedSystemId] = useState<number | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedPartition, setSelectedPartition] = useState<string>('');
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const [syncingSystemId, setSyncingSystemId] = useState<number | null>(null);

  // Fetch storage systems
  const { data: systems = [], isLoading: systemsLoading } = useQuery({
    queryKey: ['systems'],
    queryFn: systemsApi.list,
  });

  // Fetch volume groups for selected system
  const {
    data: volumeGroups = [],
    isLoading: vgLoading,
    refetch: refetchVolumeGroups
  } = useQuery({
    queryKey: ['volumeGroups', selectedSystemId],
    queryFn: async () => {
      if (!selectedSystemId) return [];
      try {
        return await volumeGroupsApi.listBySystem(selectedSystemId);
      } catch (error: any) {
        // Show error message
        setMessage({
          type: 'error',
          text: error.response?.data?.error || 'Failed to load volume groups. Please check system connection.'
        });
        // Return empty array to prevent null errors
        return [];
      }
    },
    enabled: selectedSystemId !== null,
    retry: false, // Don't retry on error
  });

  const handleSystemChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const systemId = e.target.value ? parseInt(e.target.value, 10) : null;
    setSelectedSystemId(systemId);
    setMessage(null);
  };

  const handleSync = async (systemId: number) => {
    setSyncingSystemId(systemId);
    setMessage(null);
    
    try {
      const result = await volumeGroupsApi.sync(systemId);
      setMessage({ 
        type: 'success', 
        text: `Successfully synced ${result.count} volume groups from storage system` 
      });
      setTimeout(() => setMessage(null), MESSAGE_DISPLAY_DURATION_MS);
      
      // Refetch volume groups after sync
      if (selectedSystemId === systemId) {
        refetchVolumeGroups();
      }
    } catch (error: any) {
      setMessage({ 
        type: 'error', 
        text: error.response?.data?.error || 'Failed to sync volume groups' 
      });
      setTimeout(() => setMessage(null), ERROR_MESSAGE_DISPLAY_DURATION_MS);
    } finally {
      setSyncingSystemId(null);
    }
  };

  const handleViewSchedules = (vgId: number) => {
    navigate(`/volumegroups/${vgId}/schedules`);
  };

  // Get unique partitions for filter (memoized)
  const uniquePartitions = useMemo(() => {
    const groups = Array.isArray(volumeGroups) ? volumeGroups : [];
    const partitions = new Set<string>();
    groups.forEach((vg) => {
      if (vg.partition_name) {
        partitions.add(vg.partition_name);
      }
    });
    return Array.from(partitions).sort();
  }, [volumeGroups]);

  // Filter volume groups based on search query and partition (memoized)
  const filteredVolumeGroups = useMemo(() => {
    // Ensure volumeGroups is always an array
    const groups = Array.isArray(volumeGroups) ? volumeGroups : [];
    return groups.filter((vg) => {
      // Search filter
      if (searchQuery) {
        const query = searchQuery.toLowerCase();
        const matchesSearch =
          vg.vg_name.toLowerCase().includes(query) ||
          vg.vg_id.toLowerCase().includes(query) ||
          (vg.partition_name && vg.partition_name.toLowerCase().includes(query));
        if (!matchesSearch) return false;
      }
      
      // Partition filter
      if (selectedPartition && vg.partition_name !== selectedPartition) {
        return false;
      }
      
      return true;
    });
  }, [volumeGroups, searchQuery, selectedPartition]);

  if (systemsLoading) {
    return (
      <div style={{ padding: '2rem' }}>
        <h1>Volume Groups</h1>
        <p>Loading storage systems...</p>
      </div>
    );
  }

  if (systems.length === 0) {
    return (
      <div style={{ padding: '2rem' }}>
        <h1>Volume Groups</h1>
        <div style={{
          backgroundColor: '#fff3cd',
          border: '1px solid #ffc107',
          padding: '1rem',
          borderRadius: '4px',
          marginTop: '1rem',
        }}>
          <p style={{ margin: 0 }}>
            No storage systems configured. Please add a storage system first.
          </p>
          <button
            onClick={() => navigate('/systems')}
            style={{
              marginTop: '1rem',
              padding: '0.5rem 1rem',
              backgroundColor: '#007bff',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            Go to Storage Systems
          </button>
        </div>
      </div>
    );
  }

  return (
    <>
      <style>{`
        .search-input:focus {
          border-color: #0066cc !important;
        }
      `}</style>
      <div style={{ padding: '2rem' }}>
      <h1>Volume Groups</h1>

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

      {/* System Selection */}
      <div style={{
        backgroundColor: 'white',
        padding: '1.5rem',
        borderRadius: '8px',
        boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
        marginBottom: '1.5rem',
      }}>
        <div style={{ display: 'flex', gap: '1rem', alignItems: 'center', flexWrap: 'wrap' }}>
          <div style={{ flex: '1', minWidth: '250px' }}>
            <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
              Select Storage System:
            </label>
            <select
              value={selectedSystemId || ''}
              onChange={handleSystemChange}
              style={{
                width: '100%',
                padding: '0.5rem',
                border: '1px solid #ddd',
                borderRadius: '4px',
                fontSize: '1rem',
              }}
            >
              <option value="">-- Select a system --</option>
              {systems.map((system: StorageSystem) => (
                <option key={system.id} value={system.id}>
                  {system.name} ({system.ip_address})
                </option>
              ))}
            </select>
          </div>

          {selectedSystemId && (
            <button
              onClick={() => handleSync(selectedSystemId)}
              disabled={syncingSystemId === selectedSystemId}
              style={{
                padding: '0.5rem 1.5rem',
                backgroundColor: syncingSystemId === selectedSystemId ? '#6c757d' : '#28a745',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: syncingSystemId === selectedSystemId ? 'not-allowed' : 'pointer',
                fontSize: '1rem',
                marginTop: '1.5rem',
              }}
            >
              {syncingSystemId === selectedSystemId ? 'Syncing...' : '🔄 Sync Volume Groups'}
            </button>
          )}
        </div>
      </div>

      {/* Volume Groups Table */}
      {selectedSystemId && (
        <div style={{
          backgroundColor: 'white',
          padding: '1.5rem',
          borderRadius: '8px',
          boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
        }}>
          <h2 style={{ marginTop: 0, marginBottom: '1rem' }}>Volume Groups</h2>

          {/* Search and Filter */}
          {!vgLoading && volumeGroups && volumeGroups.length > 0 && (
            <div style={{ marginBottom: '1rem' }}>
              <div style={{ display: 'flex', gap: '1rem', marginBottom: '0.5rem' }}>
                <input
                  type="text"
                  placeholder="Search volume groups by name, ID, or partition..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="search-input"
                  style={{
                    flex: 1,
                    padding: '0.75rem',
                    border: '2px solid #e0e0e0',
                    borderRadius: '6px',
                    fontSize: '1rem',
                    transition: 'border-color 0.2s',
                    outline: 'none',
                  }}
                />
                {uniquePartitions.length > 0 && (
                  <select
                    value={selectedPartition}
                    onChange={(e) => setSelectedPartition(e.target.value)}
                    style={{
                      padding: '0.75rem',
                      border: '2px solid #e0e0e0',
                      borderRadius: '6px',
                      fontSize: '1rem',
                      backgroundColor: 'white',
                      minWidth: '200px',
                    }}
                  >
                    <option value="">All Partitions</option>
                    {uniquePartitions.map((partition) => (
                      <option key={partition} value={partition}>
                        {partition}
                      </option>
                    ))}
                  </select>
                )}
              </div>
              {(searchQuery || selectedPartition) && (
                <div style={{ fontSize: '0.875rem', color: '#666' }}>
                  Found {filteredVolumeGroups.length} volume group{filteredVolumeGroups.length !== 1 ? 's' : ''}
                  {selectedPartition && ` in partition "${selectedPartition}"`}
                </div>
              )}
            </div>
          )}

          {vgLoading ? (
            <p>Loading volume groups...</p>
          ) : !volumeGroups || volumeGroups.length === 0 ? (
            <div style={{
              padding: '2rem',
              textAlign: 'center',
              color: '#666',
              backgroundColor: '#f8f9fa',
              borderRadius: '4px',
            }}>
              <p style={{ margin: 0 }}>No volume groups found.</p>
              <p style={{ marginTop: '0.5rem', fontSize: '0.9rem' }}>
                Click "Sync Volume Groups" to fetch volume groups from the storage system.
              </p>
            </div>
          ) : filteredVolumeGroups.length === 0 ? (
            <div style={{
              padding: '2rem',
              textAlign: 'center',
              color: '#666',
              backgroundColor: '#f8f9fa',
              borderRadius: '4px',
            }}>
              <p style={{ margin: 0 }}>No volume groups match your search.</p>
              <p style={{ marginTop: '0.5rem', fontSize: '0.9rem' }}>
                Try a different search term.
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
                    <th style={{ padding: '0.75rem', textAlign: 'left' }}>VG ID</th>
                    <th style={{ padding: '0.75rem', textAlign: 'left' }}>Partition</th>
                    <th style={{ padding: '0.75rem', textAlign: 'center' }}>Schedules</th>
                    <th style={{ padding: '0.75rem', textAlign: 'left' }}>Last Synced</th>
                    <th style={{ padding: '0.75rem', textAlign: 'center' }}>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredVolumeGroups.map((vg: VolumeGroup) => (
                    <tr key={vg.id} style={{ borderBottom: '1px solid #dee2e6' }}>
                      <td style={{ padding: '0.75rem' }}>
                        <strong>{vg.vg_name}</strong>
                      </td>
                      <td style={{ padding: '0.75rem', color: '#666' }}>
                        {vg.vg_id}
                      </td>
                      <td style={{ padding: '0.75rem', color: '#666' }}>
                        {vg.partition_name || 'N/A'}
                      </td>
                      <td style={{ padding: '0.75rem', textAlign: 'center' }}>
                        <span style={{
                          padding: '0.25rem 0.75rem',
                          backgroundColor: vg.schedule_count > 0 ? '#d4edda' : '#f8f9fa',
                          color: vg.schedule_count > 0 ? '#155724' : '#666',
                          borderRadius: '12px',
                          fontSize: '0.85rem',
                          fontWeight: 'bold',
                        }}>
                          {vg.schedule_count || 0}
                        </span>
                      </td>
                      <td style={{ padding: '0.75rem', color: '#666', fontSize: '0.9rem' }}>
                        {vg.last_synced_at ? new Date(vg.last_synced_at).toLocaleString() : 'Never'}
                      </td>
                      <td style={{ padding: '0.75rem', textAlign: 'center' }}>
                        <button
                          onClick={() => handleViewSchedules(vg.id)}
                          style={{
                            padding: '0.4rem 1rem',
                            backgroundColor: '#007bff',
                            color: 'white',
                            border: 'none',
                            borderRadius: '4px',
                            cursor: 'pointer',
                            fontSize: '0.9rem',
                          }}
                        >
                          Manage Schedules
                        </button>
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
    </>
  );
}

// 
