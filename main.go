package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-redis/redis/v8"
)

// RateLimiterStrategy define a interface para armazenar e consultar dados do limiter
type RateLimiterStrategy interface {
	Incr(key string) error
	Get(key string) (int, error)
}

// RedisRateLimiterStrategy implementa a interface RateLimiterStrategy usando Redis
type RedisRateLimiterStrategy struct {
	client *redis.Client
}

// Incr incrementa o valor de uma chave no Redis
func (rls *RedisRateLimiterStrategy) Incr(key string) error {
	ctx := context.Background()
	return rls.client.Incr(ctx, key).Err()
}

// Get retorna o valor de uma chave no Redis
func (rls *RedisRateLimiterStrategy) Get(key string) (int, error) {
	ctx := context.Background()
	val, err := rls.client.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return 0, err
	}
	return val, nil
}

func main() {
	// Conecte-se ao Redis
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()

	// Crie o rate limiter com limites configuráveis e usando a estratégia Redis
	limiter := NewRateLimiter(10, 100, time.Second, time.Minute, &RedisRateLimiterStrategy{client})

	r := chi.NewRouter()

	// Use o middleware do rate limiter
	r.Use(limiter.Handler)

	// Rotas
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!\n"))
	})

	http.ListenAndServe(":8080", r)
}

// RateLimiter define a estrutura do rate limiter
type RateLimiter struct {
	ipLimiter     RateLimiterStrategy
	tokenLimiter  RateLimiterStrategy
	blockDuration time.Duration
}

// NewRateLimiter cria e retorna um novo rate limiter com limites e tempo de bloqueio configuráveis
func NewRateLimiter(ipLimit, tokenLimit int, rateDuration, blockDuration time.Duration, strategy RateLimiterStrategy) *RateLimiter {
	return &RateLimiter{
		ipLimiter:     strategy,
		tokenLimiter:  strategy,
		blockDuration: blockDuration,
	}
}

// Handler é o middleware para o rate limiter
func (rl *RateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Crie uma chave única para o IP ou o token
		var key string
		token := r.Header.Get("API_KEY")
		if token != "" {
			key = fmt.Sprintf("limiter:%s", token)
		} else {
			key = fmt.Sprintf("limiter:%s", r.RemoteAddr)
		}

		// Incrementa o contador para o IP ou o token
		if err := rl.ipLimiter.Incr(key); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Verifica o valor do contador
		val, err := rl.ipLimiter.Get(key)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Verifica se o limite foi excedido
		if val > 10 { // Altere 10 para o limite desejado
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
