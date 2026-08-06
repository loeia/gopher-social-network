package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand/v2"

	"github.com/loeia/gopherSocialNetwork/internal/store"
)

// 50
var usernames = []string{
	"Alice", "Peter", "John", "Emma", "David",
	"Tom", "Jerry", "Lily", "James", "Sophia",
	"Lucas", "Mia", "Noah", "Olivia", "Ethan",
	"Ava", "Liam", "Isabella", "Mason", "Amelia",
	"Logan", "Harper", "Elijah", "Evelyn", "Aiden",
	"Abigail", "Henry", "Ella", "Jackson", "Scarlett",
	"Sebastian", "Grace", "Jack", "Chloe", "Owen",
	"Victoria", "Caleb", "Aria", "Luke", "Zoey",
	"Wyatt", "Riley", "Jayden", "Layla", "Grayson",
	"Nora", "Levi", "Luna", "Isaac", "Hannah",
}

// 50
var titles = []string{
	"Go Basics", "Concurrency Intro", "HTTP Server", "Slices Guide", "Maps in Go",
	"Structs Deep Dive", "Pointers Explained", "Interfaces Concept", "Error Handling", "Goroutines 101",
	"Channels Intro", "Mutex Usage", "Context Package", "JSON Handling", "File IO",
	"Testing in Go", "Benchmarking", "Project Structure", "Modules Explained", "Dependency Management",
	"REST API Design", "Middleware Pattern", "Database CRUD", "SQL with Go", "GORM Basics",
	"Authentication Flow", "JWT Usage", "Logging Strategy", "Config Management", "Environment Variables",
	"Caching Basics", "Redis Usage", "WebSocket Intro", "Microservices Concept", "Docker Basics",
	"Kubernetes Intro", "CI/CD Pipeline", "Error Wrapping", "Package Design", "Code Optimization",
	"Memory Model", "GC in Go", "Race Conditions", "Profiling Tools", "Benchmark Tips",
	"Clean Code", "SOLID in Go", "Design Patterns", "System Design", "Production Tips",
}

// 50
var contents = []string{
	"Learn Go syntax and basic concepts.",
	"Understand goroutines and concurrency model.",
	"Build a simple HTTP server.",
	"Work with slices efficiently.",
	"Understand map usage in Go.",
	"Learn how structs organize data.",
	"Understand pointers and memory references.",
	"Learn interfaces and polymorphism.",
	"Handle errors properly in Go.",
	"Introduction to goroutines.",
	"Learn channel communication.",
	"Use mutex for safe concurrency.",
	"Understand context package usage.",
	"Parse and generate JSON data.",
	"Read and write files in Go.",
	"Write unit tests effectively.",
	"Benchmark your Go code.",
	"Organize Go project structure.",
	"Use Go modules for dependency.",
	"Manage dependencies properly.",
	"Design REST APIs cleanly.",
	"Use middleware pattern in web apps.",
	"Basic CRUD with databases.",
	"Integrate SQL with Go.",
	"Introduction to GORM.",
	"Implement authentication system.",
	"Use JWT for auth.",
	"Design logging systems.",
	"Manage configuration cleanly.",
	"Use environment variables.",
	"Implement caching strategies.",
	"Use Redis for caching.",
	"Learn WebSocket basics.",
	"Understand microservices.",
	"Intro to Docker.",
	"Learn Kubernetes basics.",
	"CI/CD pipeline concepts.",
	"Wrap and handle errors properly.",
	"Design clean packages.",
	"Optimize Go code performance.",
	"Understand Go memory model.",
	"Learn garbage collection.",
	"Avoid race conditions.",
	"Use profiling tools.",
	"Improve benchmark results.",
	"Write clean code.",
	"Apply SOLID principles.",
	"Common design patterns.",
	"System design fundamentals.",
	"Production best practices.",
}

// 20
var tags = []string{
	"go", "backend", "api", "http", "grpc",
	"database", "sql", "redis", "cache", "docker",
	"kubernetes", "microservice", "concurrency", "goroutine", "channel",
	"json", "auth", "jwt", "testing", "performance",
}

// 20
var comments = []string{
	"Great post!",
	"Very helpful, thanks!",
	"I learned a lot from this.",
	"Can you explain more?",
	"This is confusing.",
	"Nice example.",
	"Well written!",
	"Helpful explanation.",
	"Thanks for sharing.",
	"I have a similar issue.",
	"Works perfectly.",
	"Not sure I understand.",
	"This solved my problem.",
	"Could you add more details?",
	"Interesting approach.",
	"I disagree with this.",
	"Good job!",
	"This is exactly what I needed.",
	"Please update this.",
	"Awesome content!",
}

func Seed(store *store.Storage, db *sql.DB) {
	ctx := context.Background()

	users := generateUsers()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatalln(err)
	}

	for _, user := range users {
		if err := store.Users.Create(ctx, user, tx); err != nil {
			_ = tx.Rollback()
			log.Println("Error creating user: ", err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatalln(err)
	}

	posts := generatePosts(users)
	for _, post := range posts {
		if err := store.Posts.Create(ctx, post); err != nil {
			log.Println("Error creating post:", err)
			return
		}
	}

	comments := generateComments(50, users, posts)
	for _, ct := range comments {
		if _, err := store.Comments.Create(ctx, ct); err != nil {
			log.Println("Error creating comment: ", err)
			return
		}
	}

	log.Println("Seeding complete")
}

// return 200 users
func generateUsers() (users []*store.User) {
	for range 200 {
		user := &store.User{
			Username: fmt.Sprintf("%s_%d", usernames[rand.IntN(50)], rand.IntN(112300)),
			Email:    fmt.Sprintf("%s_%d@example%d.com", usernames[rand.IntN(50)], rand.IntN(10123230), rand.IntN(5123)),
			Role: store.Role{
				Name: "user",
			},
		}
		if err := user.Password.Set("password123"); err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}
	return
}

// return 200 posts
func generatePosts(users []*store.User) []*store.Post {
	posts := make([]*store.Post, 2000)
	for i := range posts {
		user := users[rand.IntN(len(users))]
		posts[i] = &store.Post{
			UserID:  user.ID,
			Title:   titles[rand.IntN(len(titles))],
			Content: contents[rand.IntN(len(contents))],
			Tags:    generateTags(rand.IntN(5)),
		}
	}
	return posts
}

func generateTags(num int) []string {
	ts := make([]string, num)

	for i := range num {
		ts[i] = tags[rand.IntN(len(tags))]
	}

	return ts
}

func generateComments(num int, users []*store.User, posts []*store.Post) []*store.Comment {
	cs := make([]*store.Comment, num)
	for i := range num {
		cs[i] = &store.Comment{
			UserID:  users[rand.IntN(len(users))].ID,
			PostID:  posts[rand.IntN(len(posts))].ID,
			Content: comments[rand.IntN(len(comments))],
		}
	}

	return cs
}
