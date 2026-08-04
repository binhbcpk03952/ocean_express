import axios from 'axios';

const api = axios.create({
  // Đọc từ .env (VITE_API_BASE_URL); fallback về localhost cho môi trường dev.
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
  timeout: 10000,
});

// Interceptor add token vào header
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('ocean_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Interceptor xử lý lỗi chung (VD: hết hạn token)
api.interceptors.response.use(
  (response) => {
    return response.data;
  },
  (error) => {
    if (error.response && error.response.status === 401) {
      // Bỏ qua việc redirect nếu API đang gọi là API đăng nhập
      const isLoginApi = error.config?.url?.includes('/auth/login') || error.config?.url?.includes('/auth/shop/login');
      
      if (!isLoginApi) {
        // Xóa token và redirect về login nếu auth fail (đối với các API khác)
        localStorage.removeItem('ocean_token');
        localStorage.removeItem('ocean_user');
        
        // Tránh redirect loop nếu đang ở trang login
        if (window.location.pathname !== '/login' && window.location.pathname !== '/shop/login') {
          window.location.href = '/login';
        }
      }
    }
    return Promise.reject(error);
  }
);

export default api;
