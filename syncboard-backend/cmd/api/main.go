package main

import (
	"log"
	"net/http"

	"syncboard/internal/config"
	"syncboard/internal/handlers"
	"syncboard/internal/middleware"
	"syncboard/internal/repository"
	"syncboard/internal/service"

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

	// Services
	authService := service.NewAuthService(userRepo, cfg)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)
	boardHandler := handlers.NewBoardHandler(boardRepo)
	columnHandler := handlers.NewColumnHandler(columnRepo)
	cardHandler := handlers.NewCardHandler(cardRepo)

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
	api.HandleFunc("/columns", columnHandler.Create).Methods("POST")

	// Cards
	api.HandleFunc("/cards", cardHandler.Create).Methods("POST")
	api.HandleFunc("/cards/{id}/move", cardHandler.Move).Methods("PATCH")
	api.HandleFunc("/cards/{id}", cardHandler.Delete).Methods("DELETE")

	// CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)

	log.Printf("🚀 Server starting on http://localhost:%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}
