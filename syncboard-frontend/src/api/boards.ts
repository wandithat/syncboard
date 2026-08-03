import apiClient from './client';
import type { Board } from '../types';

export const boardsApi = {
    list: async (): Promise<Board[]> => {
        const response = await apiClient.get('/boards');
        return response.data;
    },
    create: async (title: string): Promise<Board> => {
        const response = await apiClient.post('/boards', { title });
        return response.data;
    },
};