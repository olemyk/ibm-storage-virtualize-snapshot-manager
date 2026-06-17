import { useState } from 'react';
import type { FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { authApi } from '../api/services';

// Validation constants
const MIN_USERNAME_LENGTH = 3;
const MAX_USERNAME_LENGTH = 50;
const MIN_PASSWORD_LENGTH = 8;
const MAX_PASSWORD_LENGTH = 128;

export default function Login() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    
    // Frontend validation
    if (!username || username.trim().length === 0) {
      setError('Username is required');
      return;
    }
    
    if (username.length < MIN_USERNAME_LENGTH) {
      setError(`Username must be at least ${MIN_USERNAME_LENGTH} characters long`);
      return;
    }
    
    if (username.length > MAX_USERNAME_LENGTH) {
      setError(`Username must not exceed ${MAX_USERNAME_LENGTH} characters`);
      return;
    }
    
    // Validate username contains only alphanumeric, underscore, hyphen, and dot
    if (!/^[a-zA-Z0-9._-]+$/.test(username)) {
      setError('Username can only contain letters, numbers, dots, underscores, and hyphens');
      return;
    }
    
    if (!password || password.length === 0) {
      setError('Password is required');
      return;
    }
    
    if (password.length < MIN_PASSWORD_LENGTH) {
      setError(`Password must be at least ${MIN_PASSWORD_LENGTH} characters long`);
      return;
    }
    
    if (password.length > MAX_PASSWORD_LENGTH) {
      setError(`Password must not exceed ${MAX_PASSWORD_LENGTH} characters`);
      return;
    }
    
    setLoading(true);

    try {
      const response = await authApi.login({ username, password });
      sessionStorage.setItem('auth_token', response.token);
      if (response.csrf_token) {
        sessionStorage.setItem('csrf_token', response.csrf_token);
      }
      navigate('/dashboard');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Login failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <style>{`
        .login-button:not(:disabled):hover {
          background-color: var(--primary-hover) !important;
          transform: translateY(-1px) !important;
          box-shadow: var(--shadow-md) !important;
        }
      `}</style>
    <div style={{
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      minHeight: '100vh',
      background: 'linear-gradient(135deg, var(--primary) 0%, var(--primary-hover) 100%)',
      padding: 'var(--spacing-xl)'
    }}>
      <div style={{
        backgroundColor: 'var(--bg-primary)',
        padding: 'var(--spacing-2xl)',
        borderRadius: 'var(--radius-lg)',
        boxShadow: 'var(--shadow-lg)',
        width: '100%',
        maxWidth: '700px'
      }}>
        {/* Logo */}
        <div style={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          marginBottom: 'var(--spacing-2xl)',
          textAlign: 'center'
        }}>
          <img
            src="/icon-snapshot-manager-bold.svg"
            alt="Snapshot Manager"
            style={{
              width: '90%',
              maxWidth: '550px',
              height: 'auto',
              display: 'block',
              margin: '0 auto'
            }}
          />
        </div>
        
        <form onSubmit={handleSubmit}>
          <div style={{ marginBottom: 'var(--spacing-lg)' }}>
            <label style={{
              display: 'block',
              marginBottom: 'var(--spacing-sm)',
              fontWeight: 600,
              fontSize: '0.875rem',
              color: 'var(--text-primary)'
            }}>
              Username
            </label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              placeholder="Enter your username"
              style={{
                width: '100%',
                padding: '0.75rem',
                border: '1px solid var(--border-subtle)',
                borderRadius: 'var(--radius-sm)',
                fontSize: '0.875rem',
                backgroundColor: 'var(--bg-primary)',
                color: 'var(--text-primary)',
                transition: 'all var(--transition-fast)'
              }}
            />
          </div>

          <div style={{ marginBottom: 'var(--spacing-xl)' }}>
            <label style={{
              display: 'block',
              marginBottom: 'var(--spacing-sm)',
              fontWeight: 600,
              fontSize: '0.875rem',
              color: 'var(--text-primary)'
            }}>
              Password
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              placeholder="Enter your password"
              style={{
                width: '100%',
                padding: '0.75rem',
                border: '1px solid var(--border-subtle)',
                borderRadius: 'var(--radius-sm)',
                fontSize: '0.875rem',
                backgroundColor: 'var(--bg-primary)',
                color: 'var(--text-primary)',
                transition: 'all var(--transition-fast)'
              }}
            />
          </div>

          {error && (
            <div style={{
              padding: 'var(--spacing-md)',
              marginBottom: 'var(--spacing-lg)',
              backgroundColor: 'var(--danger-light)',
              color: 'var(--danger)',
              borderRadius: 'var(--radius-sm)',
              fontSize: '0.875rem',
              border: '1px solid var(--danger)',
              display: 'flex',
              alignItems: 'center',
              gap: 'var(--spacing-sm)'
            }}>
              <span>⚠️</span>
              <span>{error}</span>
            </div>
          )}

          <button
            type="submit"
            disabled={loading}
            className="login-button"
            style={{
              width: '100%',
              padding: '0.875rem',
              backgroundColor: loading ? 'var(--bg-tertiary)' : 'var(--primary)',
              color: 'white',
              border: 'none',
              borderRadius: 'var(--radius-sm)',
              fontSize: '0.875rem',
              fontWeight: 600,
              cursor: loading ? 'not-allowed' : 'pointer',
              transition: 'all var(--transition-fast)',
              boxShadow: loading ? 'none' : 'var(--shadow-sm)'
            }}
          >
            {loading ? 'Signing in...' : 'Sign In'}
          </button>
        </form>

      </div>
    </div>
    </>
  );
}

// 
