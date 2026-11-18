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

func (h *ColumnHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	boardID := mux.Vars(r)["board_id"]
	userID := middleware.GetUserID(r.Context())
	ok, err := h.boards.IsMember(r.Context(), boardID, userID)
	if err != nil || !ok {
		jsonError(w, http.StatusNotFound, "board not found")
		return
	}

	var req models.ReorderColumnsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.ColumnIDs == nil {
		jsonError(w, http.StatusBadRequest, "column_ids is required")
		return
	}
	seen := make(map[string]struct{}, len(req.ColumnIDs))
	for _, id := range req.ColumnIDs {
		if id == "" {
			jsonError(w, http.StatusBadRequest, "column id cannot be empty")
			return
		}
		if _, exists := seen[id]; exists {
			jsonError(w, http.StatusBadRequest, "duplicate column id")
			return
		}
		seen[id] = struct{}{}
	}
	if err := h.repo.ReorderColumns(r.Context(), boardID, req.ColumnIDs); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.hub.Broadcast(boardID, "column.reordered", map[string]interface{}{"column_ids": req.ColumnIDs})
	jsonResponse(w, http.StatusOK, map[string]string{"status": "reordered"})
}
