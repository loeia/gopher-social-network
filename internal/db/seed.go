package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand/v2"

	"github.com/loeia/gopherSocialNetwork/internal/avatar"
	"golang.org/x/crypto/bcrypt"
)

type userRow struct {
	id       int64
	username string
}

type postRow struct {
	id     int64
	userID int64
}

type commentRow struct {
	id     int64
	userID int64
	postID int64
}

func Seed(_ interface{}, db *sql.DB) {
	ctx := context.Background()

	if err := avatar.Init(); err != nil {
		log.Fatalf("avatar init: %v", err)
	}

	pwHash, err := bcrypt.GenerateFromPassword([]byte("123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	users := insertUsers(ctx, db, pwHash)
	posts := insertPosts(ctx, db, users)
	topComments := insertTopComments(ctx, db, users, posts)
	allComments := insertReplies(ctx, db, users, posts, topComments)
	insertFollows(ctx, db, users)
	insertPostLikes(ctx, db, users, posts)
	insertCommentLikes(ctx, db, users, allComments)

	log.Printf("Seed complete: %d users, %d posts, %d comments, follows + likes inserted",
		len(users), len(posts), len(allComments))
}

// ──────────────────────────── helpers ────────────────────────────

func queryRows[T any](db *sql.DB, query string, args ...any) ([]T, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []T
	for rows.Next() {
		var v T
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

// ──────────────────────────── Step 1: users ────────────────────────────

func insertUsers(ctx context.Context, db *sql.DB, pwHash []byte) []userRow {
	usernames := []string{
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

	bios := []string{
		"Go backend developer, open source enthusiast, passionate about distributed systems.",
		"Full-stack engineer focused on Go microservices and Vue.js frontend.",
		"CS grad student researching cloud-native and container orchestration.",
		"Three years of Go experience building REST APIs and gRPC services.",
		"DevOps engineer proficient with Docker, Kubernetes, and CI/CD pipelines.",
		"Deep expertise in Go concurrency and memory management, tech blogger.",
		"Backend developer exploring Rust and WebAssembly on the side.",
		"Data engineer building data pipelines and ETL workflows with Go.",
		"Startup CTO responsible for architecture and team leadership.",
		"Freelance developer taking on Go and Python projects.",
		"Database internals enthusiast and heavy PostgreSQL user.",
		"Security engineer focused on Go application auditing and pen testing.",
		"Embedded developer turned Go programmer, love systems programming.",
		"Frontend developer turned full-stack with Vue and Go tech stack.",
		"ML engineer building model serving and inference engines in Go.",
		"Blockchain developer passionate about smart contracts and dApps.",
		"Game server developer leveraging Go for high-concurrency backends.",
		"IoT developer building device management platforms with Go.",
		"Technical writer translating and authoring Go documentation.",
		"Open source contributor maintaining several small Go libraries.",
		"Interested in functional programming, exploring FP patterns in Go.",
		"Twenty years of programming, transitioned from C++ to Go.",
		"Go beginner learning through hands-on project practice.",
		"Cloud architect designing and implementing highly available services.",
		"Performance optimization specialist, proficient with pprof and trace.",
		"Network programming enthusiast building protocols and tools in Go.",
		"TDD practitioner, Go testing package is my best friend.",
		"API design enthusiast with experience in RESTful and GraphQL.",
		"Database tuning expert skilled in SQL optimization and indexing.",
		"Cache architect with deep experience in Redis and Memcached.",
		"Message queue developer working with Kafka and RabbitMQ.",
		"Search engine developer using Bleve and Elasticsearch for full-text search.",
		"Real-time communication developer with WebSocket and SSE expertise.",
		"Task scheduling system developer building distributed job queues.",
		"Configuration management expert using Viper and environment variables.",
		"Log system developer with ELK and Loki aggregation experience.",
		"Monitoring and alerting with Prometheus and Grafana integration.",
		"Containerization expert in Docker multi-stage builds and image optimization.",
		"Service mesh explorer with Istio and Linkerd hands-on experience.",
		"GraphQL developer building API services with Go.",
		"Web framework comparison expert across Chi, Gin, and Echo.",
		"Dependency injection explorer discussing DI patterns in Go.",
		"Code generation enthusiast using go generate for productivity.",
		"Performance testing expert with benchmarking and stress testing experience.",
		"Secure coding practitioner focused on OWASP and Go security guidelines.",
		"Database migration expert using golang-migrate and goose.",
		"API versioning specialist for RESTful version control strategies.",
		"Microservice communication expert comparing gRPC and REST.",
		"Active Go community participant attending Meetups and GopherCon.",
		"Technical interviewer helping teams recruit talented Go developers.",
	}

	avatars := make([]userRow, 0, 51)

	// 50 normal users
	for i := range 50 {
		suffix := rand.IntN(900) + 100 // random 3-digit number
		username := fmt.Sprintf("%s%d", usernames[i], suffix)
		email := fmt.Sprintf("%s%d_mock@example.com", usernames[i], suffix)
		bio := bios[i]

		svg, err := avatar.Generate(int64(i + 1))
		if err != nil {
			log.Printf("avatar generate for %s: %v", username, err)
			svg = []byte{}
		}

		var id int64
		err = db.QueryRowContext(ctx, `
			INSERT INTO users (username, email, password, is_active, role_id, bio, avatar, avatar_mime, links, show_email)
			VALUES ($1, $2, $3, true, (SELECT id FROM roles WHERE name = 'user'), $4, $5, 'image/svg+xml', '{}', false)
			RETURNING id`,
			username, email, pwHash, bio, svg,
		).Scan(&id)
		if err != nil {
			log.Fatalf("insert user %s: %v", username, err)
		}

		avatars = append(avatars, userRow{id: id, username: username})
	}

	// admin user — skip if already exists
	var adminID int64
	adminSvg, _ := avatar.Generate(9999)
	if err := db.QueryRowContext(ctx, `
		INSERT INTO users (username, email, password, is_active, role_id, bio, avatar, avatar_mime, links, show_email)
		VALUES ($1, $2, $3, true, (SELECT id FROM roles WHERE name = 'admin'), '', $4, 'image/svg+xml', '{}', false)
		ON CONFLICT (username) DO UPDATE SET username = EXCLUDED.username
		RETURNING id`,
		"admin", "admin@example.com", pwHash, adminSvg,
	).Scan(&adminID); err != nil {
		// fallback: fetch existing admin
		if qErr := db.QueryRowContext(ctx, `SELECT id FROM users WHERE username = 'admin'`).Scan(&adminID); qErr != nil {
			log.Fatalf("insert admin: %v; fetch admin: %v", err, qErr)
		}
	}
	avatars = append(avatars, userRow{id: adminID, username: "admin"})

	log.Printf("Inserted %d users", len(avatars))
	return avatars
}

// ──────────────────────────── Step 2: posts ────────────────────────────

var postContents = [54]string{
	`# Go Concurrency Programming: From Goroutines to Mastery

One of the most compelling features of Go is its built-in concurrency support. Goroutines are lightweight threads managed by the Go runtime, with negligible creation and teardown costs, making concurrent programming remarkably simple.

## What Is a Goroutine

A goroutine is a lightweight execution unit managed by the Go runtime. Unlike OS threads, goroutines start with a stack of just a few KB and can grow or shrink dynamically. This makes it feasible to run tens of thousands of goroutines simultaneously in a single program.

## Creating Goroutines

Creating a goroutine is straightforward — just prepend the ` + "`go`" + ` keyword before a function call. The Go runtime handles scheduling and management automatically.

` + "```go" + `
package main

import (
    "fmt"
    "time"
)

func worker(id int) {
    fmt.Printf("Worker %d starting\n", id)
    time.Sleep(time.Second)
    fmt.Printf("Worker %d done\n", id)
}

func main() {
    for i := 1; i <= 5; i++ {
        go worker(i)
    }
    time.Sleep(2 * time.Second)
}
` + "```" + `

## The GMP Scheduling Model

Go's scheduler uses the GMP model: G stands for Goroutine, M for OS thread, and P for logical processor. P defaults to the number of CPU cores and can be adjusted via GOMAXPROCS.

## Best Practices

- Always ensure the main goroutine waits for child goroutines to complete
- Use sync.WaitGroup to synchronize multiple goroutines
- Avoid panicking inside goroutines unless you have recover
- Control goroutine count to prevent resource exhaustion

Mastering goroutines is fundamental to writing efficient Go programs.`,

	`# Go Channels In Detail: Safe Concurrent Communication

Channels are the pipes through which goroutines communicate in Go. They follow the CSP (Communicating Sequential Processes) model, providing a type-safe way to pass data between goroutines.

## Channel Basics

A channel is a typed conduit that uses the send/receive operator to pass data. Channels come in two flavors: unbuffered and buffered.

## Unbuffered Channels

Unbuffered channels require both sender and receiver to be ready before communication can occur. This synchronization characteristic makes them ideal for coordinating goroutines.

` + "```go" + `
package main

import "fmt"

func main() {
    ch := make(chan string)
    go func() {
        ch <- "Hello from goroutine"
    }()
    msg := <-ch
    fmt.Println(msg)
}
` + "```" + `

## Buffered Channels

Buffered channels don't block the sender until the buffer is full, and don't block the receiver until the buffer is empty. This is useful in producer-consumer patterns.

## Channel Direction

You can restrict send or receive operations by specifying channel direction. Send-only and receive-only channels in function signatures improve code safety.

## Closing Channels

The sender can close a channel with close(), and receivers can check with v, ok := <-ch. Closing an already-closed channel panics.

## Select Statements

Select lets you choose among multiple channel operations, similar to a switch. When multiple cases are ready, select picks one at random.

Using channels properly leads to clear, safe concurrent code without explicit locks.`,

	`# Go Mutex Locks: The Foundation of Concurrency Safety

When multiple goroutines need to access shared resources, mutexes are essential for data integrity. Go's standard library provides sync.Mutex and sync.RWMutex.

## sync.Mutex Basics

Mutex is the most basic lock — only one goroutine can hold it at a time. Use Lock and Unlock to protect critical sections.

​` + "```go" + `
package main

import (
    "fmt"
    "sync"
)

type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

func (c *SafeCounter) Get() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.count
}

func main() {
    counter := &SafeCounter{}
    var wg sync.WaitGroup
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            counter.Increment()
        }()
    }
    wg.Wait()
    fmt.Println(counter.Get()) // 1000
}
` + "```" + `

## sync.RWMutex Read-Write Lock

RWMutex allows multiple concurrent reads but writes are exclusive. In read-heavy scenarios, RWMutex significantly outperforms Mutex.

## sync/atomic Operations

For simple counter operations, the atomic package provides efficient lock-free operations. For complex data structure protection, Mutex remains the better choice.

## Avoiding Deadlocks

Deadlock is one of the most common concurrency problems. The key to avoiding it: always acquire multiple locks in the same order, and use defer to ensure locks are always released.

Mutex is the foundation of Go concurrent programming — using it correctly prevents data races.`,

	`# Go Context: Request-Scoped Lifecycle Management

Context is Go's standard mechanism for passing request-scoped values, cancellation signals, and deadlines. Proper context usage is critical in web services and microservice architectures.

## Core Context Functions

Context provides three core capabilities: passing request-scoped values, propagating cancellation signals, and setting deadlines or timeouts.

## WithCancel: Cancellation Propagation

WithCancel returns a derived context and a cancel function. Calling cancel closes the Done channel, notifying all goroutines listening on this context to stop.

` + "```go" + `
package main

import (
    "context"
    "fmt"
    "time"
)

func worker(ctx context.Context) {
    select {
    case <-ctx.Done():
        fmt.Println("Worker cancelled:", ctx.Err())
        return
    case <-time.After(2 * time.Second):
        fmt.Println("Worker completed")
    }
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()
    go worker(ctx)
    time.Sleep(3 * time.Second)
}
` + "```" + `

## WithValue: Passing Request-Scoped Values

Context can carry request-scoped values like request IDs and auth info. Remember: context values should be request-level metadata, not function parameters.

## WithTimeout and WithDeadline

These functions set operation deadlines. WithTimeout takes a relative duration, WithDeadline takes an absolute time. The context auto-cancels after expiry.

## Context Best Practices

- Always pass context as the first function parameter
- Don't store context in structs
- Never pass nil context even if the function allows it
- Use context.Value sparingly to avoid abuse

Context is the backbone of inter-service communication in Go — using it correctly builds robust distributed systems.`,

	`# Go Error Handling Philosophy: Handling Errors Gracefully

Go uses explicit error handling instead of exceptions. This design makes error handling predictable but requires developers to address every potential error.

## Error Handling Basics

Go functions typically return (result, error) pairs. Callers must check if error is nil — this mandatory error checking makes code more robust.

## errors.Is and errors.As

Go 1.13 introduced errors.Is and errors.As, making error comparison and type assertions more convenient.

` + "```go" + `
package main

import (
    "errors"
    "fmt"
)

var ErrNotFound = errors.New("resource not found")

func findUser(id int) (string, error) {
    if id == 0 {
        return "", fmt.Errorf("find user %d: %w", id, ErrNotFound)
    }
    return "Alice", nil
}

func main() {
    _, err := findUser(0)
    if errors.Is(err, ErrNotFound) {
        fmt.Println("User not found")
    }
}
` + "```" + `

## Error Wrapping

Use fmt.Errorf with the %w verb to wrap errors, preserving the original error information. This is crucial for building error chains — callers can use errors.Is and errors.As to inspect underlying errors.

## Custom Error Types

For complex error scenarios, implement custom error types. By implementing the Error interface, you can carry additional context information.

## Sentinel Errors

Sentinel errors are predefined package-level error variables like io.EOF and sql.ErrNoRows. They let callers precisely determine error types.

## Error Handling Best Practices

- Handle each error only once — don't log and then return
- Add information to error context, don't replace it
- Use sentinel errors for recoverable conditions
- Use panic only for unrecoverable errors

Go's error handling philosophy emphasizes explicitness and predictability — more code but significantly more reliable programs.`,

	`# Go Interface Design: The Art of Implicit Implementation

Go interfaces are implicitly implemented — a type satisfies an interface simply by implementing all its methods, with no explicit declaration needed. This design makes interfaces more flexible and decoupled.

## Interface Definitions

An interface defines a set of method signatures. Any type that implements these methods automatically satisfies the interface.

## The Empty Interface interface{}

The empty interface has no method requirements, so every type satisfies it. This allows handling values of any type, similar to object in other languages.

` + "```go" + `
package main

import "fmt"

type Stringer interface {
    String() string
}

type Person struct {
    Name string
    Age  int
}

func (p Person) String() string {
    return fmt.Sprintf("%s (%d)", p.Name, p.Age)
}

func printInfo(s Stringer) {
    fmt.Println(s.String())
}

func main() {
    p := Person{Name: "Alice", Age: 30}
    printInfo(p)
}
` + "```" + `

## Interface Composition

Go interfaces can be composed to build more complex interfaces. The standard library's io.ReadWriter combines io.Reader and io.Writer. Small interface composition is more flexible than large interfaces.

## io.Reader and io.Writer

These are Go's most important interfaces. Reader provides a unified read abstraction, Writer provides a unified write abstraction. Nearly all I/O operations can be implemented through these two interfaces.

## Interface Best Practices

- Keep interfaces small and focused
- Define interfaces at the consumer side
- Don't define interfaces just for testing
- Return structs, accept interfaces

Go's implicit interface implementation promotes loose coupling and better code organization.`,

	`# Go Structs In Depth: Best Practices for Data Organization

Structs are the foundation of composite data types in Go, used to group related data fields together. Proper struct usage makes code clearer and type-safe.

## Struct Definition

Structs are defined with the type keyword. Each field has a name and type. The capitalization of the first letter determines export status.

## Struct Tags

Struct tags are metadata attached to fields, commonly used for JSON serialization and database mapping.

` + "```go" + `
package main

import (
    "encoding/json"
    "fmt"
    "time"
)

type User struct {
    ID        int64     ` + "`" + `json:"id"` + "`" + `
    Username  string    ` + "`" + `json:"username"` + "`" + `
    Email     string    ` + "`" + `json:"email"` + "`" + `
    CreatedAt time.Time ` + "`" + `json:"created_at"` + "`" + `
}

func main() {
    u := User{
        ID:        1,
        Username:  "alice",
        Email:     "alice@example.com",
        CreatedAt: time.Now(),
    }
    data, _ := json.Marshal(u)
    fmt.Println(string(data))
}
` + "```" + `

## Embedded Structs

Go supports struct embedding, achieving inheritance-like effects. Embedded fields are promoted to the outer struct and can be accessed directly.

## Method Receivers

Structs can have methods bound to specific types via receivers. Value receivers don't modify the original, while pointer receivers can.

## Zero Values

A struct's zero value has all fields set to their type's zero value. Understanding zero values is important for writing robust code.

Structs are the data skeleton of Go programs — mastering them is the foundation of writing high-quality Go code.`,

	`# Go Slices In Detail: Flexible Dynamic Arrays

Slices are one of Go's most commonly used data structures, providing a dynamic view over an underlying array. Understanding slice internals is crucial for writing efficient Go code.

## Slice Internals

A slice consists of three parts: a pointer to the underlying array, length (len), and capacity (cap). This design makes slice operations highly efficient.

## Creating Slices

Slices can be created via literals, the make function, or by slicing arrays. Each approach has its use cases.

` + "```go" + `
package main

import "fmt"

func main() {
    s1 := []int{1, 2, 3, 4, 5}
    s2 := make([]int, 3, 5)
    s3 := s1[1:4]

    fmt.Println(s1, s2, s3)
    fmt.Printf("len=%d cap=%d\n", len(s3), cap(s3))
}
` + "```" + `

## The append Function

append is the primary way to add elements to a slice. When capacity is insufficient, append automatically allocates a new underlying array. Note that append may return a new slice header.

## The copy Function

copy duplicates slice contents, preventing accidental modifications caused by multiple slices sharing the same underlying array.

## Common Slice Pitfalls

- Slices are reference types — assignment and passing don't copy the underlying array
- append may trigger reallocation of the underlying array
- Watch out for capacity traps to avoid memory leaks
- Use copy instead of direct assignment to avoid shared underlying arrays

Slices are at the heart of Go programming — understanding their internals prevents many common bugs.`,

	`# Go Maps In Depth: Key-Value Data Structures

Maps are Go's built-in key-value data structures, providing fast lookup, insertion, and deletion. Understanding map concurrency safety and performance characteristics is important for efficient Go code.

## Map Basics

Maps are created via make or literals. Key types must be comparable; value types have no restrictions.

## Concurrency Safety

Go maps are not concurrency-safe. Multiple goroutines reading and writing concurrently causes data races. Use sync.Mutex or sync.RWMutex for protection.

` + "```go" + `
package main

import (
    "fmt"
    "sync"
)

type SafeMap struct {
    mu sync.RWMutex
    m  map[string]int
}

func (sm *SafeMap) Get(key string) (int, bool) {
    sm.mu.RLock()
    defer sm.mu.RUnlock()
    val, ok := sm.m[key]
    return val, ok
}

func (sm *SafeMap) Set(key string, val int) {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    sm.m[key] = val
}

func main() {
    sm := &SafeMap{m: make(map[string]int)}
    sm.Set("count", 42)
    if v, ok := sm.Get("count"); ok {
        fmt.Println(v)
    }
}
` + "```" + `

## Map Operations

- Use the delete built-in to remove key-value pairs
- Use the v, ok := m[key] pattern for safe value retrieval
- Iterate maps with range — order is random

## Map Performance

Map lookup, insertion, and deletion are O(1) on average. Under high concurrency, lock overhead can become a bottleneck — consider sharded maps in such cases.

Maps are a critical Go data structure — using them correctly produces efficient and safe concurrent programs.`,

	`# Go Pointers Complete Guide: Understanding Memory References

Pointers are Go's tool for direct memory manipulation. While Go has garbage collection, understanding pointers remains important for writing efficient programs.

## Pointer Basics

A pointer stores a memory address. Use & to get a variable's address, * to dereference and get the value.

## Pointers and Functions

Pointers allow functions to modify the caller's variables. This is useful for returning multiple results or modifying passed parameters.

​` + "```go" + `
package main

import "fmt"

func increment(p *int) {
    *p++
}

func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("division by zero")
    }
    return a / b, nil
}

func main() {
    x := 10
    increment(&x)
    fmt.Println(x) // 11

    result, err := divide(10, 3)
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Println(result)
    }
}
` + "```" + `

## new and make

new allocates memory and returns a pointer; make initializes slices, maps, and channels. Understanding the difference is key to using Go's built-in types correctly.

## Pointer Receivers

Methods can be defined on pointer receivers, allowing them to modify the receiver's state. In performance-sensitive scenarios, pointer receivers also avoid value copying.

## Escape Analysis

Go's compiler automatically performs escape analysis to decide whether variables are allocated on the stack or heap. Use ` + "`go build -gcflags=-m`" + ` to view escape analysis results.

While Go has garbage collection, understanding pointers still helps you write more efficient code.`,

	`# Go Testing in Practice: Unit Tests and Table-Driven Tests

Go's testing package provides powerful testing support. Table-driven tests are the Go community's recommended pattern for covering multiple scenarios with a single set of test cases.

## Test File Naming

Go test files end with _test.go, and test functions start with Test. This convention lets go test automatically discover and run tests.

## Table-Driven Tests

Table-driven tests organize inputs and expected outputs in a table, using a loop to execute each case. This makes it easy to add new cases, and failures clearly show which case broke.

` + "```go" + `
package main

import "testing"

func Add(a, b int) int {
    return a + b
}

func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive", 1, 2, 3},
        {"zero", 0, 0, 0},
        {"negative", -1, -2, -3},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("Add(%d, %d) = %d, want %d",
                    tt.a, tt.b, result, tt.expected)
            }
        })
    }
}
` + "```" + `

## Subtests

Use t.Run to create subtests for clearer output. Subtests can also run independently for easier debugging.

## TestMain

TestMain lets you run setup and teardown before and after all tests, such as database connections.

## Test Coverage

Use go test -cover to view test coverage. High coverage doesn't guarantee quality, but low coverage usually means insufficient testing.

Good testing is the guarantee of software quality, and Go's testing package provides strong support for writing high-quality tests.`,

	`# Go Benchmarks: The Performance Optimization Tool

Benchmarks are essential for evaluating code performance. Go's testing package has built-in benchmark support, letting you quantify code execution efficiency.

## Benchmark Functions

Benchmark functions start with Benchmark and take a *testing.B parameter. The function body uses a for loop to run the code under test.

## Running Benchmarks

Run benchmarks with go test -bench. The -benchmem flag shows memory allocation information.

` + "```go" + `
package main

import "testing"

func fibonacci(n int) int {
    if n <= 1 {
        return n
    }
    return fibonacci(n-1) + fibonacci(n-2)
}

func BenchmarkFibonacci10(b *testing.B) {
    for i := 0; i < b.N; i++ {
        fibonacci(10)
    }
}

func BenchmarkFibonacci20(b *testing.B) {
    for i := 0; i < b.N; i++ {
        fibonacci(20)
    }
}
` + "```" + `

## Performance Comparison

Use benchstat to compare benchmark results across versions, helping you quantify performance improvements.

## Memory Profiling

Combine with pprof to deeply analyze memory allocation and usage. Go's memory profiler can pinpoint memory hotspots precisely.

## Benchmark Best Practices

- Reset the timer after setup with b.ResetTimer()
- Report allocations with b.ReportAllocs()
- Prevent compiler optimizations from eliminating benchmark results
- Use b.RunParallel to test concurrent performance

Benchmarks are the foundation of performance optimization — quantitative analysis leads to targeted improvements.`,

	`# Go HTTP Server: Building Web Services

Go's standard net/http package provides powerful HTTP server capabilities. Combined with third-party routers like Chi, you can quickly build production-grade web services.

## Basic HTTP Server

Use http.ListenAndServe to quickly start an HTTP server. For more complex routing, Chi or Gin are recommended.

## Chi Router

Chi is a lightweight HTTP router fully compatible with the standard http.Handler interface.

` + "```go" + `
package main

import (
    "encoding/json"
    "log"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

func main() {
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

    r.Get("/api/users", func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]string{
            "message": "hello",
        })
    })

    log.Fatal(http.ListenAndServe(":8080", r))
}
` + "```" + `

## Middleware Pattern

Middleware functions execute before or after request handling. Composing middleware implements cross-cutting concerns like logging, auth, and CORS.

## Graceful Shutdown

Production environments need graceful shutdown to finish in-progress requests when receiving termination signals.

## Request Timeouts

Use context.WithTimeout or middleware to set request timeouts, preventing slow requests from exhausting server resources.

Go's HTTP servers perform excellently — with the right framework and best practices, you can build high-concurrency web services.`,

	`# Go Middleware: Request Processing Pipelines

Middleware is the standard pattern for implementing cross-cutting concerns (logging, auth, CORS) in web apps. In Go, middleware is a function that takes http.Handler and returns http.Handler.

## Middleware Chain

Multiple middleware can be chained to form a request processing pipeline. Each middleware executes before or after the next handler.

## Logging Middleware

Logging middleware records each request's method, path, and processing time — invaluable for debugging and monitoring.

` + "```go" + `
package main

import (
    "log"
    "net/http"
    "time"
)

func Logger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
    })
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello"))
    })
    http.ListenAndServe(":8080", Logger(mux))
}
` + "```" + `

## Auth Middleware

Auth middleware validates JWT tokens in requests, returning 401 on failure. After validation, user info is stored in context.

## CORS Middleware

CORS middleware handles cross-origin requests by setting appropriate response headers. In separated frontend-backend apps, CORS configuration is essential.

## Rate Limiting

Rate limiting middleware prevents excessive client requests, protecting server resources. Limits can be based on IP, user ID, or API key.

Middleware is key to building maintainable web apps — it decouples cross-cutting concerns from business logic.`,

	`# Go JSON Processing: Serialization and Deserialization

JSON is the most common data exchange format in web APIs. Go's encoding/json package provides complete JSON processing support.

## Serialization

Use json.Marshal to convert structs to JSON byte slices. Struct tags control JSON field names and behavior.

## Deserialization

Use json.Unmarshal to convert JSON byte slices to structs. Unknown fields are ignored by default; use DisallowUnknownFields to reject them.

` + "```go" + `
package main

import (
    "encoding/json"
    "fmt"
)

type User struct {
    ID   int    ` + "`" + `json:"id"` + "`" + `
    Name string ` + "`" + `json:"name"` + "`" + `
}

func main() {
    data := []byte(` + "`" + `{"id": 1, "name": "Alice"}` + "`" + `)
    var u User
    json.Unmarshal(data, &u)
    fmt.Printf("%+v\n", u)

    out, _ := json.Marshal(u)
    fmt.Println(string(out))
}
` + "```" + `

## Streaming Processing

Use json.Decoder and json.Encoder for streaming JSON processing, avoiding loading entire payloads into memory.

## Custom Serialization

Implement json.Marshaler and json.Unmarshaler interfaces to customize a type's JSON serialization behavior.

JSON processing is a fundamental Go web development skill — mastering it enables efficient API data exchange.`,

	`# Go File I/O: Best Practices for Reading and Writing Files

Go provides rich file I/O capabilities, from simple reads/writes to efficient streaming, with os and io packages offering a complete toolkit.

## Reading Files

os.ReadFile is the simplest way to read an entire file. For large files, use os.Open with Scanner or BufferedReader for line-by-line reading.

## Writing Files

os.WriteFile writes an entire file. os.Create creates a new file; os.OpenFile controls file open modes.

` + "```go" + `
package main

import (
    "fmt"
    "os"
)

func main() {
    content := "Hello, Go file I/O!"
    err := os.WriteFile("test.txt", []byte(content), 0644)
    if err != nil {
        fmt.Println(err)
        return
    }

    data, err := os.ReadFile("test.txt")
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(string(data))
}
` + "```" + `

## io.Reader and io.Writer

These two interfaces are the core of Go I/O. By composing decorators like io.LimitReader and io.TeeReader, you can build complex I/O pipelines.

## File Permissions

Understanding Unix file permissions is important for proper access control. 0644 means owner can read/write, others read-only.

## Temp Files

Use os.CreateTemp and os.MkdirTemp to create temporary files and directories — remember to clean up afterward.

File I/O is a fundamental capability in any language, and Go provides a clean, powerful API for file operations.`,

	`# Go Database CRUD: Using database/sql

Go's database/sql package provides a unified database access interface. Combined with a PostgreSQL driver, you can implement complete CRUD operations.

## Connecting to the Database

Use sql.Open to create a database connection pool. The actual connection is established on the first query.

## Query Operations

Use QueryRow for single rows, Query for multiple rows. Use Scan to map column values to Go variables.

` + "```go" + `
package main

import (
    "database/sql"
    "fmt"
    _ "github.com/lib/pq"
)

type User struct {
    ID       int64
    Username string
    Email    string
}

func main() {
    db, _ := sql.Open("postgres", "postgres://localhost/mydb")
    defer db.Close()

    var u User
    err := db.QueryRow(
        "SELECT id, username, email FROM users WHERE id = $1", 1,
    ).Scan(&u.ID, &u.Username, &u.Email)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Printf("%+v\n", u)
}
` + "```" + `

## Insert Operations

Use Exec for INSERT statements. Use $1, $2 parameter placeholders to prevent SQL injection.

## Update and Delete

Use Exec for UPDATE and DELETE statements, getting affected row counts via RowsAffected.

## Transaction Handling

Use db.Begin to start a transaction, tx.Exec for operations, tx.Commit or tx.Rollback.

database/sql is the foundation of Go database programming — mastering it produces safe, efficient data access code.`,

	`# Go SQL Advanced: Query Optimization and Best Practices

After mastering basic SQL CRUD, advanced techniques help you write more efficient, safer database code.

## Prepared Statements

Prepared statements improve performance and prevent SQL injection. Use Prepare to create them, then Exec/Query to execute.

## Connection Pool Management

database/sql has a built-in connection pool. Properly configuring MaxOpenConns, MaxIdleConns, and ConnMaxLifetime is critical for performance.

` + "```go" + `
package main

import (
    "database/sql"
    "time"
    _ "github.com/lib/pq"
)

func main() {
    db, _ := sql.Open("postgres", "postgres://localhost/mydb")
    defer db.Close()

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)
}
` + "```" + `

## NULL Value Handling

Use sql.NullString, sql.NullInt64, etc. to handle database NULL values. These types include a Valid field indicating whether the value is NULL.

## Batch Operations

For large inserts, use transactions and batch INSERT statements for significant performance gains. Avoid N+1 problems with row-by-row inserts.

## Index Optimization

Proper index design is key to database performance. Use EXPLAIN ANALYZE to analyze query plans and ensure queries use indexes.

Advanced SQL techniques elevate your Go application's data layer to higher performance and reliability.`,

	`# Getting Started with GORM: Go's ORM Powerhouse

GORM is Go's most popular ORM library, offering rich features and a clean API. It supports multiple databases including PostgreSQL, MySQL, and SQLite.

## Model Definition

GORM uses structs to define data models, with tags specifying table names, column names, and relationships.

## CRUD Operations

GORM provides a fluent API that makes database operations very intuitive.

` + "```go" + `
package main

import (
    "fmt"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

type User struct {
    gorm.Model
    Username string
    Email    string
}

func main() {
    dsn := "host=localhost user=admin dbname=mydb"
    db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})

    db.Create(&User{Username: "alice", Email: "alice@example.com"})

    var user User
    db.First(&user, 1)

    db.Model(&user).Update("Email", "new@example.com")

    db.Delete(&user, 1)
}
` + "```" + `

## Associations

GORM supports one-to-one, one-to-many, and many-to-many associations, defined via tags for foreign keys and join tables.

## Hooks

GORM provides rich hook functions for executing custom logic before and after create, update, and delete operations.

## Migrations

GORM's AutoMigrate can automatically create and update table structures, but professional migration tools are recommended for production.

GORM simplifies Go database operations, making it ideal for rapid development and prototyping.`,

	`# Go Authentication: Complete JWT Implementation

JSON Web Token (JWT) is the most common authentication mechanism in modern web applications. Go's golang-jwt library provides complete JWT support.

## JWT Structure

JWT consists of three parts: Header, Payload, and Signature. The Payload contains user info and claims; the Signature verifies token integrity.

## Generating Tokens

Use jwt.NewWithClaims to create tokens, then sign with a secret key.

` + "```go" + `
package main

import (
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

func main() {
    claims := jwt.MapClaims{
        "sub": 123,
        "exp": time.Now().Add(72 * time.Hour).Unix(),
        "iat": time.Now().Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, _ := token.SignedString([]byte("secret"))
    fmt.Println(tokenString)
}
` + "```" + `

## Validating Tokens

Use jwt.Parse to validate token signatures and expiry. Customize the KeyFunc to provide the verification key.

## Middleware Integration

Wrap JWT validation as HTTP middleware to verify tokens on every request and store user info in context.

## Refresh Mechanism

Implement token refresh — when an access token expires, use the refresh token to obtain a new access token.

JWT authentication is the foundation of secure web apps — correct implementation protects user accounts.`,

	`# Go Logging: Structured Logging in Practice

A good logging system is essential for production environments. Go's uber-go/zap library provides high-performance structured logging support.

## Why Structured Logging

Structured logging outputs logs as JSON format, making them easy for log aggregation systems (ELK, Loki) to parse and query.

## Zap Basics

Zap provides both Sprintf-style and structured logging APIs.

` + "```go" + `
package main

import "go.uber.org/zap"

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    logger.Info("server started",
        zap.String("addr", ":8080"),
        zap.Int("port", 8080),
    )

    logger.Error("database connection failed",
        zap.String("host", "localhost"),
        zap.Error(nil),
    )
}
` + "```" + `

## Log Levels

Use Info, Warn, and Error appropriately. Info for normal operations, Warn for potential issues, Error for error conditions.

## Log Rotation

Configure log rotation in production to prevent unbounded log file growth. Use lumberjack or system-level log management tools.

## Contextual Logging

Attach request IDs, user IDs, and other context to logs for easier issue tracing and debugging.

Structured logging is a critical component of building observable systems.`,

	`# Go Configuration Management: Viper and Environment Variables

Configuration management is fundamental to any application. The Go community has multiple approaches, with Viper and environment variables being the most common.

## Environment Variables

Use os.Getenv and os.LookupEnv to read environment variables — the 12-Factor App recommended approach.

## Viper Configuration Library

Viper supports multiple config formats (JSON, YAML, TOML, etc.) and automatically merges environment variables with config files.

` + "```go" + `
package main

import (
    "fmt"
    "os"
)

type Config struct {
    Addr      string
    DBUrl     string
    JWTSecret string
}

func LoadConfig() *Config {
    return &Config{
        Addr:      getEnv("ADDR", ":8080"),
        DBUrl:     getEnv("DB_URL", "postgres://localhost/mydb"),
        JWTSecret: getEnv("JWT_SECRET", ""),
    }
}

func getEnv(key, fallback string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return fallback
}

func main() {
    cfg := LoadConfig()
    fmt.Printf("%+v\n", cfg)
}
` + "```" + `

## Config Validation

Use the validator library to validate configuration, ensuring all required items are set and valid.

## Hot Reloading

Some configurations support hot reloading without restarting the app. Implement this carefully to ensure concurrency safety.

Good configuration management makes applications easier to deploy and maintain.`,

	`# Go Environment Variables: Best Practices

Environment variables are the foundational approach to configuration management, especially for containerized and cloud-native apps. Go's os package provides complete env var support.

## Reading Environment Variables

os.Getenv returns the value, or empty string if not set. os.LookupEnv also returns whether the variable exists.

## Default Value Pattern

Provide sensible defaults so applications work correctly in both development and production environments.

` + "```go" + `
package main

import (
    "os"
    "strconv"
    "time"
)

func getDuration(key string, fallback time.Duration) time.Duration {
    val := os.Getenv(key)
    if val == "" {
        return fallback
    }
    d, err := time.ParseDuration(val)
    if err != nil {
        return fallback
    }
    return d
}

func getInt(key string, fallback int) int {
    val := os.Getenv(key)
    if val == "" {
        return fallback
    }
    n, err := strconv.Atoi(val)
    if err != nil {
        return fallback
    }
    return n
}

func main() {
    timeout := getDuration("TIMEOUT", 30*time.Second)
    workers := getInt("WORKERS", 4)
    _ = timeout
    _ = workers
}
` + "```" + `

## .env Files

In development, use .env files to manage environment variables. In production, use real environment variables or secret management systems.

## Sensitive Information

Passwords, API keys, and other secrets should be passed via environment variables, not hardcoded in source or config files.

## Type Conversion

Environment variables are all strings — write helper functions to simplify type conversion.

Environment variables are a core 12-Factor App principle — using them correctly simplifies configuration and deployment.`,

	`# Go Caching Strategies: Redis Integration in Practice

Caching is a key technique for improving application performance. Redis is one of the most popular in-memory databases, and Go's go-redis library provides complete Redis client support.

## go-redis Basics

go-redis supports all major Redis commands, including strings, hashes, lists, sets, and more.

## Caching Patterns

Common patterns include Cache-Aside and Write-Through caching.

` + "```go" + `
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type User struct {
    ID       int64
    Username string
}

func GetUser(ctx context.Context, rdb *redis.Client, id int64) (*User, error) {
    key := fmt.Sprintf("user:%d", id)

    data, err := rdb.Get(ctx, key).Bytes()
    if err == nil {
        var user User
        json.Unmarshal(data, &user)
        return &user, nil
    }

    user := &User{ID: id, Username: "alice"}
    data, _ = json.Marshal(user)
    rdb.Set(ctx, key, data, 10*time.Minute)
    return user, nil
}
` + "```" + `

## Cache Invalidation

Set appropriate TTLs to avoid cache-database inconsistency. Use TTL or active invalidation strategies.

## Cache Penetration and Avalanche

Use bloom filters or null-value caching to prevent cache penetration. Use randomized TTLs to prevent cache avalanche.

Redis integration significantly improves application response time and throughput.`,

	`# Go Redis Advanced: Data Structures and Advanced Features

Redis is more than a simple key-value store — it supports rich data structures and advanced features for solving complex problems.

## Redis Data Structures

- String: The basic type for storing strings and numbers
- Hash: Ideal for objects, can update individual fields
- List: Ordered string list with push/pop on both ends
- Set: Unordered unique element collection
- Sorted Set: Ordered set with score associated to each element

## Pub/Sub

Redis Pub/Sub enables simple messaging systems.

` + "```go" + `
package main

import (
    "context"
    "fmt"
    "github.com/redis/go-redis/v9"
)

func main() {
    rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
    ctx := context.Background()

    rdb.Publish(ctx, "notifications", "Hello subscribers!")

    sub := rdb.Subscribe(ctx, "notifications")
    ch := sub.Channel()
    for msg := range ch {
        fmt.Println("Received:", msg.Payload)
    }
}
` + "```" + `

## Transactions and Pipeline

Redis transactions ensure atomic execution of multiple commands. Pipeline batches commands to reduce network round trips.

## Lua Scripting

Redis supports executing Lua scripts for complex atomic operations like distributed locks and rate limiters.

## Clustering and Sentinel

Redis Cluster provides automatic sharding and high availability. Sentinel provides monitoring and automatic failover.

Redis's rich feature set makes it indispensable for building high-performance applications.`,

	`# Go WebSocket Real-Time Communication

WebSocket provides full-duplex communication between browser and server. gorilla/websocket is Go's most popular WebSocket implementation.

## WebSocket Basics

Once a WebSocket connection is established, both client and server can send messages at any time without reconnecting.

## Server-Side Implementation

Implement a WebSocket server using the gorilla/websocket library.

` + "```go" + `
package main

import (
    "fmt"
    "net/http"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer conn.Close()

    for {
        messageType, message, err := conn.ReadMessage()
        if err != nil {
            break
        }
        conn.WriteMessage(messageType, message)
    }
}

func main() {
    http.HandleFunc("/ws", wsHandler)
    http.ListenAndServe(":8080", nil)
}
` + "```" + `

## Heartbeat Mechanism

Implement heartbeat detection to keep connections alive and detect disconnections promptly.

## Rooms and Broadcasting

Implement a Room concept to broadcast messages to all connections within the same room.

## Connection Management

Maintain connection lists and handle connection establishment, disconnection, and error conditions.

WebSocket is the key technology for building real-time applications like chat, collaborative editing, and live notifications.`,

	`# Go Microservice Architecture: From Monolith to Distributed

Microservice architecture splits applications into small, independently deployed services. Go's lightweight nature and excellent concurrency support make it the top choice for microservice development.

## Microservice Design Principles

- Single responsibility: each service handles one business function
- Independent deployment: services can be compiled, tested, and deployed independently
- Decentralized governance: each service chooses its own tech stack

## Inter-Service Communication

Microservices communicate via HTTP REST, gRPC, or message queues.

` + "```go" + `
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type Order struct {
    ID     int64   ` + "`" + `json:"id"` + "`" + `
    UserID int64   ` + "`" + `json:"user_id"` + "`" + `
    Total  float64 ` + "`" + `json:"total"` + "`" + `
}

func orderHandler(w http.ResponseWriter, r *http.Request) {
    order := Order{ID: 1, UserID: 123, Total: 99.99}
    json.NewEncoder(w).Encode(order)
}

func main() {
    http.HandleFunc("/api/orders", orderHandler)
    fmt.Println("Order service started on :8081")
    http.ListenAndServe(":8081", nil)
}
` + "```" + `

## Service Discovery

Use Consul, etcd, or Kubernetes DNS for service discovery.

## Configuration Center

Use Apollo, Nacos, or environment variables for distributed configuration.

## Distributed Tracing

Use Jaeger or Zipkin for distributed tracing to aid debugging and performance analysis.

Microservice architecture improves scalability and maintainability, but increases system complexity.`,

	`# Go Docker Practice: Containerizing Applications

Docker is the standard tool for modern application deployment. Go apps compile to static binaries, making them ideal for containers.

## Multi-Stage Builds

Docker multi-stage builds significantly reduce image size. Use a full Go image for building, then scratch or alpine for running.

## Dockerfile

A typical Go Dockerfile has build and run stages.

` + "```dockerfile" + `
# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server ./cmd/api

# Run stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
CMD ["./server"]
` + "```" + `

## Docker Compose

Use Docker Compose to define multi-container apps including application servers, databases, and caches.

## Health Checks

Define health checks in Dockerfile and Compose configs to ensure containers are running properly.

## Environment Variables

Pass configuration via environment variables, not hardcoded in images.

Docker containerization makes Go application deployment simple and consistent.`,

	`# Go Kubernetes Deployment: Production-Grade Orchestration

Kubernetes is the de facto standard for container orchestration. Deploying Go apps to Kubernetes enables auto-scaling, rolling updates, and self-healing.

## Deployment

Deployments define the desired state of your application, and Kubernetes maintains that state automatically.

## Service

Services provide stable network access to Pods. ClusterIP for internal access, LoadBalancer for external access.

## ConfigMaps and Secrets

Use ConfigMaps and Secrets to manage application configuration and sensitive information.

` + "```yaml" + `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gopher-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: gopher-api
  template:
    metadata:
      labels:
        app: gopher-api
    spec:
      containers:
      - name: api
        image: gopher-api:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_DSN
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: dsn
` + "```" + `

## HPA Auto-Scaling

Automatically adjust Pod count based on CPU or memory usage to handle traffic spikes.

## Health Checks

Configure liveness and readiness probes so Kubernetes can properly manage Pod lifecycle.

Kubernetes provides production-grade deployment and operations capabilities for Go applications.`,

	`# Go CI/CD Pipeline: Automated Build and Deploy

Continuous integration and continuous deployment are core practices in modern software development. GitHub Actions is one of the most popular CI/CD platforms.

## GitHub Actions Basics

GitHub Actions uses YAML files to define workflows, supporting automatic build, test, and deployment on code push.

## Go Project CI Configuration

A typical Go CI pipeline includes dependency installation, code checks, unit tests, and building.

` + "```yaml" + `
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.22'

    - name: Build
      run: go build -v ./...

    - name: Test
      run: go test -v ./...

    - name: Lint
      uses: golangci/golangci-lint-action@v4
` + "```" + `

## Code Quality Checks

Integrate golangci-lint for static code analysis to ensure code quality.

## Automated Testing

Run unit, integration, and end-to-end tests to ensure code changes don't introduce regressions.

## Deployment Strategies

Use blue-green or canary deployment strategies to safely roll new versions to production.

CI/CD pipelines are critical infrastructure for ensuring software quality and accelerating delivery.`,

	`# Go Error Wrapping and Chain Handling

Go 1.13 introduced error wrapping, allowing original error information to be preserved in the error chain. This is crucial for debugging and error handling.

## Error Wrapping

Use fmt.Errorf with the %w verb to wrap errors, preserving original error information.

## errors.Is Checking

errors.Is checks whether an error chain contains a specific error value.

## errors.As Extraction

errors.As extracts a specific error type from an error chain.

` + "```go" + `
package main

import (
    "errors"
    "fmt"
)

type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func validateAge(age int) error {
    if age < 0 {
        return &ValidationError{Field: "age", Message: "must be positive"}
    }
    return nil
}

func createUser(age int) error {
    if err := validateAge(age); err != nil {
        return fmt.Errorf("create user: %w", err)
    }
    return nil
}

func main() {
    err := createUser(-1)
    var ve *ValidationError
    if errors.As(err, &ve) {
        fmt.Printf("Validation error on field %s: %s\n", ve.Field, ve.Message)
    }
}
` + "```" + `

## Custom Error Types

Create custom error types by implementing the error interface to carry additional context.

## Sentinel Errors

Predefined package-level error variables representing specific error conditions.

Error wrapping and chain handling make Go error handling more flexible and powerful.`,

	`# Go Package Design Principles: Building Maintainable Codebases

Good package design is key to building maintainable Go codebases. Following design principles improves code readability and reusability.

## Single Responsibility per Package

Each package should have a clear responsibility. Avoid "god packages" — split different functionality into separate packages.

## Export Control

Control exports via capitalization. Only export necessary APIs, hiding internal implementation details.

## Dependency Direction

Package dependencies should point toward more abstract layers. Avoid circular dependencies — use interfaces to decouple packages.

` + "```go" + `
package store

import "context"

type UserStorage interface {
    Create(ctx context.Context, user *User) error
    GetById(ctx context.Context, id int64) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id int64) error
}

type PostgresUserStore struct {
    db *sql.DB
}

func NewPostgresUserStore(db *sql.DB) *PostgresUserStore {
    return &PostgresUserStore{db: db}
}

func (s *PostgresUserStore) Create(ctx context.Context, user *User) error {
    return nil
}
` + "```" + `

## Interface Design

Define interfaces at the consumer side, not the implementation side. Small interfaces are more flexible than large ones.

## Package Naming

Package names should be short, clear, and reflect functionality. Avoid generic names like util or common.

Good package design is the foundation of building large Go projects.`,

	`# Go Performance Optimization: From Analysis to Practice

Performance optimization is a critical Go development skill. With the right profiling tools and optimization strategies, you can significantly improve application performance.

## Profiling Tools

Go has built-in profiling tools: pprof for CPU and memory analysis, trace for scheduling and latency analysis.

## CPU Profiling

Use pprof for CPU profiling to find hot functions in your program.

## Memory Profiling

Use pprof to analyze memory allocation and find memory hotspots and leaks.

` + "```go" + `
package main

import (
    "net/http"
    _ "net/http/pprof"
    "runtime"
)

func main() {
    go func() {
        http.ListenAndServe("localhost:6060", nil)
    }()

    data := make([]byte, 1024*1024)
    _ = data
    runtime.GC()
}
` + "```" + `

## Common Optimization Strategies

- Reduce memory allocation: reuse objects with sync.Pool
- Avoid unnecessary copying: pass large structs via pointer
- Optimize string concatenation: strings.Builder is more efficient than +
- Use goroutines judiciously: avoid excessive concurrency

## Benchmark-Driven Optimization

Use benchmarks to quantify optimization effects and ensure improvements are real.

Performance optimization should be data-driven — profile first, then optimize.`,

	`# Go Memory Model and Garbage Collection

Go uses a concurrent tri-color mark-and-sweep garbage collector for efficient memory management. Understanding Go's memory model is important for writing high-performance programs.

## Memory Allocation

Go's memory allocator uses tcmalloc-inspired strategies, with different allocation approaches for different object sizes.

## GC Internals

Go's GC is concurrent and low-latency. It uses write barriers and tri-color marking to collect garbage while the application runs.

## GOGC Control

The GOGC environment variable controls GC trigger frequency. The default of 100 triggers GC when newly allocated memory reaches 100% of in-use memory.

` + "```go" + `
package main

import (
    "fmt"
    "runtime"
    "runtime/debug"
)

func main() {
    debug.SetGCPercent(50)

    runtime.GC()

    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    fmt.Printf("Alloc = %v MB\n", m.Alloc/1024/1024)
    fmt.Printf("TotalAlloc = %v MB\n", m.TotalAlloc/1024/1024)
    fmt.Printf("Sys = %v MB\n", m.Sys/1024/1024)
}
` + "```" + `

## Memory Escape

Variables may escape to heap allocation. Use ` + "`go build -gcflags=-m`" + ` to view escape analysis results.

## Memory Leaks

Common leak causes: goroutine leaks, unbounded global map growth, unclosed resources.

Understanding Go's memory model helps you write more efficient programs.`,

	`# Go Race Detector: Detecting Data Races

Data races are one of the most common bugs in concurrent programs. Go's built-in race detector can detect them at runtime.

## What Is a Data Race

A data race occurs when multiple goroutines access the same variable concurrently, and at least one is writing.

## Using the Race Detector

Enable with ` + "`go build -race`" + ` or ` + "`go test -race`" + `.

## Detection Example

` + "```go" + `
package main

import "sync"

func main() {
    counter := 0
    var wg sync.WaitGroup

    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            counter++ // data race!
        }()
    }

    wg.Wait()
}
` + "```" + `

Running ` + "`go run -race main.go`" + ` reports the location of the data race.

## Fixing Data Races

- Protect shared variables with sync.Mutex
- Use atomic operations via the atomic package
- Use channels for goroutine communication
- Use sync/atomic.Value for read-only values

## Production Use

The race detector adds 2-10x runtime overhead — not recommended for production. But always enable it in testing and CI.

The race detector is essential for finding concurrency bugs — use it widely during development and testing.`,

	`# Go Benchmarks Deep Dive: Advanced Performance Testing

Benchmarks are the foundation of performance optimization. Go's testing package provides rich benchmarking features for precise performance measurement.

## Benchmark Best Practices

- Use b.ResetTimer to exclude initialization time
- Use b.ReportAllocs to report memory allocations
- Use b.RunParallel to test concurrent performance

## Comparing Implementations

Use benchstat to compare performance across different versions.

## Concurrent Benchmarks

Use b.RunParallel to test code performance under concurrent scenarios.

` + "```go" + `
package main

import (
    "sync"
    "testing"
)

func BenchmarkMutex(b *testing.B) {
    var mu sync.Mutex
    counter := 0
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            mu.Lock()
            counter++
            mu.Unlock()
        }
    })
}

func BenchmarkAtomic(b *testing.B) {
    var counter int64
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            counter++
        }
    })
}
` + "```" + `

## Memory Allocation Analysis

Analyze memory allocation patterns of different implementations via benchmarks to choose the most memory-efficient approach.

## Performance Regression Detection

Run benchmarks in CI to automatically detect performance regressions.

Mastering benchmark techniques is key to becoming a performance optimization expert.`,

	`# Go Clean Code: Writing Highly Readable Code

Code readability is at the core of software quality. The Go community has rich style guides and best practices.

## gofmt and goimports

Use gofmt to auto-format code and goimports to manage imports automatically. These tools ensure consistent team code style.

## Naming Conventions

- Package names: short, lowercase, single word
- Exported functions: PascalCase
- Unexported functions: camelCase
- Constants: camelCase or PascalCase

## Function Design

- Functions should be short and do one thing
- Don't have too many parameters — consider struct encapsulation
- Avoid side effects — prefer pure functions

` + "```go" + `
package main

import "fmt"

func calculateDiscount(price float64, discountPercent int) float64 {
    if discountPercent < 0 || discountPercent > 100 {
        return price
    }
    return price * (1 - float64(discountPercent)/100)
}

type OrderRequest struct {
    ProductID  int64
    Quantity   int
    CouponCode string
}

func processOrder(req OrderRequest) error {
    return nil
}

func main() {
    final := calculateDiscount(100.0, 20)
    fmt.Println(final) // 80
}
` + "```" + `

## Comments

Comments should explain why, not what. Good code should be self-documenting.

## Error Handling

Never ignore errors. Handle every possible error with the ` + "`if err != nil`" + ` pattern.

Clean Code isn't a destination — it's a continuous improvement process.`,

	`# Go SOLID Principles: Object-Oriented Design Thinking

SOLID represents five fundamental principles of object-oriented design. While Go isn't a traditional OOP language, these principles still apply.

## Single Responsibility Principle (SRP)

Each type should have only one responsibility. If a struct handles multiple responsibilities, split it into smaller types.

## Open/Closed Principle (OCP)

Open for extension, closed for modification. Extend via interfaces and composition, not by modifying existing code.

## Liskov Substitution Principle (LSP)

Subtypes should be substitutable for their parent types. In Go, this means types implementing interfaces should maintain expected behavior.

## Interface Segregation Principle (ISP)

Don't force clients to depend on interfaces they don't use. Go's small interface design naturally satisfies ISP.

## Dependency Inversion Principle (DIP)

High-level modules shouldn't depend on low-level modules — both should depend on abstractions. In Go, dependency inversion is achieved via interfaces.

` + "```go" + `
package main

import "fmt"

type Notifier interface {
    Notify(message string) error
}

type EmailNotifier struct{}

func (e EmailNotifier) Notify(message string) error {
    fmt.Println("Email:", message)
    return nil
}

func sendAlert(n Notifier, msg string) {
    n.Notify(msg)
}

func main() {
    sendAlert(EmailNotifier{}, "Server is down!")
}
` + "```" + `

SOLID principles help you write more flexible, maintainable code.`,

	`# Go Design Patterns: Common Patterns and Practices

Design patterns are proven solutions to common software design problems. Go's clean syntax makes many patterns naturally idiomatic.

## Factory Pattern

Use constructors to encapsulate object creation logic, hiding implementation details.

## Singleton Pattern

Use sync.Once to ensure only one global instance exists.

## Observer Pattern

Use channels to implement publish-subscribe patterns.

## Strategy Pattern

Use interfaces to define algorithm families, allowing algorithms to vary independently from clients.

` + "```go" + `
package main

import "fmt"

type SortStrategy interface {
    Sort([]int) []int
}

type BubbleSort struct{}
func (b BubbleSort) Sort(arr []int) []int {
    result := make([]int, len(arr))
    copy(result, arr)
    for i := 0; i < len(result); i++ {
        for j := 0; j < len(result)-i-1; j++ {
            if result[j] > result[j+1] {
                result[j], result[j+1] = result[j+1], result[j]
            }
        }
    }
    return result
}

type QuickSort struct{}
func (q QuickSort) Sort(arr []int) []int {
    result := make([]int, len(arr))
    copy(result, arr)
    return result
}

type Sorter struct {
    strategy SortStrategy
}

func (s *Sorter) Sort(arr []int) []int {
    return s.strategy.Sort(arr)
}

func main() {
    sorter := &Sorter{strategy: BubbleSort{}}
    fmt.Println(sorter.Sort([]int{3, 1, 4, 1, 5}))
}
` + "```" + `

## Decorator Pattern

Wrap existing functions to add new behavior — Go middleware is the classic decorator application.

Choosing the right design patterns improves code maintainability and extensibility.`,

	`# Go System Design: Building Scalable Architectures

System design is an advanced software engineering skill. Go's concurrency features and standard library make building scalable systems relatively straightforward.

## Load Balancing

Use Nginx or HAProxy for HTTP load balancing. Go apps scale horizontally by adding instances to increase throughput.

## Database Sharding

When a single database can't handle the load, shard by user ID.

## Caching Strategy

Multi-level caching: L1 (in-process) + L2 (Redis) + L3 (database).

## Asynchronous Processing

Use message queues to decouple synchronous service calls.

` + "```go" + `
package main

import (
    "encoding/json"
    "fmt"
    "time"
)

type Event struct {
    Type    string          ` + "`" + `json:"type"` + "`" + `
    Payload json.RawMessage ` + "`" + `json:"payload"` + "`" + `
    Time    time.Time       ` + "`" + `json:"time"` + "`" + `
}

type EventHandler func(Event) error

type EventBus struct {
    handlers map[string][]EventHandler
}

func NewEventBus() *EventBus {
    return &EventBus{handlers: make(map[string][]EventHandler)}
}

func (eb *EventBus) Subscribe(eventType string, handler EventHandler) {
    eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

func (eb *EventBus) Publish(event Event) {
    for _, handler := range eb.handlers[event.Type] {
        go handler(event)
    }
}

func main() {
    bus := NewEventBus()
    bus.Subscribe("user.created", func(e Event) error {
        fmt.Println("Send welcome email")
        return nil
    })
    bus.Publish(Event{Type: "user.created", Time: time.Now()})
    time.Sleep(time.Second)
}
` + "```" + `

System design requires balancing consistency, availability, and partition tolerance.`,

	`# Go Production Best Practices

Deploying Go apps to production requires considering multiple aspects: graceful shutdown, health checks, monitoring, and logging.

## Graceful Shutdown

Listen for SIGINT and SIGTERM signals to gracefully shut down servers and clean up resources.

## Health Checks

Implement a /health endpoint for load balancers and orchestrators to check application status.

## Graceful Shutdown

Use signal.NotifyContext to listen for system signals.

` + "```go" + `
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "time"
)

func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
    defer stop()

    mux := http.NewServeMux()
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })

    server := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }

    go func() {
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()

    fmt.Println("Server started on :8080")
    <-ctx.Done()
    fmt.Println("Shutting down...")

    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    server.Shutdown(shutdownCtx)
}
` + "```" + `

## Metrics Collection

Use Prometheus to collect application metrics including request counts, latency distributions, and error rates.

## Structured Logging

Output JSON structured logs for easy log aggregation and analysis.

## Rate Limiting

Implement request rate limiting to protect the application from traffic spikes.

Production stability is the key to software success.`,

	`# Go Performance Tuning: Comprehensive CPU to Memory Optimization

Performance tuning is a systematic engineering effort requiring multi-dimensional analysis and optimization. Go provides rich tools to help you find performance bottlenecks.

## CPU Performance Optimization

Use pprof to analyze CPU usage and find hot functions. Common strategies include reducing syscalls, using more efficient algorithms, and leveraging cache friendliness.

## Memory Optimization

Use sync.Pool to reuse temporary objects and reduce GC pressure. Avoid unnecessary memory allocations — prefer stack allocation.

## Concurrency Optimization

Tune GOMAXPROCS to balance CPU utilization and context switching overhead.

## I/O Optimization

Use buffered I/O, batch operations, and async processing to optimize I/O performance.

## Network Optimization

Use HTTP/2, connection pooling, and compression to optimize network performance.

## Monitoring and Profiling

Build a comprehensive monitoring system to continuously track application performance metrics.

Performance tuning isn't a one-time effort — it's a continuous process.`,

	`# Go Testing Strategy: Building a Reliable Test Suite

High-quality testing is the guarantee of software reliability. Go provides rich testing tools and best practices.

## Test Pyramid

Unit tests should make up the majority, followed by integration tests, with end-to-end tests being the fewest.

## Table-Driven Tests

Use the table-driven test pattern to easily add new test cases.

## Mocks and Stubs

Use interfaces and dependency injection for testable code.

## Test Coverage

Track test coverage with go test -cover and coveralls.

## Benchmarks

Use benchmarks to measure performance and detect regressions in CI.

## Fuzz Testing

Go 1.18 introduced fuzz testing to automatically discover edge cases and abnormal inputs.

## Integration Tests

Use testcontainers-go for database and external service integration testing.

Building a reliable test suite requires ongoing investment and improvement.`,

	`# Go Dependency Injection: The Art of Decoupling

Dependency injection is the key technique for achieving loose coupling. Go has no built-in DI framework, but interfaces and constructors enable clean DI.

## Constructor Injection

Inject dependencies via constructors, ensuring all dependencies are ready at object creation time.

## Interface Decoupling

Define small interfaces to abstract dependencies, so high-level modules don't depend on concrete implementations.

## Wire Code Generation

Google's Wire library automates dependency injection through code generation.

` + "```go" + `
package main

import "fmt"

type UserRepository interface {
    FindByID(id int64) (string, error)
}

type PostgresUserRepo struct{}

func (r PostgresUserRepo) FindByID(id int64) (string, error) {
    return "Alice", nil
}

type UserService struct {
    repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
    return &UserService{repo: repo}
}

func (s *UserService) GetUser(id int64) (string, error) {
    return s.repo.FindByID(id)
}

func main() {
    repo := PostgresUserRepo{}
    svc := NewUserService(repo)
    name, _ := svc.GetUser(1)
    fmt.Println(name)
}
` + "```" + `

## DI in Testing

Isolate tests by injecting mock implementations for simpler, more reliable tests.

Dependency injection makes code more flexible, testable, and maintainable.`,

	`# Go Code Generation: Automating Repetitive Work

Go's code generation tools automate repetitive coding tasks. The go generate command and template system provide powerful generation capabilities.

## go generate

Add generate directives in code comments, then run go generate to execute them automatically.

## Template System

Use text/template and html/template to generate structured code.

## Stringer Tool

golang.org/x/tools/cmd/stringer automatically generates String() methods.

## go-bindata

Embed static files into Go binaries.

## Config Generation

Auto-generate configuration structs and parsing code from config files or schemas.

## API Client Generation

Auto-generate API client code from OpenAPI/Swagger specs.

## Contract Testing

Auto-generate test code from interface definitions to ensure implementations meet contracts.

Code generation improves development efficiency and reduces human error, but avoid overusing it.`,

	`# Go Build Optimization: Building Efficient Binaries

Go's compiler provides multiple optimization options for building smaller, faster binaries.

## Compile Optimization Flags

Use -ldflags to strip debug info and reduce binary size.

## Static Compilation

Set CGO_ENABLED=0 to build statically linked binaries.

## Cross-Compilation

Use GOOS and GOARCH environment variables for cross-compilation.

## Binary Size Optimization

Use -ldflags="-s -w" to remove symbol tables and debug info.

## Startup Time Optimization

Reduce init function complexity — avoid heavy computation at startup.

## Memory Usage Optimization

Tune GOGC and GOMEMLIMIT to control memory usage.

## Runtime Optimization

Use runtime.GOMAXPROCS to control concurrency, runtime.LockOSThread to bind OS threads.

Build optimization is essential for production-grade Go applications.`,

	`# Go Toolchain: The Efficiency Boosters

Go provides a rich toolchain covering development, testing, debugging, and deployment.

## go fmt

Auto-formats code to ensure consistent team code style. Nearly all Go editors support auto-formatting.

## go vet

Static analysis to detect common errors and suspicious code.

## go doc

View documentation and code examples — a great learning tool for the standard library.

## go test

Run tests and benchmarks with support for parallel and fuzz testing.

## go tool pprof

Performance profiling tool for CPU, memory, and goroutine profiles.

## go tool trace

Scheduling trace tool for analyzing program scheduling behavior and latency.

## Delve Debugger

Go's professional debugger supporting breakpoints, stepping, and variable inspection.

## gopls Language Server

Go's official language server providing code completion, go-to-definition, refactoring, and more.

Mastering these tools significantly boosts Go development efficiency and code quality.`,

	`# Go Project Structure: Organizing Large Codebases

A well-organized project structure is the foundation of large Go project success. The Go community has recommended project layouts.

## Standard Layout

Follow the golang-standards/project-layout recommended structure.

## cmd Directory

The cmd directory holds executable entry points — each subdirectory corresponds to one binary.

## internal Directory

The internal directory holds packages that shouldn't be importable externally — the compiler enforces this restriction.

## pkg Directory

The pkg directory holds library code that can be imported by external projects.

## Domain-Driven Design

Organize code by business domain, not technical layers.

## Multi-Module Repositories

For large projects, use Go workspaces or multi-module repositories to manage dependencies.

## Makefile

Wrap common development commands in a Makefile: build, test, deploy.

Good project structure improves code maintainability and team collaboration efficiency.`,

	`# Go Observability: Building Observable Systems

Observability is a key characteristic of modern distributed systems. Go applications need the three pillars: metrics, logs, and traces.

## Prometheus Metrics

Use prometheus/client_golang to expose application metrics.

## Log Aggregation

Use structured logging with ELK or Loki for log aggregation and querying.

## Distributed Tracing

Use OpenTelemetry or Jaeger for distributed tracing.

## Alerting

Set up sensible alert rules to detect issues promptly.

## SLI/SLO

Define service level indicators and objectives to quantify service quality.

## Dashboards

Build visual dashboards with Grafana to display system status at a glance.

## Chaos Engineering

Inject faults to validate system resilience.

Building observable systems requires considering monitoring and tracing needs from the development phase.`,

	`# Go Secure Coding: Protecting Applications from Attacks

Security is a critical aspect of software development that can't be ignored. Go provides multiple security features to help developers write secure code.

## SQL Injection Prevention

Use parameterized queries and prepared statements to prevent SQL injection.

## XSS Prevention

Use the html/template package to auto-escape user input and prevent cross-site scripting.

## CSRF Prevention

Use CSRF tokens to prevent cross-site request forgery.

## Authentication and Authorization

Implement secure auth mechanisms including password hashing, JWT, and OAuth.

## Input Validation

Validate all user input and reject unexpected data.

## Dependency Security

Regularly check dependencies for security vulnerabilities and update promptly.

## Security Audits

Conduct regular security audits and code reviews.

Secure coding is a continuous process requiring team-wide participation and commitment.`,

	`# Go Microservice Testing: End-to-End Validation

Testing in microservice architectures is more complex, requiring consideration of inter-service interactions and dependencies.

## Unit Tests

Isolate individual service business logic testing.

## Integration Tests

Use testcontainers to test service integration with external dependencies.

## Contract Testing

Use tools like Pact to validate API contracts between services.

## End-to-End Tests

Validate complete business flows in production-like environments.

## Performance Testing

Use k6 or vegeta for load testing and performance benchmarks.

## Chaos Testing

Inject faults to validate system resilience.

## Test Data Management

Use factory patterns and fixtures to manage test data.

Microservice testing requires a multi-layered, multi-strategy approach.`,

	`# Go and Cloud Native: Building Kubernetes-Native Applications

Cloud native is the trend in modern application development. Go's characteristics make it ideal for building cloud native applications.

## Kubernetes Operator

Build Kubernetes Operators with controller-runtime to extend cluster functionality.

## Client-Go

Use client-go to interact with the Kubernetes API.

## Health Checks

Implement liveness and readiness probes.

## Resource Management

Properly set CPU and memory requests and limits.

## Configuration Management

Use ConfigMaps and Secrets for configuration.

## Logging Standards

Follow Kubernetes logging standards — output to stdout/stderr.

## Metrics Exposure

Expose application metrics in Prometheus format.

Cloud native applications require considering containerization and orchestration needs from the design phase.`,

	`# Go Workspaces and Module Management

Go 1.18 introduced workspaces, improving the multi-module development experience.

## Go Modules Basics

go.mod defines the module and dependencies; go.sum records dependency checksums.

## Version Selection

Go uses Minimum Version Selection (MVS) to determine dependency versions.

## Workspaces

The go.work file defines a workspace for developing multiple modules simultaneously.

## Private Modules

Configure GOPRIVATE and GONOSUMDB to manage private modules.

## Replacing Dependencies

Use the replace directive to substitute dependency versions locally.

## Publishing Modules

Follow semantic versioning when publishing modules.

Module management is the foundation of Go project dependency management.`,

	`# Go CLI Tool Development: Building Command-Line Applications

Go is excellent for developing command-line tools. cobra and urfave/cli are the most popular CLI frameworks.

## cobra Basics

cobra provides subcommands, argument parsing, and help systems.

## Arguments and Flags

Support both positional arguments and named flags.

## Interactive Prompts

Use survey or bubbletea to create interactive CLI interfaces.

## Configuration Files

Support YAML, JSON, or TOML format configuration files.

## Output Formats

Support JSON, YAML, Table, and other output formats.

## Testing CLIs

Use os/exec to test CLI tool output.

## Cross-Compilation

Easily build binaries for multiple platforms.

CLI tools are indispensable assistants in developers' daily work.`,
}

var tagPool = [20]string{
	"go", "backend", "api", "http", "grpc",
	"database", "sql", "redis", "cache", "docker",
	"kubernetes", "microservice", "concurrency", "goroutine", "channel",
	"json", "auth", "jwt", "testing", "performance",
}

var commentTemplates = [25]string{
	"Great article, explains the core concepts clearly and thoroughly.",
	"I was just learning about this topic, very helpful.",
	"Could you go deeper into the error handling part?",
	"The code examples are very practical, ready to use in real projects.",
	"First time seeing such a clear concurrency tutorial.",
	"Suggest adding a complete hands-on project example.",
	"Might be a bit challenging for beginners, but overall great quality.",
	"Compared several articles, this one is the best.",
	"Can you publish a series? I want to learn systematically.",
	"You really do encounter these issues in real projects, very well summarized.",
	"The performance optimization section is very detailed, extremely helpful.",
	"The testing strategy section gave me a new perspective on Go testing.",
	"Do you have any recommended advanced reading materials?",
	"Tried this in production, works really well.",
	"Hope to see more content on secure coding.",
	"The code style is very clean, worth learning from.",
	"Very clear explanation, finally understand this concept.",
	"Very inspiring for architecture design in large projects.",
	"Well-structured article, flows nicely.",
	"Looking forward to more Go articles from you.",
	"This problem has been bugging me for a long time, finally understand after reading.",
	"Suggest adding some performance comparison data.",
	"Great practical guide, already promoted it within my team.",
	"Have you considered using generics to simplify the code?",
	"The best practices in this article are all battle-tested, very trustworthy.",
}

var replyTemplates = [10]string{
	"Thanks for your feedback, will consider adding more content.",
	"Good suggestion, hands-on project examples are really important.",
	"Yes, these lessons come from real-world projects.",
	"Already planning a series, stay tuned.",
	"Generics are a great direction, will consider adding them in future articles.",
	"Secure coding is important, will cover it in the next article.",
	"For performance comparison data, check out the benchmarking section.",
	"How did the team promotion go? Would love to hear feedback.",
	"These are common issues in real development, hope the article helped.",
	"Thanks for the support, will keep updating with more quality content.",
}

// ──────────────────────────── Step 2: insert posts ────────────────────────────

func insertPosts(ctx context.Context, db *sql.DB, users []userRow) []postRow {
	posts := make([]postRow, 0, 500)

	stmt, err := db.PrepareContext(ctx, `
		INSERT INTO posts (title, content, user_id, tags)
		VALUES ($1, $2, $3, $4)
		RETURNING id`)
	if err != nil {
		log.Fatalf("prepare posts: %v", err)
	}
	defer stmt.Close()

	for i := range 500 {
		user := users[rand.IntN(len(users)-1)] // exclude admin
		content := postContents[i%50]

		// pick 3 unique tags
		shuffled := rand.Perm(20)
		tags := []string{tagPool[shuffled[0]], tagPool[shuffled[1]], tagPool[shuffled[2]]}

		// extract title from first line of markdown
		title := extractTitle(content)

		var id int64
		err := stmt.QueryRowContext(ctx, title, content, user.id, tags).Scan(&id)
		if err != nil {
			log.Fatalf("insert post %d: %v", i, err)
		}
		posts = append(posts, postRow{id: id, userID: user.id})
	}

	log.Printf("Inserted %d posts", len(posts))
	return posts
}

func extractTitle(md string) string {
	// markdown title is the first line starting with #
	for i := 0; i < len(md); i++ {
		if md[i] == '\n' {
			title := md[:i]
			if len(title) > 100 {
				return title[:100]
			}
			return title
		}
	}
	if len(md) > 100 {
		return md[:100]
	}
	return md
}

// ──────────────────────────── Step 3: top-level comments ────────────────────────────

func insertTopComments(ctx context.Context, db *sql.DB, users []userRow, posts []postRow) []commentRow {
	comments := make([]commentRow, 0, 1250)

	stmt, err := db.PrepareContext(ctx, `
		INSERT INTO comments (user_id, post_id, content, parent_id)
		VALUES ($1, $2, $3, NULL)
		RETURNING id`)
	if err != nil {
		log.Fatalf("prepare comments: %v", err)
	}
	defer stmt.Close()

	for i := range 1250 {
		user := users[rand.IntN(len(users)-1)]
		post := posts[rand.IntN(len(posts))]

		content := commentTemplates[rand.IntN(len(commentTemplates))]
		// make some comments longer by combining templates
		if rand.IntN(3) == 0 {
			extra := commentTemplates[rand.IntN(len(commentTemplates))]
			if content != extra {
				content = content + " " + extra
			}
		}

		var id int64
		err := stmt.QueryRowContext(ctx, user.id, post.id, content).Scan(&id)
		if err != nil {
			log.Fatalf("insert comment %d: %v", i, err)
		}
		comments = append(comments, commentRow{id: id, userID: user.id, postID: post.id})
	}

	log.Printf("Inserted %d top-level comments", len(comments))
	return comments
}

// ──────────────────────────── Step 4: replies ────────────────────────────

func insertReplies(ctx context.Context, db *sql.DB, users []userRow, posts []postRow, topComments []commentRow) []commentRow {
	replies := make([]commentRow, 0, 500)

	stmt, err := db.PrepareContext(ctx, `
		INSERT INTO comments (user_id, post_id, content, parent_id, reply_to_user_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`)
	if err != nil {
		log.Fatalf("prepare replies: %v", err)
	}
	defer stmt.Close()

	for i := range 500 {
		user := users[rand.IntN(len(users)-1)]
		parent := topComments[rand.IntN(len(topComments))]

		content := replyTemplates[rand.IntN(len(replyTemplates))]

		var id int64
		err := stmt.QueryRowContext(ctx, user.id, parent.postID, content, parent.id, parent.userID).Scan(&id)
		if err != nil {
			log.Fatalf("insert reply %d: %v", i, err)
		}
		replies = append(replies, commentRow{id: id, userID: user.id, postID: parent.postID})
	}

	log.Printf("Inserted %d replies", len(replies))
	return replies
}

// ──────────────────────────── Step 5: follows ────────────────────────────

func insertFollows(ctx context.Context, db *sql.DB, users []userRow) {
	type pair struct{ a, b int64 }
	used := make(map[pair]bool)

	stmt, err := db.PrepareContext(ctx, `
		INSERT INTO followers (user_id, follower_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING`)
	if err != nil {
		log.Fatalf("prepare follows: %v", err)
	}
	defer stmt.Close()

	// exclude admin from following
	normalUsers := users[:len(users)-1]

	for _, u := range normalUsers {
		count := 0
		for count < 10 {
			target := normalUsers[rand.IntN(len(normalUsers))]
			if target.id == u.id {
				continue
			}
			p := pair{target.id, u.id}
			if used[p] {
				continue
			}
			used[p] = true

			if _, err := stmt.ExecContext(ctx, target.id, u.id); err != nil {
				log.Printf("insert follow: %v", err)
			}
			count++
		}
	}

	log.Printf("Inserted %d follows", len(used))
}

// ──────────────────────────── Step 6: post likes ────────────────────────────

func insertPostLikes(ctx context.Context, db *sql.DB, users []userRow, posts []postRow) {
	type pair struct{ a, b int64 }
	used := make(map[pair]bool)

	stmt, err := db.PrepareContext(ctx, `
		INSERT INTO post_likes (post_id, user_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING`)
	if err != nil {
		log.Fatalf("prepare post likes: %v", err)
	}
	defer stmt.Close()

	normalUsers := users[:len(users)-1]

	for _, u := range normalUsers {
		count := 0
		for count < 20 {
			post := posts[rand.IntN(len(posts))]
			// skip own posts
			if post.userID == u.id {
				continue
			}
			p := pair{post.id, u.id}
			if used[p] {
				continue
			}
			used[p] = true

			if _, err := stmt.ExecContext(ctx, post.id, u.id); err != nil {
				log.Printf("insert post like: %v", err)
			}
			count++
		}
	}

	log.Printf("Inserted %d post likes", len(used))
}

// ──────────────────────────── Step 7: comment likes ────────────────────────────

func insertCommentLikes(ctx context.Context, db *sql.DB, users []userRow, allComments []commentRow) {
	type pair struct{ a, b int64 }
	used := make(map[pair]bool)

	stmt, err := db.PrepareContext(ctx, `
		INSERT INTO comment_likes (comment_id, user_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING`)
	if err != nil {
		log.Fatalf("prepare comment likes: %v", err)
	}
	defer stmt.Close()

	normalUsers := users[:len(users)-1]

	for _, u := range normalUsers {
		count := 0
		for count < 20 {
			comment := allComments[rand.IntN(len(allComments))]
			// skip own comments
			if comment.userID == u.id {
				continue
			}
			p := pair{comment.id, u.id}
			if used[p] {
				continue
			}
			used[p] = true

			if _, err := stmt.ExecContext(ctx, comment.id, u.id); err != nil {
				log.Printf("insert comment like: %v", err)
			}
			count++
		}
	}

	log.Printf("Inserted %d comment likes", len(used))
}
