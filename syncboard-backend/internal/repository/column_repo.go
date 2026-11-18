package repository

import (
	"context"
	"fmt"
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

// ReorderColumns stores the requested column order. The board row is locked so
// concurrent reorder/create operations for the same board cannot interleave.
// The request must contain every column on the board exactly once.
func (r *ColumnRepository) ReorderColumns(ctx context.Context, boardID string, columnIDs []string) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var lockedBoardID string
	if err := tx.QueryRow(ctx, `SELECT id FROM boards WHERE id = $1 FOR UPDATE`, boardID).Scan(&lockedBoardID); err != nil {
		return err
	}

	rows, err := tx.Query(ctx, `SELECT id::text FROM columns WHERE board_id = $1 FOR UPDATE`, boardID)
	if err != nil {
		return err
	}
	existing := make(map[string]struct{})
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		existing[id] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	if len(columnIDs) != len(existing) {
		return fmt.Errorf("column_ids must contain all %d columns on the board", len(existing))
	}
	seen := make(map[string]struct{}, len(columnIDs))
	for position, id := range columnIDs {
		if _, ok := existing[id]; !ok {
			return fmt.Errorf("column %q does not belong to board", id)
		}
		if _, ok := seen[id]; ok {
			return fmt.Errorf("duplicate column id %q", id)
		}
		seen[id] = struct{}{}
		if _, err := tx.Exec(ctx, `UPDATE columns SET position = $1 WHERE id = $2 AND board_id = $3`, position, id, boardID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
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
