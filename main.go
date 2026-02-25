package main

import (
	"coffeebase-api/api/handlers"
	adminhandlers "coffeebase-api/api/handlers/admin"
	"coffeebase-api/api/routes"
	"coffeebase-api/internal/database"
	"coffeebase-api/internal/middleware/ratelimit"
	"coffeebase-api/internal/notifications"
	orderservice "coffeebase-api/internal/service/order"
	cartstore "coffeebase-api/internal/store/cart"
	couponstore "coffeebase-api/internal/store/coupon"
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
	db, err := database.NewConnection()
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	if err := database.SeedDatabase(db); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

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
	coStore := couponstore.NewStore(db)

	oService := orderservice.NewService(db, rdb, oStore, cStore, pStore, coStore)

	hub := notifications.NewHub()

	ah := &handlers.AuthHandler{Store: uStore}
	ph := &handlers.ProductHandler{Store: pStore}
	oh := &handlers.OrderHandler{Store: oStore, ProductStore: pStore, Service: oService}
	rh := &handlers.ReviewHandler{Store: rStore}
	fh := &handlers.FavoriteHandler{Store: fStore}
	uh := &handlers.UserHandler{Store: uStore}
	ch := &handlers.CartHandler{Store: cStore}
	nh := &handlers.NotificationHandler{Hub: hub}

	aph := &adminhandlers.ProductHandler{Store: pStore}
	aoh := &adminhandlers.OrderHandler{Store: oStore, Hub: hub}
	aco := &adminhandlers.CouponHandler{Store: coStore}

	r := routes.NewRouter(ah, ph, oh, rh, fh, uh, ch, aph, aoh, nh, aco)

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
