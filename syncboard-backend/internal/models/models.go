package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // ВАЖНО: никогда не уходит в JSON
	CreatedAt    time.Time `json:"created_at"`
}

type Board struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	OwnerID   string    `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	CanInvite bool      `json:"can_invite,omitempty"`
}

type BoardMember struct {
	BoardID  string    `json:"board_id"`
	UserID   string    `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type Column struct {
	ID       string `json:"id"`
	BoardID  string `json:"board_id"`
	Title    string `json:"title"`
	Position int    `json:"position"`
}

type Card struct {
	ID          string    `json:"id"`
	ColumnID    string    `json:"column_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Position    int       `json:"position"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DTO для запросов
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateBoardRequest struct {
	Title string `json:"title"`
}

type CreateColumnRequest struct {
	BoardID string `json:"board_id"`
	Title   string `json:"title"`
}

type CreateCardRequest struct {
	ColumnID string `json:"column_id"`
	Title    string `json:"title"`
}

type MoveCardRequest struct {
	ColumnID string `json:"column_id"`
	Position int    `json:"position"`
}

type BoardInvitation struct {
	ID           string    `json:"id"`
	BoardID      string    `json:"board_id"`
	BoardTitle   string    `json:"board_title"`
	InviteeEmail string    `json:"invitee_email"`
	InvitedBy    string    `json:"invited_by"`
	InviterEmail string    `json:"inviter_email,omitempty"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type BoardMemberDetails struct {
	UserID   string    `json:"user_id"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type ReorderColumnsRequest struct {
	ColumnIDs []string `json:"column_ids"`
}

type CreateInvitationRequest struct {
	Email string `json:"email"`
}
