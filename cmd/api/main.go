package main

import (
	"log"
)

func main() {
	var configSet config = config{
		addr: ":8080",
	}

	var app application = application{
		config: configSet,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
