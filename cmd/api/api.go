package main

import (
	"log"
	"net/http"
	"time"

	"github.com/elhambadri2411/social/internal/store" // internal package, serves as abstraction layer for db
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// `application` config struct which represents application context
type application struct {
	config config        // app level config settings
	store  store.Storage // "Repository"
}

// `config` struct stores application configuration, including:
type config struct {
	addr string   // port
	db   dbConfig // db config settings
	env  string   // env (PROD or DEV)
}

// `dbConfig` struct hold db related config
type dbConfig struct {
	url          string // db connection string
	maxOpenConns int    // max number of open connections
	maxIdleConns int    // max number of idle connections
	maxIdleTime  string // max idle time before connections close
}

// `mount` initializes and configures the HTTP router using `chi`.
// It sets up routes, applies middleware, and groups API endpoints.
func (app *application) mount() *chi.Mux {
	r := chi.NewRouter() // create new Chi router instance

	r.Use(middleware.RequestID) // Assign request Id to each req
	r.Use(middleware.RealIP)    // Extract client IP from header
	r.Use(middleware.Logger)    // Log incoming requests
	r.Use(middleware.Recoverer) // recoverse from panics and prevents crashes (throw 500)

	// Set a timeout for all HTTP requests to prevent hanging requests.
	// If a request takes longer than 60 seconds, it is automatically canceled.
	r.Use(middleware.Timeout(60 * time.Second))

	// Define API routes under the `/v1` prefix
	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)

		r.Route("/posts", func(r chi.Router) {
			r.Post("/", app.createPostsHandler)
			r.Route("/{postId}", func(r chi.Router) {
				r.Get("/", app.getPostByIdHandler)
			})
		})
	})

	return r
}

// starts sttp server (listens) with given router
func (app *application) run(mux *chi.Mux) error {
	// create http server with app.config options
	server := &http.Server{
		Addr:         app.config.addr,  // Set server address from config
		Handler:      mux,              // Use the `mux` router as the request handler
		WriteTimeout: time.Second * 30, // Max time for writing a response (30s)
		ReadTimeout:  time.Second * 10, // Max time for reading a request (10s)
		IdleTimeout:  time.Minute,      // Max idle time for keep-alive connections (1 min)
	}

	log.Printf("Server listening on %s\n", app.config.addr)
	return server.ListenAndServe()
}
