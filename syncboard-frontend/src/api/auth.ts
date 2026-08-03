import apiClient from './client';
import type { User } from '../types';

export const authApi = {
  register: async (email: string, password: string): Promise<User> => {
    const response = await apiClient.post('/auth/register', { email, password });
    return response.data;
  },

  login: async (email: string, password: string): Promise<string> => {
    const response = await apiClient.post('/auth/login', { email, password });
    return response.data.token;
  },
};