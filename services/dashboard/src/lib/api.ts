import axios from 'axios';
import { useAuthStore } from '../store/auth';

const api = axios.create({
  baseURL: (import.meta as any).env.VITE_API_URL,
});

api.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      useAuthStore.getState().logout();
    }
    return Promise.reject(error);
  }
);

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  role: 'admin' | 'user';
}

export interface User {
  id: number;
  email: string;
  role: string;
}

export const authApi = {
  login: async (data: LoginRequest) => {
    const response = await api.post<{ token: string }>('/api/auth/login', data);
    return response.data;
  },
  register: async (data: RegisterRequest) => {
    const response = await api.post<User>('/api/auth/register', data);
    return response.data;
  },
  getProfile: async () => {
    const response = await api.get<User>('/api/user/profile');
    return response.data;
  },
};

export const usersApi = {
  getUsers: async () => {
    const response = await api.get<User[]>('/api/admin/users');
    return response.data;
  },
};

export default api; 