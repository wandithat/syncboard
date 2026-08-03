import apiClient from './client';
import type { Card } from '../types';

export const cardsApi = {
    create: async (columnId: string, title: string): Promise<Card> => {
        const response = await apiClient.post('/cards', {column_id: columnId, title});
        return response.data;
    },
    delete: async (id: string): Promise<void> => {
        await apiClient.delete(`/cards/${id}`);
    },
};