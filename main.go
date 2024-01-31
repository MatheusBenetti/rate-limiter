package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

type RateLimiter struct {
	dbPool      *pgxpool.Pool
	maxRequests int
	timeWindow  time.Duration
}

func NewRateLimiter(dbPool *pgxpool.Pool, maxRequests int, timeWindow time.Duration) *RateLimiter {
	return &RateLimiter{
		dbPool:      dbPool,
		maxRequests: maxRequests,
		timeWindow:  timeWindow,
	}
}

func (rl *RateLimiter) Limit(c *gin.Context) {
	ip := c.ClientIP()
	token := c.GetHeader("API_KEY")

	if token != "" {
		rl.limitByToken(token)
	}

	rl.limitByIP(ip)
}

func (rl *RateLimiter) limitByIP(ip string) {
	var requests int
	err := rl.dbPool.QueryRow(context.Background(), "SELECT COUNT(*) FROM requests WHERE ip = $1 AND created_at > NOW() - INTERVAL '1 minute'", ip).Scan(&requests)

	if err != nil {
		fmt.Println("Error counting IP requests:", err)
		return
	}

	if requests > rl.maxRequests {
		rl.blockIP(ip)
	}
}

func (rl *RateLimiter) limitByToken(token string) {
	var requests int
	err := rl.dbPool.QueryRow(context.Background(), "SELECT COUNT(*) FROM requests WHERE token = $1 AND created_at > NOW() - INTERVAL '1 minute'", token).Scan(&requests)

	if err != nil {
		fmt.Println("Error counting token requests:", err)
		return
	}

	if requests > rl.maxRequests {
		rl.blockToken(token)
	}
}

func (rl *RateLimiter) blockIP(ip string) {
	_, err := rl.dbPool.Exec(context.Background(), "INSERT INTO blocked_ips (ip) VALUES ($1)", ip)

	if err != nil {
		fmt.Println("Error blocking IP:", err)
		return
	}
}

func (rl *RateLimiter) blockToken(token string) {
	_, err := rl.dbPool.Exec(context.Background(), "INSERT INTO blocked_tokens (token) VALUES ($1)", token)

	if err != nil {
		fmt.Println("Error blocking token:", err)
		return
	}
}

func main() {
	dbPool, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return
	}

	maxRequests, _ := strconv.Atoi(os.Getenv("MAX_REQUESTS"))
	timeWindow, _ := time.ParseDuration(os.Getenv("TIME_WINDOW"))

	rateLimiter := NewRateLimiter(dbPool, maxRequests, timeWindow)

	router := gin.Default()
	router.Use(rateLimiter.Limit)

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, World!",
		})
	})

	router.Run(":8080")
}
