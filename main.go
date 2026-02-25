package main

import (
	"coffeebase-api/api/handlers"
	"coffeebase-api/api/routes"
	"coffeebase-api/internal/database"
	"coffeebase-api/internal/middleware/ratelimit"
	orderservice "coffeebase-api/internal/service/order"
	cartstore "coffeebase-api/internal/store/cart"
	favoritestore "coffeebase-api/internal/store/favorite"
	orderstore "coffeebase-api/internal/store/order"
	productstore "coffeebase-api/internal/store/product"
	reviewstore "coffeebase-api/internal/store/review"
	userstore "coffeebase-api/internal/store/user"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	// Initialize database
	db, err := database.NewConnection()
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer db.Close()

	// Run Migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Redis
	rdb, err := database.NewRedisClient()
	if err != nil {
		log.Printf("Warning: Could not connect to redis: %v. Caching will be disabled.", err)
	} else {
		defer rdb.Close()
	}

	// Initialize stores
	uStore := userstore.NewStore(db)
	pStore := productstore.NewStore(db, rdb)
	oStore := orderstore.NewStore(db)
	rStore := reviewstore.NewStore(db)
	fStore := favoritestore.NewStore(db)
	cStore := cartstore.NewStore(db, rdb)

	// Initialize Services
	oService := orderservice.NewService(db, rdb, oStore, cStore, pStore)

	// Initialize handlers
	ah := &handlers.AuthHandler{Store: uStore}
	ph := &handlers.ProductHandler{Store: pStore}
	oh := &handlers.OrderHandler{Store: oStore, ProductStore: pStore, CartStore: cStore, Service: oService}
	rh := &handlers.ReviewHandler{Store: rStore}
	fh := &handlers.FavoriteHandler{Store: fStore}
	uh := &handlers.UserHandler{Store: uStore}
	ch := &handlers.CartHandler{Store: cStore}

	r := routes.NewRouter(ah, ph, oh, rh, fh, uh, ch)

	if rdb != nil {
		r.Use(ratelimit.RateLimitMiddleware(rdb, 60, time.Minute))
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Coffee Shop API starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
