# Gopher Social Network

A simple social network platform API built with Go, featuring JWT authentication, user registration with email activation, post CRUD, follow/unfollow, a paginated feed, and random public posts.

Built as a learning exercise by following the excellent [GopherSocial](https://github.com/sikozonpc/GopherSocial) project.

## Tech Stack

**Backend**
- Go 1.26 with `chi/v5` router
- PostgreSQL 16.3 (via `pgx` and `database/sql`)
- Redis 7.2 (optional user caching)
- JWT (HS256) for authentication
- bcrypt for password hashing
- `go-playground/validator` for request validation
- `uber-go/zap` for structured logging
- `gomail` for email sending via MailTrap
- In-memory fixed-window rate limiter

**Frontend**
- React 19 with TypeScript
- Vite 8 dev server and bundler
- `react-router-dom` v7

**Infrastructure**
- Docker Compose (PostgreSQL + Redis)
- Air for live-reload during development

## Routes

### Health
`GET /health`

### Authentication (public)
`POST /authentication/users` — register a new user\
`POST /authentication/token` — obtain a JWT token

### Posts (authenticated)
`POST /posts` — create a post\
`GET /posts/{postId}` — retrieve a post\
`PATCH /posts/{postId}` — update a post (owner or moderator)\
`DELETE /posts/{postId}` — delete a post (owner or admin)

### Users (authenticated)
`GET /users/feed` — paginated feed from followed users\
`GET /users/{userId}` — get a user's profile\
`PUT /users/{userId}/follow` — follow a user\
`PUT /users/{userId}/unfollow` — unfollow a user

### User Activation
`PUT /users/activate/{token}` — activate account via email token

### Public
`GET /free` — random public posts (no auth required)

## Environment Variables

Most variables have sensible defaults, but a few must be set before running.

**Required**
- `AUTH_SECRET` — JWT signing secret (any random string)
- `MAILTRAP_API_KEY` — MailTrap API key for sending emails
- `MAILTRAP_SANDBOX_USER` — MailTrap SMTP username
- `MAILTRAP_SANDBOX_PASS` — MailTrap SMTP password
- `FROM_EMAIL` — sender email address for activation emails

**Database**
- `DB_DSN` — PostgreSQL connection string (default `postgres://admin:admin123@localhost/gopher-social-network?sslmode=disable`)
- `DB_MAX_OPEN_CONNS` — max open connections (default `30`)
- `DB_MAX_IDLE_CONNS` — max idle connections (default `30`)
- `DB_MAX_IDLE_TIME` — max idle connection time (default `15m`)
- `DB_MAX_LIFE_TIME` — max connection lifetime (default `5m`)

**Server**
- `ADDR` — listen address (default `:8080`)
- `ENV` — runtime environment (default `development`)
- `FRONTEND_URL` — frontend origin for CORS (default `http://localhost:5173`)

**Auth**
- `AUTH_EXP` — JWT token expiry in hours (default `72`)
- `AUTH_ISSUER` — JWT issuer claim (default `Bearer`)

**Redis**
- `REDIS_ADDR` — Redis address (default `localhost:6379`)
- `REDIS_PASSWORD` — Redis password (default empty)
- `REDIS_DB` — Redis database number (default `0`)
- `REDIS_ENABLED` — enable Redis user caching (default `false`)

**Rate Limiter**
- `RATELIMITER_REQUESTS_COUNT` — max requests per time window (default `20`)
- `RATE_LIMITER_ENABLED` — enable rate limiting (default `true`)

You can set these in a `.envrc` file (if using direnv) or export them directly in your shell.

## Getting Started

```bash
# Start PostgreSQL and Redis
docker compose up -d

# Run database migrations and seed data
make db

# Start the API server (with Air live-reload)
make run
```

The API server starts at `http://localhost:8080`. The frontend dev server runs at `http://localhost:5173`.

## Acknowledgments

Huge thanks to [sikozonpc](https://github.com/sikozonpc) for the original [GopherSocial](https://github.com/sikozonpc/GopherSocial) project. This codebase was written as a hands-on learning exercise by closely following that project.
