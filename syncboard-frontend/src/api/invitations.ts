import apiClient from './client';

export interface BoardInvitation {
  id: string;
  board_id: string;
  board_title: string;
  invitee_email: string;
  invited_by: string;
  inviter_email?: string;
  status: 'pending' | 'accepted' | 'declined';
  created_at: string;
}

export interface BoardMember {
  user_id: string;
  email: string;
  role: string;
  joined_at: string;
}

export const invitationsApi = {
  listMine: async (): Promise<BoardInvitation[]> => {
    const response = await apiClient.get('/invitations');
    return response.data;
  },
  create: async (boardId: string, email: string): Promise<BoardInvitation> => {
    const response = await apiClient.post(`/boards/${boardId}/invitations`, { email });
    return response.data;
  },
  accept: async (id: string): Promise<{ status: string; board_id: string }> => {
    const response = await apiClient.post(`/invitations/${id}/accept`);
    return response.data;
  },
  decline: async (id: string): Promise<void> => {
    await apiClient.post(`/invitations/${id}/decline`);
  },
  listMembers: async (boardId: string): Promise<BoardMember[]> => {
    const response = await apiClient.get(`/boards/${boardId}/members`);
    return response.data;
  },
};
