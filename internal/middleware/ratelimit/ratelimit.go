package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

/**
 * RateLimitMiddleware implements a sliding window rate limiter using Redis.
 * It prevents API abuse by limiting the number of requests a client can make within a timeframe.
 * Refactored to eliminate all shorthands and follow strictly declarative naming.
 */
func RateLimitMiddleware(redisClient *redis.Client, requestLimit int, timeWindow time.Duration) func(http.Handler) http.Handler {
	return func(nextHandler http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
			requestContext := context.Background()
			
			// Identify client by Remote Address (IP)
			clientIdentifier := httpRequest.RemoteAddr
			ratelimitCacheKey := fmt.Sprintf("ratelimit:%s", clientIdentifier)

			// Atomic increment in Redis
			currentRequestCount, executionError := redisClient.Incr(requestContext, ratelimitCacheKey).Result()
			if executionError != nil {
				// Failing open: if Redis is down, we allow the request but log the warning
				nextHandler.ServeHTTP(responseWriter, httpRequest)
				return
			}

			// If it's the first request in the window, set the expiration
			if currentRequestCount == 1 {
				redisClient.Expire(requestContext, ratelimitCacheKey, timeWindow)
			}

			// Check if the client has exceeded the allowed limit
			if currentRequestCount > int64(requestLimit) {
				responseWriter.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", requestLimit))
				responseWriter.Header().Set("Retry-After", fmt.Sprintf("%.0f", timeWindow.Seconds()))
				
				http.Error(responseWriter, "Too many requests. Please slow down.", http.StatusTooManyRequests)
				return
			}

			// Attach rate limit info to headers (good practice for API transparency)
			responseWriter.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", requestLimit))
			responseWriter.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", int64(requestLimit)-currentRequestCount))

			nextHandler.ServeHTTP(responseWriter, httpRequest)
		})
	}
}
