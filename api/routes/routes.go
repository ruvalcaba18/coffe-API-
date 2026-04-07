package routes

import (
	"coffeebase-api/api/handlers"
	adminhandlers "coffeebase-api/api/handlers/admin"
	custom_middleware "coffeebase-api/internal/middleware"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"coffeebase-api/internal/cache"
	"coffeebase-api/internal/middleware/ratelimit"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// --- Public ---

func NewRouter(
	authHandler *handlers.AuthHandler,
	productHandler *handlers.ProductHandler,
	orderHandler *handlers.OrderHandler,
	reviewHandler *handlers.ReviewHandler,
	favoriteHandler *handlers.FavoriteHandler,
	userHandler *handlers.UserHandler,
	cartHandler *handlers.CartHandler,
	adminProductHandler *adminhandlers.ProductHandler,
	adminOrderHandler *adminhandlers.OrderHandler,
	adminUserHandler *adminhandlers.UserHandler,
	notificationHandler *handlers.NotificationHandler,
	adminCouponHandler *adminhandlers.CouponHandler,
	adminDashboardHandler *adminhandlers.DashboardHandler,
	billingHandler *handlers.BillingHandler,
	cacheService cache.Service,
) *chi.Mux {
	applicationRouter := chi.NewRouter()

	// OWASP A05 - Security headers on every response
	applicationRouter.Use(custom_middleware.SecurityHeaders)

	// OWASP A04 - Limit request body to 4MB to prevent resource exhaustion
	applicationRouter.Use(middleware.RequestSize(4 * 1024 * 1024))

	allowedOriginsString := os.Getenv("ALLOWED_ORIGINS")
	allowedOrigins := []string{"http://localhost:3000", "http://localhost:5173"} 
	if allowedOriginsString != "" {
		parts := strings.Split(allowedOriginsString, ",")
		var cleanedOrigins []string
		for _, part := range parts {
			cleanedOrigins = append(cleanedOrigins, strings.TrimSpace(part))
		}
		allowedOrigins = cleanedOrigins
	}

	applicationRouter.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	if cacheService != nil {
		applicationRouter.Use(ratelimit.RateLimitMiddleware(cacheService, 60, time.Minute))
	}

	applicationRouter.Use(middleware.Logger)
	applicationRouter.Use(middleware.Recoverer)
	applicationRouter.Use(middleware.Timeout(60 * time.Second))

	workingDirectory, workingDirectoryError := os.Getwd()
	if workingDirectoryError == nil {
		filesDirectory := http.Dir(filepath.Join(workingDirectory, "uploads"))
		setupFileServer(applicationRouter, "/uploads", filesDirectory)
	}

	applicationRouter.Route("/api/v1", func(apiV1Router chi.Router) {
		apiV1Router.Post("/users", authHandler.Register)
		apiV1Router.Post("/tokens", authHandler.Login)
		registerProductRoutes(apiV1Router, productHandler)
		apiV1Router.Group(func(wsRouter chi.Router) {
			wsRouter.Use(custom_middleware.AuthMiddleware)
			wsRouter.Get("/notifications/ws", notificationHandler.HandleWS)
		})

		apiV1Router.Group(func(protectedRouter chi.Router) {
			protectedRouter.Use(custom_middleware.AuthMiddleware)

			registerUserRoutes(protectedRouter, userHandler)
			registerOrderRoutes(protectedRouter, orderHandler)
			registerCartRoutes(protectedRouter, cartHandler)
			registerReviewRoutes(protectedRouter, reviewHandler)
			registerFavoriteRoutes(protectedRouter, favoriteHandler)

			protectedRouter.Route("/billing", func(billingRouter chi.Router) {
				billingRouter.Get("/wallet", billingHandler.GetWallet)
				billingRouter.Get("/payment-methods", billingHandler.GetPaymentMethods)
				billingRouter.Post("/payment-methods", billingHandler.AddPaymentMethod)
			})

			protectedRouter.Group(func(adminRouter chi.Router) {
				adminRouter.Use(custom_middleware.AdminMiddleware)
				registerAdminRoutes(adminRouter, adminProductHandler, adminOrderHandler, adminUserHandler, adminCouponHandler, adminDashboardHandler)
			})
		})
	})

	return applicationRouter
}

// --- Private ---

func registerUserRoutes(router chi.Router, userHandler *handlers.UserHandler) {
	router.Get("/profile", userHandler.GetProfile)
	router.Patch("/profile", userHandler.UpdateProfile)
	router.Post("/profile/avatar", userHandler.UploadAvatar)
}

func registerProductRoutes(router chi.Router, productHandler *handlers.ProductHandler) {
	router.Get("/products", productHandler.GetAll)
	router.Get("/products/categories", productHandler.GetCategories)
	router.Get("/products/{id}", productHandler.GetByID)
}

func registerOrderRoutes(router chi.Router, orderHandler *handlers.OrderHandler) {
	router.Post("/orders", orderHandler.Checkout)
	router.Get("/orders", orderHandler.GetHistory)
	router.Get("/orders/latest", orderHandler.GetLatest)
	router.Get("/orders/pickups", orderHandler.GetPickups)
}

func registerCartRoutes(router chi.Router, cartHandler *handlers.CartHandler) {
	router.Get("/cart", cartHandler.GetCart)
	router.Patch("/cart", cartHandler.UpdateItem)
}

func registerReviewRoutes(router chi.Router, reviewHandler *handlers.ReviewHandler) {
	router.Post("/reviews", reviewHandler.Create)
	router.Get("/reviews", reviewHandler.GetByProduct)
}

func registerFavoriteRoutes(router chi.Router, favoriteHandler *handlers.FavoriteHandler) {
	router.Get("/favorites", favoriteHandler.GetUserFavorites)
	router.Post("/favorites", favoriteHandler.Add)
	router.Delete("/favorites/{id}", favoriteHandler.Remove)
}

func registerAdminRoutes(
	router chi.Router,
	productHandler *adminhandlers.ProductHandler,
	orderHandler *adminhandlers.OrderHandler,
	userHandler *adminhandlers.UserHandler,
	couponHandler *adminhandlers.CouponHandler,
	dashboardHandler *adminhandlers.DashboardHandler,
) {
	router.Post("/admin/products", productHandler.Create)
	router.Post("/admin/products/bulk", productHandler.CreateBulk)
	router.Put("/admin/products/{id}", productHandler.Update)
	router.Delete("/admin/products/{id}", productHandler.Delete)

	router.Get("/admin/orders", orderHandler.GetAll)
	router.Patch("/admin/orders/{id}", orderHandler.UpdateStatus)

	router.Get("/admin/users", userHandler.GetAll)
	router.Patch("/admin/users/role/{id}", userHandler.UpdateRole)
	router.With(custom_middleware.SuperAdminMiddleware).Delete("/admin/users/{id}", userHandler.Delete)

	router.Post("/admin/coupons", couponHandler.Create)
	router.Get("/admin/coupons", couponHandler.GetAll)
	router.Patch("/admin/coupons/status/{id}", couponHandler.ToggleStatus)
	router.Delete("/admin/coupons/{id}", couponHandler.Delete)

	router.Get("/admin/dashboard/stats", dashboardHandler.GetStats)
}

func setupFileServer(applicationRouter chi.Router, urlPath string, rootDirectory http.FileSystem) {
	if strings.ContainsAny(urlPath, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if urlPath != "/" && urlPath[len(urlPath)-1] != '/' {
		applicationRouter.Get(urlPath, http.RedirectHandler(urlPath+"/", http.StatusMovedPermanently).ServeHTTP)
		urlPath += "/"
	}
	urlPath += "*"

	applicationRouter.Get(urlPath, func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
		if strings.Contains(httpRequest.URL.Path, "..") {
			http.Error(responseWriter, "Invalid path", http.StatusBadRequest)
			return
		}

		routeContext := chi.RouteContext(httpRequest.Context())
		pathPrefix := strings.TrimSuffix(routeContext.RoutePattern(), "/*")
		fileServerInstance := http.StripPrefix(pathPrefix, http.FileServer(rootDirectory))
		fileServerInstance.ServeHTTP(responseWriter, httpRequest)
	})
}
