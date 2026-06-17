import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useState } from 'react';
import type { ReactNode } from 'react';

interface LayoutProps {
  children: ReactNode;
}

export default function Layout({ children }: LayoutProps) {
  const navigate = useNavigate();

  const handleLogout = () => {
    localStorage.removeItem('auth_token');
    navigate('/login');
  };

  return (
    <>
      <style>{`
        .logout-button:hover {
          background-color: var(--bg-hover) !important;
          border-color: var(--border-strong) !important;
        }
        .nav-link:not(.active):hover {
          color: var(--text-primary) !important;
          background-color: var(--bg-hover) !important;
        }
      `}</style>
      <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      {/* Header */}
      <header style={{
        backgroundColor: 'var(--bg-primary)',
        borderBottom: '1px solid var(--border-subtle)',
        padding: 'var(--spacing-lg) var(--spacing-xl)',
        boxShadow: 'var(--shadow-sm)'
      }}>
        <div style={{
          maxWidth: '1400px',
          margin: '0 auto',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center'
        }}>
          <div style={{ display: 'flex', alignItems: 'center' }}>
            <img
              src="/icon-snapshot-manager-bold.svg"
              alt="Snapshot Manager"
              style={{
                height: '130px',
                width: 'auto'
              }}
            />
          </div>
          <button
            onClick={handleLogout}
            className="logout-button"
            style={{
              padding: '0.625rem 1.25rem',
              backgroundColor: 'var(--bg-secondary)',
              color: 'var(--text-primary)',
              border: '1px solid var(--border-subtle)',
              borderRadius: 'var(--radius-sm)',
              cursor: 'pointer',
              fontSize: '0.875rem',
              fontWeight: 500,
              transition: 'all var(--transition-fast)'
            }}
          >
            Logout
          </button>
        </div>
      </header>

      {/* Navigation */}
      <nav style={{
        backgroundColor: 'var(--bg-primary)',
        borderBottom: '1px solid var(--border-subtle)',
        padding: '0 var(--spacing-xl)',
        boxShadow: 'var(--shadow-sm)'
      }}>
        <div style={{
          maxWidth: '1400px',
          margin: '0 auto',
          display: 'flex',
          gap: 'var(--spacing-sm)'
        }}>
          <NavLink to="/dashboard">Dashboard</NavLink>
          <NavLink to="/systems">Storage Systems</NavLink>
          <NavLink to="/volume-groups">Volume Groups</NavLink>
          <NavLink to="/schedules">Schedules</NavLink>
          <NavLink to="/executions">Executions</NavLink>
          <NavDropdown label="Notifications">
            <NavLink to="/notifications/channels">Channels</NavLink>
            <NavLink to="/notifications/rules">Alert Rules</NavLink>
            <NavLink to="/notifications/history">History</NavLink>
          </NavDropdown>
          <NavLink to="/audit-logs">Audit Logs</NavLink>
          <NavLink to="/settings">Settings</NavLink>
        </div>
      </nav>

      {/* Main Content */}
      <main style={{
        flex: 1,
        backgroundColor: 'var(--bg-secondary)',
        padding: 'var(--spacing-xl)'
      }}>
        <div style={{ maxWidth: '1400px', margin: '0 auto' }}>
          {children}
        </div>
      </main>

      {/* Footer */}
      <footer style={{
        backgroundColor: 'var(--bg-primary)',
        borderTop: '1px solid var(--border-subtle)',
        padding: 'var(--spacing-lg) var(--spacing-xl)',
        textAlign: 'center',
        color: 'var(--text-tertiary)',
        fontSize: '0.875rem'
      }}>
        <div style={{ maxWidth: '1400px', margin: '0 auto' }}>
          IBM Storage Virtualize Snapshot Manager © 2026
        </div>
      </footer>
      </div>
    </>
  );
}

interface NavLinkProps {
  to: string;
  children: ReactNode;
}

function NavLink({ to, children }: NavLinkProps) {
  const location = useLocation();
  const isActive = location.pathname === to ||
    (to === '/volume-groups' && location.pathname.startsWith('/volumegroups'));
  
  return (
    <Link
      to={to}
      className={isActive ? 'nav-link active' : 'nav-link'}
      style={{
        padding: 'var(--spacing-md) var(--spacing-lg)',
        color: isActive ? 'var(--primary)' : 'var(--text-secondary)',
        textDecoration: 'none',
        fontWeight: isActive ? 600 : 500,
        fontSize: '0.875rem',
        borderBottom: isActive ? '2px solid var(--primary)' : '2px solid transparent',
        transition: 'all var(--transition-fast)',
        display: 'inline-block'
      }}
    >
      {children}
    </Link>
  );
}

interface NavDropdownProps {
  label: string;
  children: ReactNode;
}

function NavDropdown({ label, children }: NavDropdownProps) {
  const [isOpen, setIsOpen] = useState(false);
  const location = useLocation();
  const isActive = location.pathname.startsWith('/notifications');

  return (
    <div
      style={{ position: 'relative', display: 'inline-block' }}
      onMouseEnter={() => setIsOpen(true)}
      onMouseLeave={() => setIsOpen(false)}
    >
      <div
        className={isActive ? 'nav-link active' : 'nav-link'}
        style={{
          padding: 'var(--spacing-md) var(--spacing-lg)',
          color: isActive ? 'var(--primary)' : 'var(--text-secondary)',
          fontWeight: isActive ? 600 : 500,
          fontSize: '0.875rem',
          borderBottom: isActive ? '2px solid var(--primary)' : '2px solid transparent',
          transition: 'all var(--transition-fast)',
          cursor: 'pointer',
          display: 'inline-block'
        }}
      >
        {label} ▾
      </div>
      {isOpen && (
        <div
          style={{
            position: 'absolute',
            top: '100%',
            left: 0,
            backgroundColor: 'var(--bg-primary)',
            border: '1px solid var(--border-subtle)',
            borderRadius: 'var(--radius-sm)',
            boxShadow: 'var(--shadow-md)',
            minWidth: '180px',
            zIndex: 1000,
            marginTop: '2px'
          }}
        >
          {children}
        </div>
      )}
    </div>
  );
}

// 
