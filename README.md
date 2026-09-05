English |[中文](README.zh.md)

# Gopher Social Network

![Go Version](https://img.shields.io/badge/Go-1.27-00ADD8?style=flat-square&logo=go)
![Vue Version](https://img.shields.io/badge/Vue-3.5-4FC08D?style=flat-square&logo=vuedotjs)
![Docker](https://img.shields.io/badge/Docker-24.0-2496ED?style=flat-square&logo=docker)
![Redis](https://img.shields.io/badge/Redis-7.2-DC382D?style=flat-square&logo=redis)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16.3-336791?style=flat-square&logo=postgresql)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)

[📚 Project Documentation](https://loeia.github.io/gopher-social-network/)

A social network application built with Go backend and Vue 3 frontend, supporting user authentication, post publishing, comment interactions, follow system, and more.

## Features

- **User System**: Registration, login, email activation, password reset
- **Post Management**: Create, edit, delete posts with tags and search support
- **Comment System**: Multi-level comments and replies with like support
- **Follow System**: Follow/unfollow users, view activity feed
- **Avatar Upload**: Image upload and cropping support
- **Real-time Notifications**: Email notifications and password reset
- **Performance Optimization**: Redis caching and rate limiting

## Tech Stack

### Backend
- **Language**: Go 1.27
- **Router**: Chi Router
- **Database**: PostgreSQL 16
- **Cache**: Redis 7
- **Authentication**: JWT (JSON Web Token)
- **Email**: MailTrap
- **Logging**: Zap
- **Hot Reload**: Air

### Frontend
- **Framework**: Vue 3.5
- **Language**: TypeScript
- **Build Tool**: Vite 8
- **UI Library**: Element Plus
- **State Management**: Pinia
- **Routing**: Vue Router 4

## Quick Start

### Prerequisites

- Go 1.27+
- Node.js 22+ or 24+
- PostgreSQL 16
- Redis 7
- Docker (optional, for containerized deployment)

### Installation Steps

1. **Clone the Repository**
   ```bash
   git clone https://github.com/loeia/gopher-social-network.git
   cd gopher-social-network
   ```

2. **Configure Environment Variables**
   ```bash
   cp .envrc.example .envrc  # Edit .envrc file with your configuration
   cp web/.env.example web/.env
   ```

3. **Start Database Services**
   ```bash
   # Start PostgreSQL and Redis using Docker Compose
   docker compose up -d
   ```

4. **Database Migration**
   ```bash
   # Install goose: go install github.com/pressly/goose/v3/cmd/goose@latest
   # See: https://github.com/pressly/goose
   goose up
   make seed
   ```

5. **Start Backend Service**
   ```bash
   make run
   ```

6. **Start Frontend Service**
   ```bash
   cd web
   npm install
   npm run dev
   ```

### Docker Deployment

```bash
# Start all services with one command
make up

# Stop all services
make down
```

## Environment Variables

### Backend Environment Variables (.envrc)

#### Server Configuration
- `ADDR`: Server listen address, default `:8080`
- `ENV`: Runtime environment, default `development`, optional `production`
- `FRONTEND_URL`: Frontend address, default `http://localhost:5173`

#### Database Configuration
- `DB_DSN`: PostgreSQL connection string, default `postgres://admin:admin123@localhost/gopher-social-network?sslmode=disable`
- `DB_MAX_OPEN_CONNS`: Maximum open connections, default `30`
- `DB_MAX_IDLE_CONNS`: Maximum idle connections, default `30`
- `DB_MAX_IDLE_TIME`: Idle connection timeout, default `15m`
- `DB_MAX_LIFE_TIME`: Maximum connection lifetime, default `5m`

#### Authentication Configuration
- `AUTH_SECRET`: JWT secret key (required)
- `AUTH_EXP`: JWT expiration time (hours), default `72`
- `AUTH_ISSUER`: JWT issuer, default `Bearer`

#### Email Configuration
- `FROM_EMAIL`: Sender email (required)
- `MAILTRAP_API_KEY`: MailTrap API key
- `MAILTRAP_SANDBOX_USER`: MailTrap sandbox username
- `MAILTRAP_SANDBOX_PASS`: MailTrap sandbox password

#### Redis Configuration
- `REDIS_ENABLED`: Enable Redis, default `false`
- `REDIS_ADDR`: Redis address, default `localhost:6379`
- `REDIS_PASSWORD`: Redis password
- `REDIS_DB`: Redis database number, default `0`

#### Rate Limiter Configuration
- `RATE_LIMITER_ENABLED`: Enable rate limiter, default `true`
- `RATELIMITER_REQUESTS_COUNT`: Request count limit within time window, default `20`
- `RATELIMITER_TIME_FRAME`: Time window (seconds), default `5`

#### CORS Configuration
- `CORS_ALLOWED_ORIGIN`: Allowed CORS origin, default `http://localhost:5173`

#### Goose Migration Configuration (Optional)
- `GOOSE_DRIVER`: Database driver, default `pgx`
- `GOOSE_DBSTRING`: Migration database connection string
- `GOOSE_MIGRATION_DIR`: Migration files directory, default `./cmd/migrate/migrations/`

### Frontend Environment Variables (web/.env)

- `VITE_API_URL`: Backend API address, default `http://localhost:8080`

## API Routes Documentation

### Health Check
- `GET /health`: Check if service is running

### Posts (/posts)
- `GET /posts/search`: Search posts (supports pagination and filtering)
- `GET /posts/free`: Get random posts list
- `POST /posts`: Create new post
- `GET /posts/{postId}`: Get post details (public)
- `PATCH /posts/{postId}`: Update post
- `DELETE /posts/{postId}`: Delete post
- `PUT /posts/{postId}/like`: Like post
- `DELETE /posts/{postId}/like`: Unlike post

### Post Comments (/posts/{postId}/comments)
- `GET /posts/{postId}/comments`: Get post comments list (supports pagination)
- `POST /posts/{postId}/comments`: Create comment
- `POST /posts/{postId}/comments/{commentId}/replies`: Reply to comment

### Users (/users)
- `GET /users/feed`: Get user activity feed
- `PATCH /users/me/password`: Reset password
- `PUT /users/me/avatar`: Upload avatar
- `DELETE /users/me/avatar`: Delete avatar
- `PATCH /users/me/profile`: Update profile
- `GET /users/me/comment-likes`: Get user's liked comments
- `PATCH /users/me/username`: Change username

### User Details (/users/{userId})
- `GET /users/{userId}`: Get user information (public)
- `GET /users/{userId}/posts`: Get user posts list
- `GET /users/{userId}/post-likes`: Get user's liked posts
- `GET /users/{userId}/comments`: Get user comments list
- `GET /users/{userId}/followers`: Get user followers list
- `GET /users/{userId}/following`: Get user following list
- `PUT /users/{userId}/follow`: Follow user
- `DELETE /users/{userId}/follow`: Unfollow user
- `GET /users/{userId}/avatar`: Get user avatar

### Comments (/comments)
- `GET /comments/{commentId}`: Get comment details (public)
- `GET /comments/{commentId}/replies`: Get comment replies list
- `DELETE /comments/{commentId}`: Delete comment
- `PUT /comments/{commentId}/like`: Like comment
- `DELETE /comments/{commentId}/like`: Unlike comment

### Authentication (/auth)
- `POST /auth/register`: User registration
- `POST /auth/token`: User login, get JWT token
- `POST /auth/activate/{token}`: Activate user account
- `POST /auth/forgot-password`: Send password reset email
- `POST /auth/reset-password`: Reset password via token

### Admin Functions (/admin)
- `PATCH /admin/users/{username}/ban`: Ban user
- `PATCH /admin/users/{username}/unban`: Unban user

## Project Structure

```
gopher-social-network/
├── cmd/                    # Application entry point
│   ├── api/               # API server
│   └── migrate/           # Database migration tool
├── internal/              # Internal packages
│   ├── auth/              # Authentication module
│   ├── avatar/            # Avatar generation
│   ├── db/                # Database connection
│   ├── env/               # Environment variable handling
│   ├── mailer/            # Email service
│   ├── ratelimiter/       # Rate limiting
│   └── store/             # Data storage layer
├── web/                   # Frontend application
│   ├── src/               # Source code
│   ├── public/            # Static assets
│   └── package.json       # Frontend dependencies
├── docker-compose.yaml    # Docker orchestration file
├── Makefile              # Build scripts
├── go.mod                # Go module file
└── .envrc                # Environment variables configuration
```

## Screenshots

### Login Page
![Login](screenshots/login.png)

### Home Page
![Home](screenshots/home.png)

### Post Detail
![Post Detail](screenshots/post-detail.png)

### Create Post
![Create Post](screenshots/create-post.png)

### Search
![Search](screenshots/search.png)

### Profile
![Profile](screenshots/profile.png)

### Subscription List
![Subscription](screenshots/subscription.png)

## Contributing Guide

Contributions are welcome! Please follow these steps:

1. Fork this repository
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Create a Pull Request

### Code Standards

- Backend follows Go official code standards
- Frontend uses ESLint and Prettier for code formatting
- Commit messages follow Conventional Commits specification

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details
