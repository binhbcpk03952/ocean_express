import { defineStore } from 'pinia';
import { ref } from 'vue';
import api from '../services/api';

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('ocean_token') || null);
  const user = ref(JSON.parse(localStorage.getItem('ocean_user')) || null);

  const login = async (identifier, password) => {
    try {
      const response = await api.post('/auth/login', { identifier, password });
      if (response.success) {
        token.value = response.data.token;
        user.value = {
          id: response.data.employee.id,
          name: response.data.employee.name,
          phone: response.data.employee.phone,
          email: response.data.employee.email,
          role: response.data.employee.role,
        };
        localStorage.setItem('ocean_token', token.value);
        localStorage.setItem('ocean_user', JSON.stringify(user.value));
        return true;
      }
      return false;
    } catch (error) {
      throw error;
    }
  };

  // Đăng nhập portal cho đối tác Shop (email/phone + mật khẩu). Trả về JWT role 'shop'.
  const loginShop = async (identifier, password) => {
    try {
      const response = await api.post('/auth/shop/login', { identifier, password });
      if (response.success) {
        token.value = response.data.token;
        user.value = {
          id: response.data.shop.id,
          name: response.data.shop.name,
          email: response.data.shop.email,
          role: response.data.shop.role, // 'shop'
        };
        localStorage.setItem('ocean_token', token.value);
        localStorage.setItem('ocean_user', JSON.stringify(user.value));
        return true;
      }
      return false;
    } catch (error) {
      throw error;
    }
  };

  const logout = () => {
    token.value = null;
    user.value = null;
    localStorage.removeItem('ocean_token');
    localStorage.removeItem('ocean_user');
  };

  const isAuthenticated = () => {
    return !!token.value;
  };

  const hasRole = (...roles) => {
    return !!user.value && roles.includes(user.value.role);
  };

  // Route mặc định sau khi đăng nhập tùy theo role.
  // Admin -> admin panel; shop -> shop portal; còn lại -> giao diện member.
  const homeRoute = () => {
    if (user.value?.role === 'admin') return { name: 'Dashboard' };
    if (user.value?.role === 'shop') return { name: 'ShopDashboard' };
    return { name: 'MemberTasks' };
  };

  return { token, user, login, loginShop, logout, isAuthenticated, hasRole, homeRoute };
});
