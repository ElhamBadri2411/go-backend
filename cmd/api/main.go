package main

import (
	"log"

	"github.com/elhambadri2411/social/internal/db"
	"github.com/elhambadri2411/social/internal/env"
	"github.com/elhambadri2411/social/internal/store"
	"github.com/joho/godotenv"
)

func main() {
	// Loads env

	err := godotenv.Load()
	if err != nil {
		log.Fatal("WARNING: Env failed to load")
	}

	config := config{
		addr: env.GetString("ADDR", ":8080"),
		db: dbConfig{
			url:          env.GetString("DB_URL", "postgres://user:password@localhost/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
	}

	db, err := db.New(config.db.url, config.db.maxOpenConns, config.db.maxIdleConns, config.db.maxIdleTime)
	store := store.NewStorage(db)

	var app application = application{
		config: config,
		store:  store,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
