package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/loeia/gopherSocialNetwork/internal/store"
	"go.uber.org/zap"
)

type application struct {
	config
	store *store.Storage
	logger *zap.SugaredLogger
}

type config struct {
	addr string
	db   dbConfig
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

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(time.Second * 60))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("all is well"))
	})

	r.Route("/posts", func(r chi.Router) {
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
			r.Get("/feed", app.getUserFeedHandler)
		})

		r.Route("/{userId}", func(r chi.Router) {
			r.Use(app.usersContextMiddleware)

			r.Get("/", app.getUserHandler)

			r.Put("/follow", app.followUserHandler)
			r.Put("/unfollow", app.unfollowUserHandler)
		})
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

	app.logger.Infow("server has started","addr",app.config.addr)
	return server.ListenAndServe()
}
