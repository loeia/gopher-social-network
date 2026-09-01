package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/loeia/gopherSocialNetwork/internal/store"
)

// AuthTokenMiddleware validates the JWT token and sets the authenticated user in the request context.
func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("authorization header is missing"))
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != app.config.auth.token.iss {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("authorization header is malformed"))
			return
		}

		token := parts[1]
		jwtToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		claims, _ := jwtToken.Claims.(jwt.MapClaims)

		userId, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		ctx := r.Context()
		user, err := app.getUser(ctx, userId)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		if ver, ok := claims["ver"].(float64); !ok || int(ver) != user.TokenVer {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("token has been revoked"))
			return
		}

		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getUserFromCtx retrieves the authenticated user from the request context.
func getUserFromCtx(r *http.Request) *store.User {
	return r.Context().Value(userCtx).(*store.User)
}

// postsContextMiddleware loads the post by ID from the URL and sets it in the request context.
func (app *application) postsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postIdStr := chi.URLParam(r, "postId")
		postId, err := strconv.ParseInt(postIdStr, 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		ctx := r.Context()

		post, err := app.getPost(ctx, postId)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundError(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getPostFromCtx retrieves the post from the request context.
func getPostFromCtx(r *http.Request) *store.Post {
	return r.Context().Value(postCtx).(*store.Post)
}

// checkPostOwnerShip verifies the user owns the post or has the required role to proceed.
func (app *application) checkPostOwnerShip(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromCtx(r)
		post := getPostFromCtx(r)

		// if it is the user post
		if user.ID == post.UserID {
			next.ServeHTTP(w, r)
			return
		}

		// role precedence check
		allowed, err := app.checkRolePrecedence(r.Context(), user, requiredRole)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		if !allowed {
			app.forbiddenResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// checkRolePrecedence checks if the user's role level meets the required role level.
func (app *application) checkRolePrecedence(c context.Context, user *store.User, roleName string) (bool, error) {
	role, err := app.store.Roles.GetByName(c, roleName)
	if err != nil {
		return false, err
	}

	return user.Role.Level >= role.Level, nil
}

// getUser retrieves a user by ID, using cache when enabled.
func (app *application) getUser(c context.Context, userId int64) (*store.User, error) {
	if !app.config.redisCfg.enabled {
		return app.store.Users.GetById(c, userId)
	}

	user, err := app.cache.User.Get(c, userId)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user, err = app.store.Users.GetById(c, userId)
		if err != nil {
			return nil, err
		}
		if err := app.cache.User.Set(c, user); err != nil {
			return nil, err
		}
	}

	return user, nil
}

// rateLimiterMiddleware enforces rate limiting based on the client's IP address.
func (app *application) rateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.rateLimiter.Enabled {
			if allow, retryAfter := app.rateLimiter.Allow(r.RemoteAddr); !allow {
				app.rateLimitExceededResponse(w, r, retryAfter.String())
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// commentsContextMiddleware loads the comment by ID from the URL and sets it in the request context.
func (app *application) commentsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		commentId, err := strconv.ParseInt(chi.URLParam(r, "commentId"), 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
		ctx := r.Context()

		comment, err := app.store.Comments.GetById(ctx, commentId)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundError(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, CommentCtx, comment)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getCommentFromCtx retrieves the comment from the request context.
func getCommentFromCtx(r *http.Request) *store.Comment {
	return r.Context().Value(CommentCtx).(*store.Comment)
}

// checkCommentOwnerShip verifies the user owns the comment or has the required role to proceed.
func (app *application) checkCommentOwnerShip(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromCtx(r)
		comment := getCommentFromCtx(r)

		if user.ID == comment.UserID {
			next.ServeHTTP(w, r)
			return
		}

		// role precedence check
		allowed, err := app.checkRolePrecedence(r.Context(), user, requiredRole)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		if !allowed {
			app.forbiddenResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// getPost retrieves a post by ID, using cache when enabled.
func (app *application) getPost(c context.Context, postId int64) (*store.Post, error) {
	if !app.config.redisCfg.enabled {
		return app.store.Posts.GetById(c, postId)
	}

	post, err := app.cache.Post.Get(c, postId)
	if err != nil {
		return nil, err
	}

	if post == nil {
		post, err = app.store.Posts.GetById(c, postId)
		if err != nil {
			return nil, err
		}
		if err := app.cache.Post.Set(c, post); err != nil {
			return nil, err
		}
	}

	return post, nil
}

func (app *application) verifyAdminPermMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromCtx(r)

		allowed, err := app.checkRolePrecedence(r.Context(), user, "admin")
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		if !allowed {
			app.forbiddenResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
