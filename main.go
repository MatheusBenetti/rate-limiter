package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v4/pgxpool"
)

type RateLimiter struct {
	dbPool      *pgxpool.Pool
	maxRequests int
	timeWindow  time.Duration
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func NewRateLimiter(dbPool *pgxpool.Pool, maxRequests int, timeWindow time.Duration) (*RateLimiter, error) {
	if maxRequests <= 0 {
		return nil, fmt.Errorf("max requests must be greater than zero")
	}
	if timeWindow <= 0 {
		return nil, fmt.Errorf("time window must be greater than zero")
	}

	return &RateLimiter{
		dbPool:      dbPool,
		maxRequests: maxRequests,
		timeWindow:  timeWindow,
	}, nil
}

func (rl *RateLimiter) Limit(c *gin.Context) {
	ip := c.ClientIP()
	token := c.GetHeader("API_KEY")

	var err error
	if token != "" {
		err = rl.limitByToken(token)
	} else {
		err = rl.limitByIP(ip)
	}

	if err != nil {
		c.JSON(http.StatusTooManyRequests, ErrorResponse{Message: err.Error()})
		c.Abort()
		return
	}
}

func (rl *RateLimiter) limitByIP(ip string) error {
	var requests int
	err := rl.dbPool.QueryRow(context.Background(), "SELECT COUNT(*) FROM requests WHERE ip = $1 AND created_at > NOW() - INTERVAL '1 minute'", ip).Scan(&requests)

	if err != nil {
		return fmt.Errorf("error counting IP requests: %w", err)
	}

	if requests > rl.maxRequests {
		return rl.blockIP(ip)
	}

	return nil
}

func (rl *RateLimiter) limitByToken(token string) error {
	var requests int
	err := rl.dbPool.QueryRow(context.Background(), "SELECT COUNT(*) FROM requests WHERE token = $1 AND created_at > NOW() - INTERVAL '1 minute'", token).Scan(&requests)

	if err != nil {
		return fmt.Errorf("error counting token requests: %w", err)
	}

	if requests > rl.maxRequests {
		return rl.blockToken(token)
	}

	return nil
}

func (rl *RateLimiter) blockIP(ip string) error {
	_, err := rl.dbPool.Exec(context.Background(), "INSERT INTO blocked_ips (ip) VALUES ($1)", ip)

	if err != nil {
		return fmt.Errorf("error blocking IP: %w", err)
	}

	return fmt.Errorf("too many requests from IP address %s", ip)
}

func (rl *RateLimiter) blockToken(token string) error {
	_, err := rl.dbPool.Exec(context.Background(), "INSERT INTO blocked_tokens (token) VALUES ($1)", token)

	if err != nil {
		return fmt.Errorf("error blocking token: %w", err)
	}

	return fmt.Errorf("too many requests from token %s", token)
}

func main() {
	dbPool, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return
	}

	maxRequests, _ := strconv.Atoi(os.Getenv("MAX_REQUESTS"))
	timeWindow, _ := time.ParseDuration(os.Getenv("TIME_WINDOW"))

	rl, err := NewRateLimiter(dbPool, maxRequests, timeWindow)
	if err != nil {
		fmt.Println("Error creating rate limiter:", err)
		return
	}

	validate := validator.New()
	if err := validate.Struct(rl); err != nil {
		fmt.Println("Error validating rate limiter:", err)
		return
	}

	router := gin.Default()
	router.Use(rl.Limit)

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, World!",
		})
	})

	router.Run(":8080")
}
