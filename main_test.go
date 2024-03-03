package main

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-redis/redis/v8"
)

// Server starts and listens on port 8080
func TestServerStartsAndListensOnPort8080(t *testing.T) {
	// Mock Redis client connection
	mockClient := &redis.Client{}
	mockClient.On("Close").Return(nil)

	// Create rate limiter with mock Redis client
	limiter := NewRateLimiter(10, 100, time.Second, time.Minute, &RedisRateLimiterStrategy{mockClient})

	// Create a new router
	r := chi.NewRouter()

	// Use the rate limiter middleware
	r.Use(limiter.Handler)

	// Define a test route
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!\n"))
	})

	// Start the server in a separate goroutine
	go func() {
		err := http.ListenAndServe(":8080", r)
		if err != nil {
			t.Errorf("Failed to start server: %v", err)
		}
	}()

	// Wait for the server to start listening
	time.Sleep(100 * time.Millisecond)

	// Send a GET request to the test route
	resp, err := http.Get("http://localhost:8080/")
	if err != nil {
		t.Errorf("Failed to send GET request: %v", err)
	}

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Stop the server gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := http.Shutdown(ctx); err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}

	// Assert that the Redis client connection was closed
	mockClient.AssertCalled(t, "Close")
}

// Redis client connection fails
func TestRedisClientConnectionFails(t *testing.T) {
	// Mock Redis client connection error
	mockClient := &redis.Client{}
	mockClient.On("Close").Return(errors.New("connection failed"))

	// Create rate limiter with mock Redis client
	limiter := NewRateLimiter(10, 100, time.Second, time.Minute, &RedisRateLimiterStrategy{mockClient})

	// Create a new router
	r := chi.NewRouter()

	// Use the rate limiter middleware
	r.Use(limiter.Handler)

	// Define a test route
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!\n"))
	})

	// Start the server in a separate goroutine
	go func() {
		err := http.ListenAndServe(":8080", r)
		if err != nil {
			t.Errorf("Failed to start server: %v", err)
		}
	}()

	// Wait for the server to start listening
	time.Sleep(100 * time.Millisecond)

	// Send a GET request to the test route
	_, err := http.Get("http://localhost:8080/")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	// Stop the server gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := http.Shutdown(ctx); err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}

	// Assert that the Redis client connection was closed
	mockClient.AssertCalled(t, "Close")
}
