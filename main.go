package main

import (
	"coffeebase-api/api/handlers"
	adminhandlers "coffeebase-api/api/handlers/admin"
	"coffeebase-api/api/routes"
	"coffeebase-api/internal/database"
	"coffeebase-api/internal/notifications"
	orderservice "coffeebase-api/internal/service/order"
	cartstore "coffeebase-api/internal/store/cart"
	couponstore "coffeebase-api/internal/store/coupon"
	favoritestore "coffeebase-api/internal/store/favorite"
	orderstore "coffeebase-api/internal/store/order"
	productstore "coffeebase-api/internal/store/product"
	reviewstore "coffeebase-api/internal/store/review"
	userstore "coffeebase-api/internal/store/user"
	"database/sql"
	output "fmt"
	systemLog "log"
	webServer "net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Load environment variables from .env file and override system variables
	if err := godotenv.Overload(); err != nil {
		systemLog.Println("Info: No .env file found, using system environment variables")
	}

	databaseConnection := initializeDatabaseAndRunMigrations()
	defer databaseConnection.Close()

	redisClient := initializeRedisConnection()
	if redisClient != nil {
		defer redisClient.Close()
	}

	router := buildApplicationRouter(databaseConnection, redisClient)

	startApiServer(router)
}

func initializeDatabaseAndRunMigrations() *sql.DB {
	databaseConnection, databaseConnectionError := database.NewConnection()
	if databaseConnectionError != nil {
		systemLog.Fatalf("Could not connect to database: %v", databaseConnectionError)
	}

	migrationError := database.RunMigrations(databaseConnection)
	if migrationError != nil {
		systemLog.Fatalf("Failed to run migrations: %v", migrationError)
	}

	seedingError := database.SeedDatabase(databaseConnection)
	if seedingError != nil {
		systemLog.Fatalf("Failed to seed database: %v", seedingError)
	}

	return databaseConnection
}

func initializeRedisConnection() *redis.Client {
	redisClient, redisConnectionError := database.NewRedisClient()
	if redisConnectionError != nil {
		systemLog.Printf("Warning: Could not connect to redis: %v. Caching will be disabled.", redisConnectionError)
		return nil
	}
	return redisClient
}

func buildApplicationRouter(databaseConnection *sql.DB, redisClient *redis.Client) *chi.Mux {
	// Initialize Stores
	userStoreInstance := userstore.NewStore(databaseConnection)
	productStoreInstance := productstore.NewStore(databaseConnection, redisClient)
	orderStoreInstance := orderstore.NewStore(databaseConnection)
	reviewStoreInstance := reviewstore.NewStore(databaseConnection)
	favoriteStoreInstance := favoritestore.NewStore(databaseConnection)
	cartStoreInstance := cartstore.NewStore(databaseConnection, redisClient)
	couponStoreInstance := couponstore.NewStore(databaseConnection)

	// Initialize Business Services
	orderBusinessService := orderservice.NewService(databaseConnection, redisClient, orderStoreInstance, cartStoreInstance, productStoreInstance, couponStoreInstance)
	notificationHub := notifications.NewHub()

	// Initialize Handlers
	authHandler := &handlers.AuthHandler{UserStore: userStoreInstance, NotificationHub: notificationHub}
	productHandler := &handlers.ProductHandler{ProductStore: productStoreInstance}
	orderHandler := &handlers.OrderHandler{OrderStore: orderStoreInstance, ProductStore: productStoreInstance, OrderService: orderBusinessService}
	reviewHandler := &handlers.ReviewHandler{ReviewStore: reviewStoreInstance}
	favoriteHandler := &handlers.FavoriteHandler{FavoriteStore: favoriteStoreInstance}
	userHandler := &handlers.UserHandler{UserStore: userStoreInstance}
	cartHandler := &handlers.CartHandler{CartStore: cartStoreInstance}
	notificationHandler := &handlers.NotificationHandler{NotificationHub: notificationHub}

	adminProductHandler := &adminhandlers.ProductHandler{ProductStore: productStoreInstance}
	adminOrderHandler := &adminhandlers.OrderHandler{OrderStore: orderStoreInstance, NotificationHub: notificationHub}
	adminUserHandler := &adminhandlers.UserHandler{UserStore: userStoreInstance}
	adminCouponHandler := &adminhandlers.CouponHandler{CouponStore: couponStoreInstance}
	adminDashboardHandler := &adminhandlers.DashboardHandler{OrderStore: orderStoreInstance, UserStore: userStoreInstance}

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
		redisClient,
	)
}

func startApiServer(applicationRouter *chi.Mux) {
	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	output.Printf("Coffee Shop API starting on port %s...\n", serverPort)
	serverRunError := webServer.ListenAndServe(":"+serverPort, applicationRouter)
	if serverRunError != nil {
		systemLog.Fatalf("Failed to start server: %v", serverRunError)
	}
}
