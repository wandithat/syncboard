package handlers

import (
	"encoding/json"
	"net/http"
	"syncboard/internal/middleware"
	"syncboard/internal/models"
	"syncboard/internal/repository"

	"github.com/gorilla/mux"
)

type BoardHandler struct {
	repo *repository.BoardRepository
}

func NewBoardHandler(repo *repository.BoardRepository) *BoardHandler {
	return &BoardHandler{repo: repo}
}

func (h *BoardHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req models.CreateBoardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Title == "" {
		jsonError(w, http.StatusBadRequest, "title is required")
		return
	}

	board := &models.Board{Title: req.Title, OwnerID: userID}
	if err := h.repo.Create(r.Context(), board); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, board)
}

func (h *BoardHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	boards, err := h.repo.GetUserBoards(r.Context(), userID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if boards == nil {
		boards = []models.Board{}
	}

	jsonResponse(w, http.StatusOK, boards)
}

func (h *BoardHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	board, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		jsonError(w, http.StatusNotFound, "board not found")
		return
	}
	jsonResponse(w, http.StatusOK, board)
}
