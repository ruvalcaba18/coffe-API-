package main

import (
	"coffeebase-api/api/handlers"
	adminhandlers "coffeebase-api/api/handlers/admin"
	"coffeebase-api/api/routes"
	"coffeebase-api/internal/cache"
	"coffeebase-api/internal/database"
	"coffeebase-api/internal/middleware"
	"coffeebase-api/internal/notifications"
	orderservice "coffeebase-api/internal/service/order"
	billingstore "coffeebase-api/internal/store/billing"
	cartstore "coffeebase-api/internal/store/cart"
	couponstore "coffeebase-api/internal/store/coupon"
	favoritestore "coffeebase-api/internal/store/favorite"
	orderstore "coffeebase-api/internal/store/order"
	productstore "coffeebase-api/internal/store/product"
	reviewstore "coffeebase-api/internal/store/review"
	userstore "coffeebase-api/internal/store/user"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

// --- Public ---

func main() {
	if error := godotenv.Overload(); error != nil {
		log.Println("Info: No .env file found, using system environment variables")
	}

	// OWASP A02 — Validate JWT secret before starting
	middleware.ValidateJWTSecret()

	databaseConnection := initializeDatabaseAndRunMigrations()
	defer databaseConnection.Close()

	redisClient := initializeRedisConnection()
	var cacheService cache.Service
	if redisClient != nil {
		defer redisClient.Close()
		cacheService = cache.NewRedisCache(redisClient)
	}

	applicationRouter := buildApplicationRouter(databaseConnection, cacheService)

	startApiServer(applicationRouter)
}

// --- Private ---

func initializeDatabaseAndRunMigrations() *sql.DB {
	databaseConnection, databaseConnectionError := database.NewConnection()
	if databaseConnectionError != nil {
		log.Fatalf("Could not connect to database: %v", databaseConnectionError)
	}

	migrationError := database.RunMigrations(databaseConnection)
	if migrationError != nil {
		log.Fatalf("Failed to run migrations: %v", migrationError)
	}

	seedingError := database.SeedDatabase(databaseConnection)
	if seedingError != nil {
		log.Fatalf("Failed to seed database: %v", seedingError)
	}

	return databaseConnection
}

func initializeRedisConnection() *redis.Client {
	redisClient, redisConnectionError := database.NewRedisClient()
	if redisConnectionError != nil {
		log.Printf("Warning: Could not connect to redis: %v. Caching will be disabled.", redisConnectionError)
		return nil
	}
	return redisClient
}

func buildApplicationRouter(databaseConnection *sql.DB, cacheService cache.Service) *chi.Mux {
	userStoreInstance := userstore.NewStore(databaseConnection)
	productStoreInstance := productstore.NewStore(databaseConnection, cacheService)
	orderStoreInstance := orderstore.NewStore(databaseConnection)
	reviewStoreInstance := reviewstore.NewStore(databaseConnection)
	favoriteStoreInstance := favoritestore.NewStore(databaseConnection)
	cartStoreInstance := cartstore.NewStore(databaseConnection, cacheService)
	couponStoreInstance := couponstore.NewStore(databaseConnection)
	billingStoreInstance := billingstore.NewStore(databaseConnection)

	orderBusinessService := orderservice.NewService(databaseConnection, cacheService, orderStoreInstance, cartStoreInstance, productStoreInstance, couponStoreInstance)
	notificationHub := notifications.NewHub()

	authHandler := handlers.NewAuthHandler(userStoreInstance, notificationHub)
	productHandler := handlers.NewProductHandler(productStoreInstance)
	orderHandler := handlers.NewOrderHandler(orderStoreInstance, productStoreInstance, orderBusinessService)
	reviewHandler := handlers.NewReviewHandler(reviewStoreInstance)
	favoriteHandler := handlers.NewFavoriteHandler(favoriteStoreInstance)
	userHandler := handlers.NewUserHandler(userStoreInstance)
	cartHandler := handlers.NewCartHandler(cartStoreInstance)
	notificationHandler := handlers.NewNotificationHandler(notificationHub)
	billingHandler := handlers.NewBillingHandler(billingStoreInstance)

	adminProductHandler := adminhandlers.NewProductHandler(productStoreInstance)
	adminOrderHandler := adminhandlers.NewOrderHandler(orderStoreInstance, notificationHub)
	adminUserHandler := adminhandlers.NewUserHandler(userStoreInstance)
	adminCouponHandler := adminhandlers.NewCouponHandler(couponStoreInstance)
	adminDashboardHandler := adminhandlers.NewDashboardHandler(orderStoreInstance, userStoreInstance, couponStoreInstance)

	return routes.NewRouter(
		authHandler,
		productHandler,
		orderHandler,
		reviewHandler,
		favoriteHandler,
		userHandler,
		cartHandler,
		adminProductHandler,
		adminOrderHandler,
		adminUserHandler,
		notificationHandler,
		adminCouponHandler,
		adminDashboardHandler,
		billingHandler,
		cacheService,
	)
}

func startApiServer(applicationRouter *chi.Mux) {
	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	fmt.Printf("Coffee Shop API starting on port %s...\n", serverPort)
	serverRunError := http.ListenAndServe(":"+serverPort, applicationRouter)
	if serverRunError != nil {
		log.Fatalf("Failed to start server: %v", serverRunError)
	}
}
