package handlers

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"

	"syncboard/internal/middleware"
	"syncboard/internal/models"
	"syncboard/internal/realtime"
	"syncboard/internal/repository"

	"github.com/gorilla/mux"
)

type InvitationHandler struct {
	boards *repository.BoardRepository
	users  *repository.UserRepository
	hub    *realtime.Hub
}

func NewInvitationHandler(boards *repository.BoardRepository, users *repository.UserRepository, hub *realtime.Hub) *InvitationHandler {
	return &InvitationHandler{boards: boards, users: users, hub: hub}
}

func (h *InvitationHandler) Create(w http.ResponseWriter, r *http.Request) {
	boardID := mux.Vars(r)["board_id"]
	userID := middleware.GetUserID(r.Context())
	admin, err := h.boards.IsAdmin(r.Context(), boardID, userID)
	if err != nil || !admin {
		jsonError(w, http.StatusForbidden, "only board admins can invite members")
		return
	}

	var req models.CreateInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		jsonError(w, http.StatusBadRequest, "valid email is required")
		return
	}
	currentUser, err := h.users.GetByID(r.Context(), userID)
	if err != nil || currentUser == nil {
		jsonError(w, http.StatusUnauthorized, "user not found")
		return
	}
	if strings.EqualFold(currentUser.Email, email) {
		jsonError(w, http.StatusBadRequest, "you are already a board member")
		return
	}
	member, err := h.users.GetByEmail(r.Context(), email)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "could not check invitee")
		return
	}
	if member != nil {
		isMember, err := h.boards.IsMember(r.Context(), boardID, member.ID)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "could not check board membership")
			return
		}
		if isMember {
			jsonError(w, http.StatusConflict, "user is already a board member")
			return
		}
	}

	invitation := &models.BoardInvitation{BoardID: boardID, InviteeEmail: email, InvitedBy: userID}
	if err := h.boards.CreateInvitation(r.Context(), invitation); err != nil {
		jsonError(w, http.StatusConflict, "a pending invitation for this email already exists")
		return
	}
	h.hub.Broadcast(boardID, "invitation.created", map[string]string{"email": email})
	jsonResponse(w, http.StatusCreated, invitation)
}

func (h *InvitationHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	user, err := h.users.GetByID(r.Context(), middleware.GetUserID(r.Context()))
	if err != nil || user == nil {
		jsonError(w, http.StatusUnauthorized, "user not found")
		return
	}
	invitations, err := h.boards.ListPendingInvitations(r.Context(), user.Email)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "could not load invitations")
		return
	}
	jsonResponse(w, http.StatusOK, invitations)
}

func (h *InvitationHandler) Accept(w http.ResponseWriter, r *http.Request) {
	h.respond(w, r, true)
}

func (h *InvitationHandler) Decline(w http.ResponseWriter, r *http.Request) {
	h.respond(w, r, false)
}

func (h *InvitationHandler) respond(w http.ResponseWriter, r *http.Request, accept bool) {
	user, err := h.users.GetByID(r.Context(), middleware.GetUserID(r.Context()))
	if err != nil || user == nil {
		jsonError(w, http.StatusUnauthorized, "user not found")
		return
	}
	boardID, err := h.boards.RespondToInvitation(r.Context(), mux.Vars(r)["id"], user.Email, user.ID, accept)
	if err != nil {
		jsonError(w, http.StatusNotFound, "invitation not found or no longer pending")
		return
	}
	status := "declined"
	if accept {
		status = "accepted"
		h.hub.Broadcast(boardID, "member.joined", map[string]string{"user_id": user.ID, "email": user.Email})
	}
	jsonResponse(w, http.StatusOK, map[string]string{"status": status, "board_id": boardID})
}

func (h *InvitationHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	boardID := mux.Vars(r)["board_id"]
	ok, err := h.boards.IsMember(r.Context(), boardID, middleware.GetUserID(r.Context()))
	if err != nil || !ok {
		jsonError(w, http.StatusNotFound, "board not found")
		return
	}
	members, err := h.boards.ListMembers(r.Context(), boardID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "could not load members")
		return
	}
	jsonResponse(w, http.StatusOK, members)
}
