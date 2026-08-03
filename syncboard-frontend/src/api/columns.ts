import apiClient from './client';
import type { Column } from '../types';

export const columnsApi = {
    create: async (boardID: string, title: string): Promise<Column> => {
        const response = await apiClient.post('/columns', { board_id: boardID, title });
        return response.data;
    },
};