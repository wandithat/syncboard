import apiClient from './client';
import type { Column } from '../types';

export const columnsApi = {
  // GET /api/boards/{board_id}/columns — уже есть на бэкенде
  getByBoard: async (boardId: string): Promise<Column[]> => {
    const response = await apiClient.get(`/boards/${boardId}/columns`);
    return response.data;
  },

  create: async (boardId: string, title: string): Promise<Column> => {
    const response = await apiClient.post('/columns', { board_id: boardId, title });
    return response.data;
  },
};