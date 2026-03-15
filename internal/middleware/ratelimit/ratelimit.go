package ratelimit

import (
	"coffeebase-api/internal/cache"
	"context"
	"fmt"
	"net/http"
	"time"
)

func RateLimitMiddleware(cacheService cache.Service, requestLimit int, timeWindow time.Duration) func(http.Handler) http.Handler {
	return func(nextHandler http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
			requestContext := context.Background()
			
			clientIdentifier := httpRequest.RemoteAddr
			ratelimitCacheKey := fmt.Sprintf("ratelimit:%s", clientIdentifier)

			currentRequestCount, executionError := cacheService.Incr(requestContext, ratelimitCacheKey)
			if executionError != nil {
				nextHandler.ServeHTTP(responseWriter, httpRequest)
				return
			}

			if currentRequestCount == 1 {
				cacheService.Expire(requestContext, ratelimitCacheKey, timeWindow)
			}

			if currentRequestCount > int64(requestLimit) {
				responseWriter.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", requestLimit))
				responseWriter.Header().Set("Retry-After", fmt.Sprintf("%.0f", timeWindow.Seconds()))
				
				http.Error(responseWriter, "Too many requests. Please slow down.", http.StatusTooManyRequests)
				return
			}

			responseWriter.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", requestLimit))
			responseWriter.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", int64(requestLimit)-currentRequestCount))

			nextHandler.ServeHTTP(responseWriter, httpRequest)
		})
	}
}
