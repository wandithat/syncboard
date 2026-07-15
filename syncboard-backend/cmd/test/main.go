package main

import (
	"fmt"
	"log"
	"syncboard/internal/config"
	"syncboard/internal/repository"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := repository.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}
	defer db.Close()

	fmt.Println("✅ Database connected successfully!")
}
