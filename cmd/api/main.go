package main

import (
	"time"

	"github.com/elhambadri2411/social/internal/auth"
	"github.com/elhambadri2411/social/internal/db"  // internal package for handling db connections
	"github.com/elhambadri2411/social/internal/env" // internal package for extracting and loading env variables
	"github.com/elhambadri2411/social/internal/mailer"
	"github.com/elhambadri2411/social/internal/ratelimiter"
	"github.com/elhambadri2411/social/internal/store" // internal package, serves as abstraction layer for db
	"github.com/elhambadri2411/social/internal/store/cache"
	"github.com/joho/godotenv" // package for loading environment variables
	"go.uber.org/zap"
)

const version = "0.0.1"

//	@title			GolangSocial
//	@description	API for Social media for developers
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

// @license.name				Apache 2.0
// @license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description
func main() {
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()
	// Loads env from a .env file
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("WARNING: Env failed to load")
	}

	/*
		Create a configuration object to store important application settings.

		- `addr`: The address (port) where the application server will run.
		- `env`: The environment mode (e.g., "DEV" for development, "PROD" for production).
		- `db`: A nested struct (`dbConfig`) holding database connection settings:
			- `url`: The database connection string. (which has info like user, password, dbname)
			- `maxOpenConns`: The maximum number of open database connections.
			- `maxIdleConns`: The maximum number of idle database connections.
			- `maxIdleTime`: The maximum amount of time a connection can stay idle before being closed.

		Each configuration value is fetched using the `env` package, which retrieves environment
		variables, providing default values when the variables are not set.
	*/
	config := config{
		addr:        env.GetString("ADDR", ":8080"),
		env:         env.GetString("ENV", "DEV"),
		apiUrl:      env.GetString("EXTERNAL_URL", "localhost:3000"),
		frontendUrl: env.GetString("FRONTEND_URL", "localhost:3001"),
		db: dbConfig{
			url:          env.GetString("DB_URL", "postgres://user:password@localhost/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		redis: redisConfig{
			address:   env.GetString("REDIS_ADDR", "localhost:6379"),
			password:  env.GetString("REDIS_PASS", ""),
			db:        env.GetInt("REDIS_DB", 0),
			isEnabled: env.GetBoolean("REDIS_IS_ENABLED", true),
		},
		mail: mailConfig{
			exp:       time.Hour * 8,
			fromEmail: env.GetString("FROM_EMAIL", "test@mail.com"),
			sendGrid: sendGridConfig{
				apiKey: env.GetString("SENDGRID_API_KEY", ""),
			},
		},

		auth: authConfig{
			basic: basicConfig{
				username: env.GetString("BASIC_USERNAME", "user"),
				password: env.GetString("BASIC_PASSWORD", "password"),
			},
			token: tokenConfig{
				secret:     env.GetString("SECRET", "secretwaffle123"),
				expiration: time.Hour * 36,
				issuer:     "devsocial",
			},
		},
		rateLimiter: ratelimiter.Config{
			RequestsPerTimeFrame: env.GetInt("RATELIMITER_REQUESTS", 20),
			TimeFrame:            time.Second * 30,
		},
	}

	// Init a new db connections with configuration setup
	// db.New returns a database instance
	db, err := db.New(config.db.url, config.db.maxOpenConns, config.db.maxIdleConns, config.db.maxIdleTime)
	if err != nil {
		logger.Fatal(err)
		db.Close()
	}
	defer db.Close()

	// Initialize a new `cache` which is the interface for a redis cache
	redis := cache.NewRedisClient(config.redis.address, config.redis.password, config.redis.db)

	limiter := ratelimiter.NewFixedWindowRateLimiter(config.rateLimiter.RequestsPerTimeFrame, config.rateLimiter.TimeFrame)

	// Initialize a new storage layer (`store`) which acts as an interface between
	// the application and the database. The `store` package is responsible
	// for querying, inserting, updating, and deleting records in the database.
	store := store.NewStorage(db)

	// Initialize a new `mailer` which is the interface for sending emails
	mailer := mailer.NewSendgrid(config.mail.sendGrid.apiKey, config.mail.fromEmail)

	// Initialize a new `jwtAuthenticator` an interface for generating and validaitng jwt tokens
	jwtAuthenticator := auth.NewJWTAuthenticator(config.auth.token.secret, config.auth.token.issuer, config.auth.token.issuer)

	// Create an `application` instance which encapsulates configuration settings
	// and storage, making them accessible throughout the application.
	app := application{
		config:        config,
		store:         store,
		logger:        logger,
		mailer:        mailer,
		authenticator: jwtAuthenticator,
		cache:         cache.NewCacheStorage(redis),
		rateLimiter:   limiter,
	}

	// Mount the application's HTTP handlers (routes) onto a multiplexer (`mux`).
	// This function likely sets up all API endpoints, middlewares, and routing logic.
	mux := app.mount()

	// Start the application server and log any errors that occur.
	// `app.run(mux)` is expected to start an HTTP server and listen for requests.
	logger.Info(app.run(mux))
}
