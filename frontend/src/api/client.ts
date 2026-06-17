import axios from 'axios';

// Use relative URL for containerized deployment (Nginx proxy)
// Falls back to direct backend URL for local development
const API_BASE_URL = import.meta.env.VITE_API_URL || '/api';

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add token and CSRF token to requests if available
apiClient.interceptors.request.use((config) => {
  const token = sessionStorage.getItem('auth_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  
  const csrfToken = sessionStorage.getItem('csrf_token');
  if (csrfToken) {
    config.headers['X-CSRF-Token'] = csrfToken;
  }
  
  return config;
});

// Handle 401 and 403 responses
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    // Handle unauthorized (invalid/expired JWT token)
    if (error.response?.status === 401) {
      sessionStorage.removeItem('auth_token');
      sessionStorage.removeItem('csrf_token');
      window.location.href = '/login';
    }
    // Handle forbidden (invalid/expired CSRF token)
    if (error.response?.status === 403 && error.response?.data?.error?.includes('CSRF')) {
      sessionStorage.removeItem('auth_token');
      sessionStorage.removeItem('csrf_token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export default apiClient;

// 
