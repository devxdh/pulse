# Go HTTP and Integration Testing Guide

Testing in Go is simple, powerful, and built directly into the standard library. Because you are already using table-driven tests for your configuration, applying those patterns to HTTP handlers and middleware is a natural next step.

---

## 1. What You SHOULD and SHOULD NOT Test

Before writing code, let's draw clear boundaries around testing responsibilities:

### 🟢 What to Test
* **HTTP Status Codes and Response Formats:** Does `/api/job` return `201 Created` with the correct JSON structure? Does `/api/job/999` return `404 Not Found`?
* **Request Validation:** Does sending malformed JSON to `POST /api/job` return a `400 Bad Request`?
* **Middleware Logic:** Does your token bucket rate-limiter block requests (returning `429 Too Many Requests`) once the limit is exceeded?
* **Async Job Worker Side Effects:** When a worker finishes processing a job, does the state in the database transition to `completed`?

### 🔴 What NOT to Test
* **Standard Library and Third-Party Packages:** Do not write unit tests to verify that `pgx` can connect to PostgreSQL, or that `http.ServeMux` knows how to match routes. Trust their library-level test suites.
* **Trivial Logic:** Do not test basic getters/setters or simple logging statements.
* **Your `main()` Function Entrypoint:** Testing the `main()` function is notoriously difficult because it starts the server listening on a port and blocks. Leave `main.go` for integration/E2E testing or manual validation.

---

## 2. HTTP Testing in Go: The Core Concepts

Go provides the `net/http/httptest` package in the standard library. This package allows you to test HTTP handlers **in-memory**, without spinning up a real TCP listener or sending traffic over network sockets. This makes HTTP unit tests incredibly fast (sub-millisecond).

Here are the two most important tools in `httptest`:

### `httptest.NewRequest(...)`
Creates a mock `*http.Request` directly in memory. You can set the HTTP method, URL path, headers, and request body.
```go
req := httptest.NewRequest("POST", "/api/job", strings.NewReader(`{"payload": {"task": "run"}}`))
```

### `httptest.NewRecorder()`
Returns a `*httptest.ResponseRecorder`, which implements the `http.ResponseWriter` interface. When you pass it to your handler, the handler writes to it just like it would to a real client. Afterward, you can inspect the status code, headers, and response body:
```go
rr := httptest.NewRecorder()
handler.ServeHTTP(rr, req)

// Check the results
if rr.Code != http.StatusCreated {
    t.Errorf("expected 201, got %d", rr.Code)
}
```

---

## 3. Dealing with the Database Dependency

Your API handlers require a connection to PostgreSQL via `pgxpool.Pool` to function. You have two strategies to deal with this in tests:

### Strategy A: Real Database (Integration Testing) - *Recommended for DB-Heavy Apps*
Create a dedicated test database (e.g., `pulse_test`). When the tests start, run migrations/DDL, execute your HTTP handler tests using a pool pointing to the test DB, and truncate the tables between tests.
* **Pros:** Highly accurate. Mocks can lie; the real database won’t.
* **Cons:** Slower, and requires a running database instance to run tests.

### Strategy B: Mocking (Unit Testing)
Abstract the database queries into a Go interface and swap the real implementation with a mock in tests.
* **Pros:** Blazing fast; requires no external services.
* **Cons:** Requires refactoring code to use interfaces, and you aren't testing that your SQL queries are syntax-correct.

Given Phase 1's goal is to master PostgreSQL state machines, **Strategy A (Integration testing with a clean test database)** is the most robust and realistic way to learn.

---

## 4. Let's Build a Test: The Rate Limiter Middleware

The rate limiter in `internal/middleware/ratelimit.go` is perfect for a pure unit test because it doesn't touch the database.

Here is how you write a test for it:

```go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimiter(t *testing.T) {
	// 1. Create a rate limiter that allows exactly 2 requests per minute, with a capacity of 2
	rl := NewRateLimiter(2.0/60.0, 2.0)

	// 2. Create a dummy handler that we want to protect
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// 3. Wrap it with our middleware
	limitedHandler := rl.Limit(nextHandler)

	// 4. Test Case 1: First request should be allowed (200 OK)
	req1 := httptest.NewRequest("GET", "/", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	rr1 := httptest.NewRecorder()
	limitedHandler.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr1.Code)
	}

	// 5. Test Case 2: Second request should be allowed (200 OK)
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	rr2 := httptest.NewRecorder()
	limitedHandler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr2.Code)
	}

	// 6. Test Case 3: Third request exceeds capacity and should be blocked (429 Too Many Requests)
	req3 := httptest.NewRequest("GET", "/", nil)
	req3.RemoteAddr = "192.168.1.1:1234"
	rr3 := httptest.NewRecorder()
	limitedHandler.ServeHTTP(rr3, req3)

	if rr3.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", rr3.Code)
	}
}
```

---

## 5. Writing HTTP Integration Tests with a Database

To test `CreateJob` or `GetJobByID`, you'll want to run them against a local test database.

Here is a template structure showing how you can test `CreateJob`:

```go
package apihandler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devxdh/pulse/pkg/db"
)

func TestCreateJob_Integration(t *testing.T) {
	// 1. Set up a connection pool pointing to a local test database
	testConnStr := "postgresql://postgres:postgres@localhost:5432/pulse_test"
	
	pool, err := db.InitDB(testConnStr)
	if err != nil {
		t.Skip("Skipping integration test; test database not reachable:", err)
	}
	defer pool.Close()

	// 2. Ensure schema is injected
	if err := db.InjectDDL(pool); err != nil {
		t.Fatalf("Failed to run test DDL: %v", err)
	}

	// 3. Clean up the database table before/after running the test
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		pool.Exec(ctx, "TRUNCATE TABLE jobs RESTART IDENTITY CASCADE;")
	})

	// 4. Instantiate the handler environment with the test pool
	env := New(pool, 10)

	// 5. Setup the HTTP mock request and response recorder
	body := []byte(`{"payload": {"test_key": "test_value"}}`)
	req := httptest.NewRequest("POST", "/api/job", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// 6. Invoke the handler directly
	env.CreateJob(rr, req)

	// 7. Verify assertions
	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", rr.Code, rr.Body.String())
	}
}
```
