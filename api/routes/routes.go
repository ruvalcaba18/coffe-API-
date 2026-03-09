package routes

import (
	"coffeebase-api/api/handlers"
	adminhandlers "coffeebase-api/api/handlers/admin"
	custom_middleware "coffeebase-api/internal/middleware"
	webServer "net/http"
	"os"
	"path/filepath"
	stringManipulation "strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Serve static files for uploads
	workDirectory, workingDirectoryError := os.Getwd()
	if workingDirectoryError == nil {
		filesDirectory := webServer.Dir(filepath.Join(workDirectory, "uploads"))
		FileServer(router, "/uploads", filesDirectory)
	}

	router.Route("/api/v1", func(apiV1Router chi.Router) {
		// Public/Session routes
		apiV1Router.Post("/users", authHandler.Register)
		apiV1Router.Post("/tokens", authHandler.Login)

		// WebSocket Notification Route
		apiV1Router.Group(func(wsRouter chi.Router) {
			wsRouter.Use(custom_middleware.AuthMiddleware)
			wsRouter.Get("/notifications/ws", notificationHandler.HandleWS)
		})

		// Protected routes
		apiV1Router.Group(func(protectedRouter chi.Router) {
			protectedRouter.Use(custom_middleware.AuthMiddleware)

			registerUserRoutes(protectedRouter, userHandler)
			registerProductRoutes(protectedRouter, productHandler)
			registerOrderRoutes(protectedRouter, orderHandler)
			registerCartRoutes(protectedRouter, cartHandler)
			registerReviewRoutes(protectedRouter, reviewHandler)
			registerFavoriteRoutes(protectedRouter, favoriteHandler)

			// Admin Section
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
	router.Patch("/profile", handler.UpdateLanguage)
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
	// Admin Products management
	router.Post("/admin/products", productH.Create)
	router.Post("/admin/products/bulk", productH.CreateBulk)
	router.Put("/admin/products/{id}", productH.Update)
	router.Delete("/admin/products/{id}", productH.Delete)

	// Admin Orders management
	router.Get("/admin/orders", orderH.GetAll)
	router.Patch("/admin/orders/{id}", orderH.UpdateStatus)

	// Admin User management
	router.Get("/admin/users", userH.GetAll)
	router.Patch("/admin/users/role/{id}", userH.UpdateRole)

	// Admin Coupons management
	router.Post("/admin/coupons", couponH.Create)
	router.Get("/admin/coupons", couponH.GetAll)
	router.Patch("/admin/coupons/status/{id}", couponH.ToggleStatus)
	router.Delete("/admin/coupons/{id}", couponH.Delete)

	// Admin Dashboard
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
		// Extra security layer: Block directory traversal attempts explicitly
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
