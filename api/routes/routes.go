package routes

import (
	"coffeebase-api/api/handlers"
	adminhandlers "coffeebase-api/api/handlers/admin"
	custom_middleware "coffeebase-api/internal/middleware"
	webServer "net/http"
	"os"
	"path/filepath"
	stringManipulation "strings"

	"coffeebase-api/internal/middleware/ratelimit"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/redis/go-redis/v9"
)

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
	redisClient *redis.Client,
) *chi.Mux {
	router := chi.NewRouter()

	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	allowedOrigins := []string{"http://localhost:3000", "http://localhost:5173"} // Valores por defecto seguros
	if allowedOriginsStr != "" {
		allowedOrigins = stringManipulation.Split(allowedOriginsStr, ",")
	}

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	if redisClient != nil {
		router.Use(ratelimit.RateLimitMiddleware(redisClient, 60, time.Minute))
	}

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	workDirectory, workingDirectoryError := os.Getwd()
	if workingDirectoryError == nil {
		filesDirectory := webServer.Dir(filepath.Join(workDirectory, "uploads"))
		FileServer(router, "/uploads", filesDirectory)
	}

	router.Route("/api/v1", func(apiV1Router chi.Router) {
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
			})

			protectedRouter.Group(func(adminRouter chi.Router) {
				adminRouter.Use(custom_middleware.AdminMiddleware)
				registerAdminRoutes(adminRouter, adminProductHandler, adminOrderHandler, adminUserHandler, adminCouponHandler, adminDashboardHandler)
			})
		})
	})

	return router
}

func registerUserRoutes(router chi.Router, handler *handlers.UserHandler) {
	router.Get("/profile", handler.GetProfile)
	router.Patch("/profile", handler.UpdateProfile)
	router.Post("/profile/avatar", handler.UploadAvatar)
}

func registerProductRoutes(router chi.Router, handler *handlers.ProductHandler) {
	router.Get("/products", handler.GetAll)
	router.Get("/products/categories", handler.GetCategories)
	router.Get("/products/{id}", handler.GetByID)
}

func registerOrderRoutes(router chi.Router, handler *handlers.OrderHandler) {
	router.Post("/orders", handler.Checkout)
	router.Get("/orders", handler.GetHistory)
	router.Get("/orders/latest", handler.GetLatest)
	router.Get("/orders/pickups", handler.GetPickups)
}

func registerCartRoutes(router chi.Router, handler *handlers.CartHandler) {
	router.Get("/cart", handler.GetCart)
	router.Patch("/cart", handler.UpdateItem)
}

func registerReviewRoutes(router chi.Router, handler *handlers.ReviewHandler) {
	router.Post("/reviews", handler.Create)
	router.Get("/reviews", handler.GetByProduct)
}

func registerFavoriteRoutes(router chi.Router, handler *handlers.FavoriteHandler) {
	router.Get("/favorites", handler.GetUserFavorites)
	router.Post("/favorites", handler.Add)
	router.Delete("/favorites/{id}", handler.Remove)
}

func registerAdminRoutes(
	router chi.Router,
	productH *adminhandlers.ProductHandler,
	orderH *adminhandlers.OrderHandler,
	userH *adminhandlers.UserHandler,
	couponH *adminhandlers.CouponHandler,
	dashboardH *adminhandlers.DashboardHandler,
) {
	router.Post("/admin/products", productH.Create)
	router.Post("/admin/products/bulk", productH.CreateBulk)
	router.Put("/admin/products/{id}", productH.Update)
	router.Delete("/admin/products/{id}", productH.Delete)

	router.Get("/admin/orders", orderH.GetAll)
	router.Patch("/admin/orders/{id}", orderH.UpdateStatus)

	router.Get("/admin/users", userH.GetAll)
	router.Patch("/admin/users/role/{id}", userH.UpdateRole)

	router.Post("/admin/coupons", couponH.Create)
	router.Get("/admin/coupons", couponH.GetAll)
	router.Patch("/admin/coupons/status/{id}", couponH.ToggleStatus)
	router.Delete("/admin/coupons/{id}", couponH.Delete)

	router.Get("/admin/dashboard/stats", dashboardH.GetStats)
}

func FileServer(router chi.Router, path string, root webServer.FileSystem) {
	if stringManipulation.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		router.Get(path, webServer.RedirectHandler(path+"/", webServer.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	router.Get(path, func(responseWriter webServer.ResponseWriter, httpRequest *webServer.Request) {
		if stringManipulation.Contains(httpRequest.URL.Path, "..") {
			webServer.Error(responseWriter, "Invalid path", webServer.StatusBadRequest)
			return
		}

		routeContext := chi.RouteContext(httpRequest.Context())
		pathPrefix := stringManipulation.TrimSuffix(routeContext.RoutePattern(), "/*")
		fileServerInstance := webServer.StripPrefix(pathPrefix, webServer.FileServer(root))
		fileServerInstance.ServeHTTP(responseWriter, httpRequest)
	})
}
