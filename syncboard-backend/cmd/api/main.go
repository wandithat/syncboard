package main

import (
	"log"
	"net/http"

	"syncboard/internal/config"
	"syncboard/internal/handlers"
	"syncboard/internal/middleware"
	"syncboard/internal/realtime"
	"syncboard/internal/repository"
	"syncboard/internal/service"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Config error:", err)
	}

	db, err := repository.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("DB error:", err)
	}
	defer db.Close()

	// Repositories
	userRepo := repository.NewUserRepository(db)
	boardRepo := repository.NewBoardRepository(db)
	columnRepo := repository.NewColumnRepository(db)
	cardRepo := repository.NewCardRepository(db)
	hub := realtime.NewHub()

	// Services
	authService := service.NewAuthService(userRepo, cfg)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)
	boardHandler := handlers.NewBoardHandler(boardRepo)
	columnHandler := handlers.NewColumnHandler(columnRepo, boardRepo, hub)
	cardHandler := handlers.NewCardHandler(cardRepo, columnRepo, boardRepo, hub)
	invitationHandler := handlers.NewInvitationHandler(boardRepo, userRepo, hub)

	// Router
	r := mux.NewRouter()

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET")

	// Public
	r.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")
	r.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")

	// Protected
	api := r.PathPrefix("/api").Subrouter()
	api.Use(func(next http.Handler) http.Handler {
		return middleware.JWTAuth(cfg, next)
	})

	// Boards
	api.HandleFunc("/boards", boardHandler.Create).Methods("POST")
	api.HandleFunc("/boards", boardHandler.List).Methods("GET")
	api.HandleFunc("/boards/{id}", boardHandler.Get).Methods("GET")

	// Columns
	api.HandleFunc("/boards/{board_id}/columns", columnHandler.GetByBoard).Methods("GET")
	api.HandleFunc("/boards/{board_id}/columns/reorder", columnHandler.Reorder).Methods("PATCH")
	api.HandleFunc("/columns", columnHandler.Create).Methods("POST")

	// Invitations and board members
	api.HandleFunc("/boards/{board_id}/invitations", invitationHandler.Create).Methods("POST")
	api.HandleFunc("/boards/{board_id}/members", invitationHandler.ListMembers).Methods("GET")
	api.HandleFunc("/invitations", invitationHandler.ListMine).Methods("GET")
	api.HandleFunc("/invitations/{id}/accept", invitationHandler.Accept).Methods("POST")
	api.HandleFunc("/invitations/{id}/decline", invitationHandler.Decline).Methods("POST")

	// Cards
	api.HandleFunc("/columns/{column_id}/cards", cardHandler.GetByColumn).Methods("GET")
	api.HandleFunc("/cards", cardHandler.Create).Methods("POST")
	api.HandleFunc("/cards/{id}/move", cardHandler.Move).Methods("PATCH")
	api.HandleFunc("/cards/{id}", cardHandler.Delete).Methods("DELETE")

	// WebSocket: JWT передаётся query-параметром, так как браузерный WebSocket API
	// не позволяет задать произвольный Authorization header.
	r.HandleFunc("/ws/boards/{id}", func(w http.ResponseWriter, req *http.Request) {
		token := req.URL.Query().Get("token")
		userID, err := middleware.ParseToken(cfg, token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		boardID := mux.Vars(req)["id"]
		ok, err := boardRepo.IsMember(req.Context(), boardID, userID)
		if err != nil || !ok {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		hub.ServeBoard(w, req, boardID)
	}).Methods("GET")

	// CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		ExposedHeaders:   []string{"Content-Length"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)

	log.Printf("🚀 Server starting on http://localhost:%s", cfg.Port)
	server := &http.Server{Addr: ":" + cfg.Port, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	log.Fatal(server.ListenAndServe())
}
