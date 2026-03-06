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
			
			// Account/Profile
			protectedRouter.Get("/profile", userHandler.GetProfile)
			protectedRouter.Patch("/profile", userHandler.UpdateLanguage)
			protectedRouter.Post("/profile/avatar", userHandler.UploadAvatar)

			// Products
			protectedRouter.Get("/products", productHandler.GetAll)
			protectedRouter.Get("/products/categories", productHandler.GetCategories)
			protectedRouter.Get("/products/{id}", productHandler.GetByID)
			
			// Orders
			protectedRouter.Post("/orders", orderHandler.Checkout)
			protectedRouter.Get("/orders", orderHandler.GetHistory)
			
			// Shopping Cart
			protectedRouter.Get("/cart", cartHandler.GetCart)
			protectedRouter.Patch("/cart", cartHandler.UpdateItem)
			
			// Reviews
			protectedRouter.Post("/reviews", reviewHandler.Create)
			protectedRouter.Get("/reviews", reviewHandler.GetByProduct)
			
			// Favorites
			protectedRouter.Get("/favorites", favoriteHandler.GetUserFavorites)
			protectedRouter.Post("/favorites", favoriteHandler.Add)
			protectedRouter.Delete("/favorites/{id}", favoriteHandler.Remove)

			// Admin Section
			protectedRouter.Group(func(adminRouter chi.Router) {
				adminRouter.Use(custom_middleware.AdminMiddleware)
				
				// Admin Products management
				adminRouter.Post("/admin/products", adminProductHandler.Create)
				adminRouter.Post("/admin/products/bulk", adminProductHandler.CreateBulk)
				adminRouter.Put("/admin/products/{id}", adminProductHandler.Update)
				adminRouter.Delete("/admin/products/{id}", adminProductHandler.Delete)
				
				// Admin Orders management
				adminRouter.Get("/admin/orders", adminOrderHandler.GetAll)
				adminRouter.Patch("/admin/orders/{id}", adminOrderHandler.UpdateStatus)

				// Admin User management
				adminRouter.Get("/admin/users", adminUserHandler.GetAll)
				adminRouter.Patch("/admin/users/role/{id}", adminUserHandler.UpdateRole)

				// Admin Coupons management
				adminRouter.Post("/admin/coupons", adminCouponHandler.Create)
				adminRouter.Get("/admin/coupons", adminCouponHandler.GetAll)
				adminRouter.Patch("/admin/coupons/status/{id}", adminCouponHandler.ToggleStatus)
				adminRouter.Delete("/admin/coupons/{id}", adminCouponHandler.Delete)

				// Admin Dashboard
				adminRouter.Get("/admin/dashboard/stats", adminDashboardHandler.GetStats)
			})
		})
	})

	return router
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
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
		routeContext := chi.RouteContext(httpRequest.Context())
		pathPrefix := stringManipulation.TrimSuffix(routeContext.RoutePattern(), "/*")
		fileServerInstance := webServer.StripPrefix(pathPrefix, webServer.FileServer(root))
		fileServerInstance.ServeHTTP(responseWriter, httpRequest)
	})
}
