import apiClient from './client';
import type { Card } from '../types';

export const cardsApi = {
  getByColumn: async (columnId: string): Promise<Card[]> => {
    const response = await apiClient.get(`/columns/${columnId}/cards`);
    return response.data;
  },

  create: async (columnId: string, title: string): Promise<Card> => {
    const response = await apiClient.post('/cards', { column_id: columnId, title });
    return response.data;
  },

  // НОВОЕ: перемещение карточки
  move: async (id: string, columnId: string, position: number): Promise<void> => {
    await apiClient.patch(`/cards/${id}/move`, {
      column_id: columnId,
      position,
    });
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/cards/${id}`);
  },
};
