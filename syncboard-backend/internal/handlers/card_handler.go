package handlers

import (
	"encoding/json"
	"net/http"
	"syncboard/internal/models"
	"syncboard/internal/repository"

	"github.com/gorilla/mux"
)

type CardHandler struct {
	repo *repository.CardRepository
}

func NewCardHandler(repo *repository.CardRepository) *CardHandler {
	return &CardHandler{repo: repo}
}

func (h *CardHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	card := &models.Card{ColumnID: req.ColumnID, Title: req.Title}
	if err := h.repo.Create(r.Context(), card); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusCreated, card)
}

func (h *CardHandler) Move(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var req models.MoveCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if err := h.repo.Move(r.Context(), id, req.ColumnID, req.Position); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"status": "moved"})
}

func (h *CardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.repo.Delete(r.Context(), id); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
