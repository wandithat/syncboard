package handlers

import (
	"encoding/json"
	"net/http"
	"syncboard/internal/models"
	"syncboard/internal/repository"

	"github.com/gorilla/mux"
)

type ColumnHandler struct {
	repo *repository.ColumnRepository
}

func NewColumnHandler(repo *repository.ColumnRepository) *ColumnHandler {
	return &ColumnHandler{repo: repo}
}

func (h *ColumnHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateColumnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	col := &models.Column{BoardID: req.BoardID, Title: req.Title}
	if err := h.repo.Create(r.Context(), col); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusCreated, col)
}

func (h *ColumnHandler) GetByBoard(w http.ResponseWriter, r *http.Request) {
	boardID := mux.Vars(r)["board_id"]
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
