package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/elhambadri2411/social/docs" // internal package, used for generating swagger docs
	"github.com/elhambadri2411/social/internal/auth"
	"github.com/elhambadri2411/social/internal/mailer"
	"github.com/elhambadri2411/social/internal/store" // internal package, serves as abstraction layer for db
	"github.com/elhambadri2411/social/internal/store/cache"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

// `application` config struct which represents application context
type application struct {
	config        config             // app level config settings
	store         store.Storage      // "Repository"
	logger        *zap.SugaredLogger // logger
	mailer        mailer.Client
	authenticator auth.Authenticator
	cache         cache.Storage
}

// `authConfig` struct stores applciation auth configuration
type authConfig struct {
	basic basicConfig // basic auth config options
	token tokenConfig // configuration for auth tokens
}

// `basicConfig` struct stores applciation basic style auth configuration
type basicConfig struct {
	username string // username for basic auth
	password string // password for basic auth
}

// `tokenConfig` struct stores config for token based auth
type tokenConfig struct {
	secret     string
	expiration time.Duration
	issuer     string
}

// `config` struct stores application configuration, including:
type config struct {
	addr        string     // port
	db          dbConfig   // db config settings
	env         string     // env (PROD or DEV)
	apiUrl      string     // the external url
	mail        mailConfig // mail config settings
	auth        authConfig // auth config settings
	frontendUrl string     // url for the frontend
	redis       redisConfig
}

type redisConfig struct {
	address   string
	password  string
	db        int
	isEnabled bool
}

// `dbConfig` struct hold db related config
type dbConfig struct {
	url          string // db connection string
	maxOpenConns int    // max number of open connections
	maxIdleConns int    // max number of idle connections
	maxIdleTime  string // max idle time before connections close
}

// `mailConfig hols mail related config`
type mailConfig struct {
	exp       time.Duration  // how long before invites expire
	fromEmail string         // the email from which the sends come from
	sendGrid  sendGridConfig // sendGrid config type
}

type sendGridConfig struct {
	apiKey string // sendgrid api key
}

// `mount` initializes and configures the HTTP router using `chi`.
// It sets up routes, applies middleware, and groups API endpoints.
func (app *application) mount() *chi.Mux {
	// Docs
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiUrl
	docs.SwaggerInfo.BasePath = "/v1"
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
		r.With(app.BasicAuthMiddleware()).Get("/health", app.healthCheckHandler)
		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)

		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		r.Route("/posts", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)

			r.Post("/", app.createPostsHandler)
			r.Route("/{postId}", func(r chi.Router) {
				r.Use(app.postsContextMiddleware)

				r.Get("/", app.getPostByIdHandler)
				r.Delete("/", app.checkPostOwnership("admin", app.deletePostByIdHandler))
				r.Patch("/", app.checkPostOwnership("moderator", app.updatePostByIdHandler))
			})
			// r.Get("/", app.getPostsHandler)
		})

		r.Route("/users", func(r chi.Router) {
			r.Route("/{userId}", func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)

				r.Get("/", app.getUserHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})
			r.Put("/activate/{token}", app.activateUserHandler)

			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Get("/feed", app.getUserFeedHandler)
			})
		})

		// Public
		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
			r.Post("/token", app.createTokenHandler)
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

	app.logger.Infow("Server listening", "addr", app.config.addr, "env", app.config.env)

	return server.ListenAndServe()
}
