package main

import (
	"github.com/elhambadri2411/social/internal/env"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("WARNING: Env failed to load")
	}
	var configSet config = config{
		addr: env.GetString("ADDR", ":8080"),
	}

	var app application = application{
		config: configSet,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
