CREATE INDEX IF NOT EXISTS idx_board_members_user ON board_members(user_id);
CREATE INDEX IF NOT EXISTS idx_cards_column_position ON cards(column_id, position);
CREATE INDEX IF NOT EXISTS idx_columns_board_position ON columns(board_id, position);
