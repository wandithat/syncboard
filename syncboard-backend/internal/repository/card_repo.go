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
	var maxPos int
	_ = r.db.Pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(position), -1) FROM cards WHERE column_id = $1
	`, card.ColumnID).Scan(&maxPos)
	card.Position = maxPos + 1

	query := `
		INSERT INTO cards (column_id, title, description, position)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	return r.db.Pool.QueryRow(ctx, query, card.ColumnID, card.Title, card.Description, card.Position).
		Scan(&card.ID, &card.CreatedAt, &card.UpdatedAt)
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
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE cards
		SET column_id = $1, position = $2, updated_at = NOW()
		WHERE id = $3
	`, columnID, position, id)
	return err
}

func (r *CardRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM cards WHERE id = $1`, id)
	return err
}
