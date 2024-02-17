package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

func main() {
	redisClient := connectToRedis()
	defer redisClient.Close()

	os.Setenv("IP_LIMIT", "5")

	ipLimitStr := os.Getenv("IP_LIMIT")
	ipLimit, err := strconv.Atoi(ipLimitStr)
	if err != nil {
		fmt.Println("Erro ao ler IP_LIMIT:", err)
		return
	}

	os.Setenv("TOKEN_LIMIT", "10")

	tokenLimitStr := os.Getenv("TOKEN_LIMIT")
	tokenLimit, err := strconv.Atoi(tokenLimitStr)
	if err != nil {
		fmt.Println("Erro ao ler TOKEN_LIMIT:", err)
		return
	}

	os.Setenv("DURATION", "1s")

	durationStr := os.Getenv("DURATION")
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		fmt.Println("Erro ao ler DURATION:", err)
		return
	}

	rl := NewRateLimiter(redisClient)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})

	http.Handle("/", RateLimitMiddleware(rl, ipLimit, tokenLimit, duration)(mux))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server started at port", port)
	http.ListenAndServe(":"+port, nil)
}

var ctx = context.Background()

func connectToRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379", // Endereço do servidor Redis
		Password: "",           // Senha (se necessário)
		DB:       0,            // Número do banco de dados
	})

	// Testa a conexão com o Redis
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		fmt.Println("Erro ao conectar ao Redis:", err)
	} else {
		fmt.Println("Conexão estabelecida com o Redis:", pong)
	}

	return rdb
}

type RateLimiter struct {
	redisClient *redis.Client
}

func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
	}
}

func (rl *RateLimiter) LimitByIP(ip string, limit int, duration time.Duration) bool {
	key := fmt.Sprintf("ip:%s", ip)
	count, err := rl.redisClient.Incr(ctx, key).Result()
	if err != nil {
		fmt.Println("Erro ao incrementar contador do IP:", err)
		return false
	}

	if count == 1 {
		rl.redisClient.Expire(ctx, key, duration)
	}

	return count <= int64(limit)
}

func (rl *RateLimiter) LimitByToken(token string, limit int, duration time.Duration) bool {
	key := fmt.Sprintf("token:%s", token)
	count, err := rl.redisClient.Incr(ctx, key).Result()
	if err != nil {
		fmt.Println("Erro ao incrementar contador do token:", err)
		return false
	}

	if count == 1 {
		rl.redisClient.Expire(ctx, key, duration)
	}

	return count <= int64(limit)
}

func RateLimitMiddleware(rl *RateLimiter, ipLimit int, tokenLimit int, duration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			token := r.Header.Get("API_KEY")

			if rl.LimitByToken(token, tokenLimit, duration) {
				if rl.LimitByIP(ip, ipLimit, duration) {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "You have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
		})
	}
}
