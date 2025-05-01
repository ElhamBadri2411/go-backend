package main

import (
	"log"

	"github.com/elhambadri2411/social/internal/db"
	"github.com/elhambadri2411/social/internal/env"
	"github.com/elhambadri2411/social/internal/store"
)

func main() {
	addr := env.GetString("DB_URL", "postgres://user:password@localhost/social?sslmode=disable")
	conn, err := db.New(addr, 5, 5, "10m")
	if err != nil {
		log.Fatal(err)
	}
	store := store.NewStorage(conn)
	db.Seed(store)
}
