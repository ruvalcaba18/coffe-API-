package ratelimit

import (
	"coffeebase-api/internal/cache"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setupTestCache(t *testing.T) cache.Service {
	miniRedis, miniRedisError := miniredis.Run()
	if miniRedisError != nil {
		t.Fatalf("failed to run miniredis: %s", miniRedisError)
	}
	t.Cleanup(func() { miniRedis.Close() })

	redisClient := redis.NewClient(&redis.Options{
		Addr: miniRedis.Addr(),
	})

	return cache.NewRedisCache(redisClient)
}

func TestRateLimit_AllowsUnderLimit(t *testing.T) {
	cacheService := setupTestCache(t)

	handler := RateLimitMiddleware(cacheService, 5, time.Minute)(
		http.HandlerFunc(func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
			responseWriter.WriteHeader(http.StatusOK)
		}),
	)

	for requestIndex := 0; requestIndex < 5; requestIndex++ {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.RemoteAddr = "192.168.1.1:12345"
		recorder := httptest.NewRecorder()

		handler.ServeHTTP(recorder, request)
		assert.Equal(t, http.StatusOK, recorder.Code, "Request %d should be allowed", requestIndex+1)
	}
}

func TestRateLimit_BlocksOverLimit(t *testing.T) {
	cacheService := setupTestCache(t)

	handler := RateLimitMiddleware(cacheService, 3, time.Minute)(
		http.HandlerFunc(func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
			responseWriter.WriteHeader(http.StatusOK)
		}),
	)

	// First 3 should pass
	for requestIndex := 0; requestIndex < 3; requestIndex++ {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.RemoteAddr = "10.0.0.1:12345"
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		assert.Equal(t, http.StatusOK, recorder.Code)
	}

	// 4th request should be blocked
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.RemoteAddr = "10.0.0.1:12345"
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusTooManyRequests, recorder.Code)
}

func TestRateLimit_DifferentClientsIndependent(t *testing.T) {
	cacheService := setupTestCache(t)

	handler := RateLimitMiddleware(cacheService, 2, time.Minute)(
		http.HandlerFunc(func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
			responseWriter.WriteHeader(http.StatusOK)
		}),
	)

	// Client A uses 2 requests (at limit)
	for requestIndex := 0; requestIndex < 2; requestIndex++ {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.RemoteAddr = "10.0.0.1:12345"
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		assert.Equal(t, http.StatusOK, recorder.Code)
	}

	// Client B should still be allowed
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.RemoteAddr = "10.0.0.2:12345"
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusOK, recorder.Code)
}
