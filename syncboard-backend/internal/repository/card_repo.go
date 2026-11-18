package repository

import (
	"context"
	"syncboard/internal/models"
)

type CardRepository struct {
	db *DB
}

func NewCardRepository(db *DB) *CardRepository {
	return &CardRepository{db: db}
}

func (r *CardRepository) Create(ctx context.Context, card *models.Card) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var boardID string
	if err := tx.QueryRow(ctx, `SELECT board_id FROM columns WHERE id=$1 FOR UPDATE`, card.ColumnID).Scan(&boardID); err != nil {
		return err
	}
	if err := tx.QueryRow(ctx, `SELECT COALESCE(MAX(position), -1)+1 FROM cards WHERE column_id=$1`, card.ColumnID).Scan(&card.Position); err != nil {
		return err
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO cards (column_id, title, description, position)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`, card.ColumnID, card.Title, card.Description, card.Position).Scan(&card.ID, &card.CreatedAt, &card.UpdatedAt)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *CardRepository) GetByColumnID(ctx context.Context, columnID string) ([]models.Card, error) {
	query := `
		SELECT id, column_id, title, description, position, created_at, updated_at
		FROM cards
		WHERE column_id = $1
		ORDER BY position ASC
	`
	rows, err := r.db.Pool.Query(ctx, query, columnID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var c models.Card
		if err := rows.Scan(&c.ID, &c.ColumnID, &c.Title, &c.Description,
			&c.Position, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, nil
}

func (r *CardRepository) Move(ctx context.Context, id, columnID string, position int) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var sourceColumn string
	err = tx.QueryRow(ctx, `SELECT column_id FROM cards WHERE id=$1 FOR UPDATE`, id).Scan(&sourceColumn)
	if err != nil {
		return err
	}

	// Lock both columns to serialize concurrent reorder operations.
	if _, err = tx.Exec(ctx, `SELECT id FROM columns WHERE id IN ($1,$2) FOR UPDATE`, sourceColumn, columnID); err != nil {
		return err
	}

	rows, err := tx.Query(ctx, `SELECT id FROM cards WHERE column_id=$1 AND id<>$2 ORDER BY position,id`, columnID, id)
	if err != nil {
		return err
	}
	targetIDs := make([]string, 0)
	for rows.Next() {
		var cardID string
		if err := rows.Scan(&cardID); err != nil {
			rows.Close()
			return err
		}
		targetIDs = append(targetIDs, cardID)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	if position < 0 {
		position = 0
	}
	if position > len(targetIDs) {
		position = len(targetIDs)
	}
	targetIDs = append(targetIDs, "")
	copy(targetIDs[position+1:], targetIDs[position:])
	targetIDs[position] = id

	// Compact the source column when the card moves out of it.
	if sourceColumn != columnID {
		sourceRows, err := tx.Query(ctx, `SELECT id FROM cards WHERE column_id=$1 AND id<>$2 ORDER BY position,id`, sourceColumn, id)
		if err != nil {
			return err
		}
		sourceIDs := make([]string, 0)
		for sourceRows.Next() {
			var cardID string
			if err := sourceRows.Scan(&cardID); err != nil {
				sourceRows.Close()
				return err
			}
			sourceIDs = append(sourceIDs, cardID)
		}
		sourceRows.Close()
		if err := sourceRows.Err(); err != nil {
			return err
		}
		for i, cardID := range sourceIDs {
			if _, err := tx.Exec(ctx, `UPDATE cards SET position=$1 WHERE id=$2`, i, cardID); err != nil {
				return err
			}
		}
	}

	for i, cardID := range targetIDs {
		if _, err := tx.Exec(ctx, `UPDATE cards SET column_id=$1, position=$2, updated_at=NOW() WHERE id=$3`, columnID, i, cardID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *CardRepository) Delete(ctx context.Context, id string) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var columnID string
	if err := tx.QueryRow(ctx, `SELECT column_id FROM cards WHERE id=$1 FOR UPDATE`, id).Scan(&columnID); err != nil {
		return err
	}
	var lockedColumn string
	if err := tx.QueryRow(ctx, `SELECT id FROM columns WHERE id=$1 FOR UPDATE`, columnID).Scan(&lockedColumn); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM cards WHERE id=$1`, id); err != nil {
		return err
	}

	rows, err := tx.Query(ctx, `SELECT id FROM cards WHERE column_id=$1 ORDER BY position,id`, columnID)
	if err != nil {
		return err
	}
	ids := make([]string, 0)
	for rows.Next() {
		var cardID string
		if err := rows.Scan(&cardID); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, cardID)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for i, cardID := range ids {
		if _, err := tx.Exec(ctx, `UPDATE cards SET position=$1 WHERE id=$2`, i, cardID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *CardRepository) BoardIDByCard(ctx context.Context, cardID string) (string, error) {
	var boardID string
	err := r.db.Pool.QueryRow(ctx, `SELECT col.board_id FROM cards c JOIN columns col ON col.id=c.column_id WHERE c.id=$1`, cardID).Scan(&boardID)
	return boardID, err
}
