package repository

import (
	"context"
	"syncboard/internal/models"
)

type BoardRepository struct {
	db *DB
}

func NewBoardRepository(db *DB) *BoardRepository {
	return &BoardRepository{db: db}
}

func (r *BoardRepository) Create(ctx context.Context, board *models.Board) error {
	query := `
		INSERT INTO boards (title, owner_id)
		VALUES ($1, $2)
		RETURNING id, created_at
	`
	err := r.db.Pool.QueryRow(ctx, query, board.Title, board.OwnerID).
		Scan(&board.ID, &board.CreatedAt)
	if err != nil {
		return err
	}

	// Добавляем владельца как member с ролью 'admin'
	_, err = r.db.Pool.Exec(ctx, `
		INSERT INTO board_members (board_id, user_id, role)
		VALUES ($1, $2, 'admin')
	`, board.ID, board.OwnerID)
	return err
}

func (r *BoardRepository) GetByID(ctx context.Context, id string) (*models.Board, error) {
	board := &models.Board{}
	query := `SELECT id, title, owner_id, created_at FROM boards WHERE id = $1`
	err := r.db.Pool.QueryRow(ctx, query, id).
		Scan(&board.ID, &board.Title, &board.OwnerID, &board.CreatedAt)
	return board, err
}

func (r *BoardRepository) GetUserBoards(ctx context.Context, userID string) ([]models.Board, error) {
	query := `
		SELECT b.id, b.title, b.owner_id, b.created_at
		FROM boards b
		JOIN board_members bm ON b.id = bm.board_id
		WHERE bm.user_id = $1
		ORDER BY b.created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var boards []models.Board
	for rows.Next() {
		var b models.Board
		if err := rows.Scan(&b.ID, &b.Title, &b.OwnerID, &b.CreatedAt); err != nil {
			return nil, err
		}
		boards = append(boards, b)
	}
	return boards, nil
}
