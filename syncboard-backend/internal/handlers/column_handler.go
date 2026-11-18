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

type ColumnHandler struct {
	repo   *repository.ColumnRepository
	boards *repository.BoardRepository
	hub    *realtime.Hub
}

func NewColumnHandler(repo *repository.ColumnRepository, boards *repository.BoardRepository, hub *realtime.Hub) *ColumnHandler {
	return &ColumnHandler{repo: repo, boards: boards, hub: hub}
}

func (h *ColumnHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateColumnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		jsonError(w, http.StatusBadRequest, "title is required")
		return
	}
	ok, err := h.boards.IsMember(r.Context(), req.BoardID, middleware.GetUserID(r.Context()))
	if err != nil || !ok {
		jsonError(w, http.StatusNotFound, "board not found")
		return
	}
	col := &models.Column{BoardID: req.BoardID, Title: req.Title}
	if err := h.repo.Create(r.Context(), col); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.hub.Broadcast(col.BoardID, "column.created", col)
	jsonResponse(w, http.StatusCreated, col)
}

func (h *ColumnHandler) GetByBoard(w http.ResponseWriter, r *http.Request) {
	boardID := mux.Vars(r)["board_id"]
	ok, accessErr := h.boards.IsMember(r.Context(), boardID, middleware.GetUserID(r.Context()))
	if accessErr != nil || !ok {
		jsonError(w, http.StatusNotFound, "board not found")
		return
	}
	cols, err := h.repo.GetByBoardID(r.Context(), boardID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if cols == nil {
		cols = []models.Column{}
	}
	jsonResponse(w, http.StatusOK, cols)
}
