import { useQuery } from '@tanstack/react-query';
import { dashboardApi, executionsApi } from '../api/services';
import { Link } from 'react-router-dom';

export default function Dashboard() {
  const { data: stats, isLoading: statsLoading } = useQuery({
    queryKey: ['dashboard-stats'],
    queryFn: dashboardApi.getStats,
  });

  const { data: executions, isLoading: executionsLoading } = useQuery({
    queryKey: ['recent-executions'],
    queryFn: () => executionsApi.list({ limit: 10 }),
  });

  if (statsLoading) {
    return <div style={{ padding: '2rem' }}>Loading dashboard...</div>;
  }

  return (
    <>
      <style>{`
        .stat-card:hover {
          transform: translateY(-2px) !important;
        }
      `}</style>
      <div style={{ padding: '2rem' }}>
      <h1 style={{ marginBottom: '2rem' }}>Dashboard</h1>

      {/* Stats Cards */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
        gap: '1rem',
        marginBottom: '2rem'
      }}>
        <StatCard
          title="Storage Systems"
          value={stats?.total_systems || 0}
          link="/systems"
        />
        <StatCard
          title="Volume Groups"
          value={stats?.total_volume_groups || 0}
          link="/volume-groups"
        />
        <StatCard
          title="Active Schedules"
          value={stats?.active_schedules || 0}
          link="/schedules"
        />
        <StatCard
          title="Recent Executions (24h)"
          value={stats?.recent_executions || 0}
          link="/executions"
        />
        <StatCard
          title="Successful (24h)"
          value={stats?.successful_executions || 0}
          color="#28a745"
        />
        <StatCard
          title="Failed (24h)"
          value={stats?.failed_executions || 0}
          color="#dc3545"
        />
      </div>

      {/* Recent Executions */}
      <div style={{
        backgroundColor: 'white',
        padding: '1.5rem',
        borderRadius: '8px',
        boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
      }}>
        <h2 style={{ marginBottom: '1rem' }}>Recent Executions</h2>
        
        {executionsLoading ? (
          <p>Loading executions...</p>
        ) : executions && executions.length > 0 ? (
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ borderBottom: '2px solid #ddd' }}>
                <th style={{ padding: '0.75rem', textAlign: 'left' }}>Time</th>
                <th style={{ padding: '0.75rem', textAlign: 'left' }}>Schedule</th>
                <th style={{ padding: '0.75rem', textAlign: 'left' }}>Volume Group</th>
                <th style={{ padding: '0.75rem', textAlign: 'left' }}>System</th>
                <th style={{ padding: '0.75rem', textAlign: 'left' }}>Status</th>
              </tr>
            </thead>
            <tbody>
              {executions.map((exec) => (
                <tr key={exec.id} style={{ borderBottom: '1px solid #eee' }}>
                  <td style={{ padding: '0.75rem' }}>
                    {new Date(exec.execution_time).toLocaleString()}
                  </td>
                  <td style={{ padding: '0.75rem' }}>{exec.schedule_name}</td>
                  <td style={{ padding: '0.75rem' }}>{exec.vg_name}</td>
                  <td style={{ padding: '0.75rem' }}>{exec.system_name}</td>
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
                </tr>
              ))}
            </tbody>
          </table>
        ) : (
          <p>No recent executions</p>
        )}
        
        <div style={{ marginTop: '1rem', textAlign: 'right' }}>
          <Link to="/executions" style={{ color: '#0066cc', textDecoration: 'none' }}>
            View all executions →
          </Link>
        </div>
      </div>
      </div>
    </>
  );
}

interface StatCardProps {
  title: string;
  value: number;
  link?: string;
  color?: string;
}

function StatCard({ title, value, link, color = '#0066cc' }: StatCardProps) {
  const content = (
    <div
      className={link ? 'stat-card' : ''}
      style={{
        backgroundColor: 'white',
        padding: '1.5rem',
        borderRadius: '8px',
        boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
        cursor: link ? 'pointer' : 'default',
        transition: 'transform 0.2s',
      }}
    >
      <div style={{ fontSize: '0.875rem', color: '#666', marginBottom: '0.5rem' }}>
        {title}
      </div>
      <div style={{ fontSize: '2rem', fontWeight: 'bold', color }}>
        {value}
      </div>
    </div>
  );

  return link ? <Link to={link} style={{ textDecoration: 'none' }}>{content}</Link> : content;
}

// 
