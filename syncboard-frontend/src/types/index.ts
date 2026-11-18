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
  can_invite?: boolean;
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

// НОВОЕ: колонка с вложенными карточками
export interface ColumnWithCards extends Column {
  cards: Card[];
}

// НОВОЕ: доска с колонками и карточками
export interface BoardWithDetails extends Board {
  columns: ColumnWithCards[];
}