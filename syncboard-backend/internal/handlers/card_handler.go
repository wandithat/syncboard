package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"syncboard/internal/middleware"
	"syncboard/internal/models"
	"syncboard/internal/realtime"
	"syncboard/internal/repository"

	"github.com/gorilla/mux"
)

type CardHandler struct {
	repo    *repository.CardRepository
	columns *repository.ColumnRepository
	boards  *repository.BoardRepository
	hub     *realtime.Hub
}

func NewCardHandler(repo *repository.CardRepository, columns *repository.ColumnRepository, boards *repository.BoardRepository, hub *realtime.Hub) *CardHandler {
	return &CardHandler{repo: repo, columns: columns, boards: boards, hub: hub}
}

func (h *CardHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		jsonError(w, http.StatusBadRequest, "title is required")
		return
	}
	boardID, err := h.columns.BoardIDByColumn(r.Context(), req.ColumnID)
	if err != nil {
		jsonError(w, http.StatusNotFound, "column not found")
		return
	}
	ok, err := h.boards.IsMember(r.Context(), boardID, middleware.GetUserID(r.Context()))
	if err != nil || !ok {
		jsonError(w, http.StatusNotFound, "column not found")
		return
	}
	card := &models.Card{ColumnID: req.ColumnID, Title: req.Title}
	if err := h.repo.Create(r.Context(), card); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.hub.Broadcast(boardID, "card.created", card)
	jsonResponse(w, http.StatusCreated, card)
}

func (h *CardHandler) Move(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var req models.MoveCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	boardID, err := h.repo.BoardIDByCard(r.Context(), id)
	if err != nil {
		jsonError(w, http.StatusNotFound, "card not found")
		return
	}
	targetBoardID, err := h.columns.BoardIDByColumn(r.Context(), req.ColumnID)
	if err != nil || targetBoardID != boardID {
		jsonError(w, http.StatusBadRequest, "invalid target column")
		return
	}
	ok, err := h.boards.IsMember(r.Context(), boardID, middleware.GetUserID(r.Context()))
	if err != nil || !ok {
		jsonError(w, http.StatusNotFound, "card not found")
		return
	}
	if req.Position < 0 {
		jsonError(w, http.StatusBadRequest, "position must be non-negative")
		return
	}
	if err := h.repo.Move(r.Context(), id, req.ColumnID, req.Position); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.hub.Broadcast(boardID, "card.moved", map[string]interface{}{"id": id, "column_id": req.ColumnID, "position": req.Position})
	jsonResponse(w, http.StatusOK, map[string]string{"status": "moved"})
}

func (h *CardHandler) GetByColumn(w http.ResponseWriter, r *http.Request) {
	columnID := mux.Vars(r)["column_id"]

	boardID, err := h.columns.BoardIDByColumn(r.Context(), columnID)
	if err != nil {
		jsonError(w, http.StatusNotFound, "column not found")
		return
	}
	ok, err := h.boards.IsMember(r.Context(), boardID, middleware.GetUserID(r.Context()))
	if err != nil || !ok {
		jsonError(w, http.StatusNotFound, "column not found")
		return
	}
	cards, err := h.repo.GetByColumnID(r.Context(), columnID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if cards == nil {
		cards = []models.Card{} // чтобы в JSON приходил [], а не null
	}

	jsonResponse(w, http.StatusOK, cards)
}

func (h *CardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	boardID, err := h.repo.BoardIDByCard(r.Context(), id)
	if err != nil {
		jsonError(w, http.StatusNotFound, "card not found")
		return
	}
	ok, err := h.boards.IsMember(r.Context(), boardID, middleware.GetUserID(r.Context()))
	if err != nil || !ok {
		jsonError(w, http.StatusNotFound, "card not found")
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.hub.Broadcast(boardID, "card.deleted", map[string]string{"id": id})
	w.WriteHeader(http.StatusNoContent)
}
