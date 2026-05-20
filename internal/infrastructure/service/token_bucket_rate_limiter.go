package service

import (
	"sync"
	"time"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/service"
	"golang.org/x/time/rate"
)

type ipLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// TokenBucketRateLimiter implementa a interface RateLimiter usando token bucket
// (algoritmo de rate limiting do golang.org/x/time/rate)
type TokenBucketRateLimiter struct {
	mu        sync.Mutex
	limiters  map[string]*ipLimiterEntry
	rps       float64 // requisições por segundo
	burst     int
	cleanupTk *time.Ticker
}

// NewTokenBucketRateLimiter cria um novo rate limiter com token bucket
// rps: requisições por segundo permitidas (ex: 10.0)
// burst: quantidade de requisições que podem estourar o limite (ex: 2)
func NewTokenBucketRateLimiter(rps float64, burst int) service.RateLimiter {
	rl := &TokenBucketRateLimiter{
		limiters:  make(map[string]*ipLimiterEntry),
		rps:       rps,
		burst:     burst,
		cleanupTk: time.NewTicker(time.Minute),
	}

	// Goroutine de limpeza: remove entradas inativas a cada minuto
	go func() {
		for range rl.cleanupTk.C {
			rl.cleanup()
		}
	}()

	return rl
}

// Allow verifica se a requisição é permitida para o userID/IP
func (r *TokenBucketRateLimiter) Allow(userID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.limiters[userID]
	if !exists {
		entry = &ipLimiterEntry{
			limiter:  rate.NewLimiter(rate.Limit(r.rps), r.burst),
			lastSeen: time.Now(),
		}
		r.limiters[userID] = entry
	}

	entry.lastSeen = time.Now()
	return entry.limiter.Allow()
}

// cleanup remove entradas que não foram vistas há mais de 3 minutos
func (r *TokenBucketRateLimiter) cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for userID, entry := range r.limiters {
		if now.Sub(entry.lastSeen) > 3*time.Minute {
			delete(r.limiters, userID)
		}
	}
}

// Stop para o rate limiter (limpeza de recursos)
func (r *TokenBucketRateLimiter) Stop() {
	r.cleanupTk.Stop()
}
