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
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

// --- Public ---

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	if error := godotenv.Overload(); error != nil {
		slog.Warn("No .env file found, using system environment variables")
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
		slog.Error("Could not connect to database", "error", databaseConnectionError)
		os.Exit(1)
	}

	migrationError := database.RunMigrations(databaseConnection)
	if migrationError != nil {
		slog.Error("Failed to run migrations", "error", migrationError)
		os.Exit(1)
	}

	seedingError := database.SeedDatabase(databaseConnection)
	if seedingError != nil {
		slog.Error("Failed to seed database", "error", seedingError)
		os.Exit(1)
	}

	return databaseConnection
}

func initializeRedisConnection() *redis.Client {
	redisClient, redisConnectionError := database.NewRedisClient()
	if redisConnectionError != nil {
		slog.Warn("Could not connect to redis. Caching will be disabled.", "error", redisConnectionError)
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

	slog.Info("Coffee Shop API starting", "port", serverPort)
	serverRunError := http.ListenAndServe(":"+serverPort, applicationRouter)
	if serverRunError != nil {
		slog.Error("Failed to start server", "error", serverRunError)
		os.Exit(1)
	}
}
