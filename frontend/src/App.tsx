import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Systems from './pages/Systems';
import VolumeGroups from './pages/VolumeGroups';
import Schedules from './pages/Schedules';
import VolumeGroupSchedules from './pages/VolumeGroupSchedules';
import Executions from './pages/Executions';
import AuditLogs from './pages/AuditLogs';
import Settings from './pages/Settings';
import NotificationChannels from './pages/NotificationChannels';
import AlertRules from './pages/AlertRules';
import NotificationHistory from './pages/NotificationHistory';
import Layout from './components/Layout';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
});

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const token = sessionStorage.getItem('auth_token');
  return token ? <>{children}</> : <Navigate to="/login" />;
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route
            path="/*"
            element={
              <PrivateRoute>
                <Layout>
                  <Routes>
                    <Route path="/dashboard" element={<Dashboard />} />
                    <Route path="/systems" element={<Systems />} />
                    <Route path="/volume-groups" element={<VolumeGroups />} />
                    <Route path="/volumegroups/:id/schedules" element={<VolumeGroupSchedules />} />
                    <Route path="/schedules" element={<Schedules />} />
                    <Route path="/executions" element={<Executions />} />
                    <Route path="/notifications/channels" element={<NotificationChannels />} />
                    <Route path="/notifications/rules" element={<AlertRules />} />
                    <Route path="/notifications/history" element={<NotificationHistory />} />
                    <Route path="/audit-logs" element={<AuditLogs />} />
                    <Route path="/settings" element={<Settings />} />
                    <Route path="/" element={<Navigate to="/dashboard" />} />
                    <Route path="*" element={<div style={{ padding: '2rem' }}>Page not found</div>} />
                  </Routes>
                </Layout>
              </PrivateRoute>
            }
          />
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}

export default App;

// 
