package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/loeia/gopherSocialNetwork/internal/auth"
	"github.com/loeia/gopherSocialNetwork/internal/env"
	"github.com/loeia/gopherSocialNetwork/internal/mailer"
	"github.com/loeia/gopherSocialNetwork/internal/ratelimiter"
	"github.com/loeia/gopherSocialNetwork/internal/store"
	"github.com/loeia/gopherSocialNetwork/internal/store/cache"
	"go.uber.org/zap"
)

type application struct {
	config
	store         *store.Storage
	cache         *cache.Storage
	logger        *zap.SugaredLogger
	mailer        mailer.Client
	authenticator auth.Authenticator
	rateLimiter   ratelimiter.Limiter
}

type config struct {
	addr        string
	env         string
	frontendURL string
	db          dbConfig
	mail        mailConfig
	auth        authConfig
	redisCfg    redisConfig
	rateLimiter ratelimiter.Config
}

type mailConfig struct {
	fromEmail string
	mailTrap  mailTrapConfig
	exp       time.Duration
	resetExp  time.Duration
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
type redisConfig struct {
	addr     string
	password string
	db       int
	enabled  bool
}

// mount configures and returns the HTTP router with all routes and middleware.
func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	// Basic CORS
	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{env.GetString("CORS_ALLOWED_ORIGIN", "http://localhost:5173")},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(time.Second * 60))

	if app.config.rateLimiter.Enabled {
		r.Use(app.rateLimiterMiddleware)
	}

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("all is well"))
	})

	r.Route("/posts", func(r chi.Router) {
		r.Route("/{postId}", func(r chi.Router) {
			r.Use(app.postsContextMiddleware)
			r.Get("/", app.getPostHandler)
			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Patch("/", app.checkPostOwnerShip("moderator", app.updatePostHandler))
				r.Delete("/", app.checkPostOwnerShip("admin", app.deletePostHandler))
				r.Put("/like", app.likePostHandler)
				r.Put("/dislike", app.unlikePostHandler)
			})
			r.Route("/comments", func(r chi.Router) {
				r.Get("/", app.getPostCommentsHandler)
				r.Group(func(r chi.Router) {
					r.Use(app.AuthTokenMiddleware)
					r.Post("/", app.createCommentHandler)
					r.Post("/{commentId}/reply", app.createCommentHandler)
				})
			})
		})
		r.Group(func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)
			r.Post("/", app.createPostHandler)
		})
		r.Get("/search", app.getSearchPostHandler)
		r.Get("/free", app.getFreePostsHandler)
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/forgot-password", app.forgetPassHandler)
		r.Post("/reset-password", app.resetPasswordFromTokenHandler)
		r.Group(func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)
			r.Get("/feed", app.getUserFeedHandler)
			r.Patch("/reset", app.resetPasswordHandler)
			r.Put("/me/avatar", app.uploadAvatarHandler)
			r.Put("/me/profile", app.updateProfileHandler)
			r.Get("/comment-likes", app.getUserCommentLikesHandler)
			r.Patch("/rename", app.userRenameHandler)
		})
		r.Put("/activate/{token}", app.activateUserHandler)
		r.Route("/{userId}", func(r chi.Router) {
			r.Get("/", app.getUserHandler)
			r.Get("/posts", app.getUserPostsHandler)
			r.Get("/post-likes", app.getUserLikedPostsHandler)
			r.Get("/comments", app.getUserCommentsHandler)
			r.Get("/followers", app.getUserFollowersHandler)
			r.Get("/following", app.getUserFollowingHandler)
			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})
		})
		r.Get("/{userId}/avatar", app.getAvatarHandler)
	})

	r.Route("/comments", func(r chi.Router) {
		r.Route("/{commentId}", func(r chi.Router) {
			r.Use(app.commentsContextMiddleware)
			r.Get("/", app.getCommentHandler)
			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Delete("/", app.checkCommentOwnerShip("moderator", app.deleteCommentHandler))
				r.Put("/like", app.likeCommentHandler)
				r.Put("/dislike", app.dislikeCommentHandler)
			})
		})
	})

	r.Route("/authentication", func(r chi.Router) {
		r.Post("/users", app.registerUserHandler)
		r.Post("/token", app.createTokenHandler)
	})

	return r
}

// run starts the HTTP server and handles graceful shutdown.
func (app *application) run(mux http.Handler) error {
	server := &http.Server{
		Addr:         app.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 90,
		ReadTimeout:  time.Second * 90,
		IdleTimeout:  time.Second * 60,
	}

	shutdown := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		app.logger.Infow("singal caught", "signal", s.String())
		shutdown <- server.Shutdown(ctx)
	}()

	app.logger.Infow("server has started", "addr", app.config.addr)

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	if err := <-shutdown; err != nil {
		return err
	}

	return nil
}
