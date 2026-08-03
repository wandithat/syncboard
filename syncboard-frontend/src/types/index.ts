export interface User {
  id: string;
  email: string;
  created_at: string;
}

export interface Board {
  id: string;
  title: string;
  owner_id: string;
  created_at: string;
}

export interface Column {
  id: string;
  board_id: string;
  title: string;
  position: number;
}

export interface Card {
  id: string;
  column_id: string;
  title: string;
  description: string;
  position: number;
  created_at: string;
  updated_at: string;
}

export interface BoardWithDetails extends Board {
  columns: (Column & {cards: Card[]})[]
}