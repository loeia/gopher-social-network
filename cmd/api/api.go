package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/loeia/gopherSocialNetwork/internal/auth"
	"github.com/loeia/gopherSocialNetwork/internal/mailer"
	"github.com/loeia/gopherSocialNetwork/internal/store"
	"go.uber.org/zap"
)

type application struct {
	config
	store         *store.Storage
	logger        *zap.SugaredLogger
	mailer        mailer.Client
	authenticator auth.Authenticator
}

type config struct {
	addr string
	env  string
	db   dbConfig
	mail mailConfig
	auth authConfig

	frontendURL string
}

type mailConfig struct {
	fromEmail string
	mailTrap  mailTrapConfig
	exp       time.Duration
}

type mailTrapConfig struct {
	apiKey      string
	sandboxUser string
	sandboxPass string
}

type authConfig struct {
	token tokenConfig
}
type tokenConfig struct {
	secret string
	exp    time.Duration
	iss    string
}
type dbConfig struct {
	dsn          string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
	maxLifeTime  string
}

// 注册中间件和路由
func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	// Basic CORS
	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(time.Second * 60))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("all is well"))
	})

	r.Route("/posts", func(r chi.Router) {
		r.Use(app.AuthTokenMiddleware)
		r.Post("/", app.createPostHandler)

		r.Route("/{postId}", func(r chi.Router) {
			r.Use(app.postsContextMiddleware)

			r.Get("/", app.getPostHandler)
			r.Delete("/", app.deletePostHandler)
			r.Patch("/", app.updatePostHandler)
		})
	})

	r.Route("/users", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)
			r.Get("/feed", app.getUserFeedHandler)
		})

		r.Put("/activate/{token}", app.activateUserHandler)

		r.Route("/{userId}", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)

			r.Get("/", app.getUserHandler)
			r.Put("/follow", app.followUserHandler)
			r.Put("/unfollow", app.unfollowUserHandler)
		})
	})

	// public routes
	r.Route("/authentication", func(r chi.Router) {
		r.Post("/users", app.registerUserHandler)
		r.Post("/token", app.createTokenHandler)
	})

	return r
}

func (app *application) run(mux http.Handler) error {
	server := &http.Server{
		Addr:         app.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 90,
		ReadTimeout:  time.Second * 90,
		IdleTimeout:  time.Second * 60,
	}

	app.logger.Infow("server has started", "addr", app.config.addr)
	return server.ListenAndServe()
}
