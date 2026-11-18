package repository

import (
	"context"
	"errors"
	"syncboard/internal/models"
)

type BoardRepository struct {
	db *DB
}

func NewBoardRepository(db *DB) *BoardRepository {
	return &BoardRepository{db: db}
}

func (r *BoardRepository) Create(ctx context.Context, board *models.Board) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO boards (title, owner_id) VALUES ($1, $2)
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO boards (title, owner_id) VALUES ($1, $2)
		RETURNING id, created_at
	`, board.Title, board.OwnerID).Scan(&board.ID, &board.CreatedAt)
	`, board.Title, board.OwnerID).Scan(&board.ID, &board.CreatedAt)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO board_members (board_id, user_id, role) VALUES ($1, $2, 'admin')
	_, err = tx.Exec(ctx, `
		INSERT INTO board_members (board_id, user_id, role) VALUES ($1, $2, 'admin')
	`, board.ID, board.OwnerID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
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

func (r *BoardRepository) IsMember(ctx context.Context, boardID, userID string) (bool, error) {
	var exists bool
	err := r.db.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM board_members WHERE board_id=$1 AND user_id=$2)`, boardID, userID).Scan(&exists)
	return exists, err
}

func (r *BoardRepository) IsAdmin(ctx context.Context, boardID, userID string) (bool, error) {
	var allowed bool
	err := r.db.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM board_members
			WHERE board_id = $1 AND user_id = $2 AND role IN ('owner', 'admin')
		)
	`, boardID, userID).Scan(&allowed)
	return allowed, err
}

func (r *BoardRepository) ReorderColumns(ctx context.Context, boardID string, columnIDs []string) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var lockedBoardID string
	if err := tx.QueryRow(ctx, `SELECT id FROM boards WHERE id = $1 FOR UPDATE`, boardID).Scan(&lockedBoardID); err != nil {
		return err
	}

	var count int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM columns WHERE board_id = $1`, boardID).Scan(&count); err != nil {
		return err
	}
	if count != len(columnIDs) {
		return errors.New("column list does not match board columns")
	}

	// First move all positions out of the target range to avoid transient collisions
	// if a uniqueness constraint is added in a later migration.
	if _, err := tx.Exec(ctx, `UPDATE columns SET position = position + 1000000 WHERE board_id = $1`, boardID); err != nil {
		return err
	}
	for position, columnID := range columnIDs {
		tag, err := tx.Exec(ctx, `UPDATE columns SET position = $1 WHERE id = $2 AND board_id = $3`, position, columnID, boardID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return errors.New("column does not belong to board")
		}
	}
	return tx.Commit(ctx)
}

func (r *BoardRepository) CreateInvitation(ctx context.Context, invitation *models.BoardInvitation) error {
	return r.db.Pool.QueryRow(ctx, `
		INSERT INTO board_invitations (board_id, invitee_email, invited_by)
		VALUES ($1, LOWER($2), $3)
		RETURNING id, status, created_at
	`, invitation.BoardID, invitation.InviteeEmail, invitation.InvitedBy).
		Scan(&invitation.ID, &invitation.Status, &invitation.CreatedAt)
}

func (r *BoardRepository) ListPendingInvitations(ctx context.Context, email string) ([]models.BoardInvitation, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT i.id, i.board_id, b.title, i.invitee_email, i.invited_by,
		       u.email, i.status, i.created_at
		FROM board_invitations i
		JOIN boards b ON b.id = i.board_id
		JOIN users u ON u.id = i.invited_by
		WHERE LOWER(i.invitee_email) = LOWER($1) AND i.status = 'pending'
		ORDER BY i.created_at DESC
	`, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invitations := make([]models.BoardInvitation, 0)
	for rows.Next() {
		var invitation models.BoardInvitation
		if err := rows.Scan(&invitation.ID, &invitation.BoardID, &invitation.BoardTitle,
			&invitation.InviteeEmail, &invitation.InvitedBy, &invitation.InviterEmail,
			&invitation.Status, &invitation.CreatedAt); err != nil {
			return nil, err
		}
		invitations = append(invitations, invitation)
	}
	return invitations, rows.Err()
}

func (r *BoardRepository) RespondToInvitation(ctx context.Context, invitationID, email, userID string, accept bool) (string, error) {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var boardID string
	var status string
	err = tx.QueryRow(ctx, `
		SELECT board_id, status FROM board_invitations
		WHERE id = $1 AND LOWER(invitee_email) = LOWER($2)
		FOR UPDATE
	`, invitationID, email).Scan(&boardID, &status)
	if err != nil {
		return "", err
	}
	if status != "pending" {
		return "", errors.New("invitation is no longer pending")
	}

	newStatus := "declined"
	if accept {
		newStatus = "accepted"
		if _, err := tx.Exec(ctx, `
			INSERT INTO board_members (board_id, user_id, role)
			VALUES ($1, $2, 'member')
			ON CONFLICT (board_id, user_id) DO NOTHING
		`, boardID, userID); err != nil {
			return "", err
		}
	}
	if _, err := tx.Exec(ctx, `
		UPDATE board_invitations SET status = $1, responded_at = NOW() WHERE id = $2
	`, newStatus, invitationID); err != nil {
		return "", err
	}
	return boardID, tx.Commit(ctx)
}

func (r *BoardRepository) ListMembers(ctx context.Context, boardID string) ([]models.BoardMemberDetails, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT bm.user_id, u.email, bm.role, bm.joined_at
		FROM board_members bm
		JOIN users u ON u.id = bm.user_id
		WHERE bm.board_id = $1
		ORDER BY CASE bm.role WHEN 'owner' THEN 0 WHEN 'admin' THEN 1 ELSE 2 END, bm.joined_at
	`, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]models.BoardMemberDetails, 0)
	for rows.Next() {
		var member models.BoardMemberDetails
		if err := rows.Scan(&member.UserID, &member.Email, &member.Role, &member.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, rows.Err()
}
