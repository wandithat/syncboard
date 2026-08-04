import apiClient from './client';
import type { Card } from '../types';

export const cardsApi = {
  // GET /api/columns/{column_id}/cards — новый эндпоинт (см. Часть 5)
  getByColumn: async (columnId: string): Promise<Card[]> => {
    const response = await apiClient.get(`/columns/${columnId}/cards`);
    return response.data;
  },

  create: async (columnId: string, title: string): Promise<Card> => {
    const response = await apiClient.post('/cards', { column_id: columnId, title });
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/cards/${id}`);
  },
};