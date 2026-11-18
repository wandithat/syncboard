package repository

import (
	"context"
	"syncboard/internal/models"
)

type ColumnRepository struct {
	db *DB
}

func NewColumnRepository(db *DB) *ColumnRepository {
	return &ColumnRepository{db: db}
}

func (r *ColumnRepository) Create(ctx context.Context, col *models.Column) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var ownerID string
	if err := tx.QueryRow(ctx, `SELECT owner_id FROM boards WHERE id=$1 FOR UPDATE`, col.BoardID).Scan(&ownerID); err != nil {
		return err
	}
	if err := tx.QueryRow(ctx, `SELECT COALESCE(MAX(position), -1)+1 FROM columns WHERE board_id=$1`, col.BoardID).Scan(&col.Position); err != nil {
		return err
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO columns (board_id, title, position) VALUES ($1, $2, $3) RETURNING id
	`, col.BoardID, col.Title, col.Position).Scan(&col.ID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *ColumnRepository) GetByBoardID(ctx context.Context, boardID string) ([]models.Column, error) {
	query := `
		SELECT id, board_id, title, position
		FROM columns
		WHERE board_id = $1
		ORDER BY position ASC
	`
	rows, err := r.db.Pool.Query(ctx, query, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []models.Column
	for rows.Next() {
		var c models.Column
		if err := rows.Scan(&c.ID, &c.BoardID, &c.Title, &c.Position); err != nil {
			return nil, err
		}
		cols = append(cols, c)
	}
	return cols, nil
}

func (r *ColumnRepository) UpdateTitle(ctx context.Context, id, title string) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE columns SET title = $1 WHERE id = $2`, title, id)
	return err
}

func (r *ColumnRepository) BoardIDByColumn(ctx context.Context, columnID string) (string, error) {
	var boardID string
	err := r.db.Pool.QueryRow(ctx, `SELECT board_id FROM columns WHERE id=$1`, columnID).Scan(&boardID)
	return boardID, err
}
