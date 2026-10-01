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

type RouterBuilder struct {
	authHandler            *handlers.AuthHandler
	productHandler         *handlers.ProductHandler
	orderHandler           *handlers.OrderHandler
	reviewHandler          *handlers.ReviewHandler
	favoriteHandler        *handlers.FavoriteHandler
	userHandler            *handlers.UserHandler
	cartHandler            *handlers.CartHandler
	attendanceHandler      *handlers.AttendanceHandler
	adminProductHandler    *adminhandlers.ProductHandler
	adminOrderHandler      *adminhandlers.OrderHandler
	adminUserHandler       *adminhandlers.UserHandler
	notificationHandler    *handlers.NotificationHandler
	adminCouponHandler     *adminhandlers.CouponHandler
	adminDashboardHandler  *adminhandlers.DashboardHandler
	adminAttendanceHandler *adminhandlers.AdminAttendanceHandler
	billingHandler         *handlers.BillingHandler
}

// --- Public ---

func NewRouter(
	authHandler *handlers.AuthHandler,
	productHandler *handlers.ProductHandler,
	orderHandler *handlers.OrderHandler,
	reviewHandler *handlers.ReviewHandler,
	favoriteHandler *handlers.FavoriteHandler,
	userHandler *handlers.UserHandler,
	cartHandler *handlers.CartHandler,
	attendanceHandler *handlers.AttendanceHandler,
	adminProductHandler *adminhandlers.ProductHandler,
	adminOrderHandler *adminhandlers.OrderHandler,
	adminUserHandler *adminhandlers.UserHandler,
	notificationHandler *handlers.NotificationHandler,
	adminCouponHandler *adminhandlers.CouponHandler,
	adminDashboardHandler *adminhandlers.DashboardHandler,
	adminAttendanceHandler *adminhandlers.AdminAttendanceHandler,
	billingHandler *handlers.BillingHandler,
	cacheService cache.Service,
) *chi.Mux {
	builder := &RouterBuilder{
		authHandler:            authHandler,
		productHandler:         productHandler,
		orderHandler:           orderHandler,
		reviewHandler:          reviewHandler,
		favoriteHandler:        favoriteHandler,
		userHandler:            userHandler,
		cartHandler:            cartHandler,
		attendanceHandler:      attendanceHandler,
		adminProductHandler:    adminProductHandler,
		adminOrderHandler:      adminOrderHandler,
		adminUserHandler:       adminUserHandler,
		notificationHandler:    notificationHandler,
		adminCouponHandler:     adminCouponHandler,
		adminDashboardHandler:  adminDashboardHandler,
		adminAttendanceHandler: adminAttendanceHandler,
		billingHandler:         billingHandler,
	}

	applicationRouter := chi.NewRouter()

	applicationRouter.Use(custom_middleware.SecurityHeaders)
	applicationRouter.Use(custom_middleware.AuditMiddleware) // OWASP A09
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

	applicationRouter.Route("/api/v1", builder.registerAPIV1)

	return applicationRouter
}

// --- RouterBuilder Methods (sin nested functions) ---

func (b *RouterBuilder) registerAPIV1(apiV1Router chi.Router) {
	// Rutas públicas de auth
	apiV1Router.Post("/users", b.authHandler.Register)
	apiV1Router.Post("/tokens", b.authHandler.Login)
	apiV1Router.Post("/tokens/refresh", b.authHandler.Refresh)
	apiV1Router.Post("/logout", b.authHandler.Logout)

	// Catálogo público
	registerProductRoutes(apiV1Router, b.productHandler)

	// WebSocket con auth
	apiV1Router.Group(b.registerWebSocketRoutes)

	// Rutas protegidas (empleados y clientes)
	apiV1Router.Group(b.registerProtectedRoutes)
}

func (b *RouterBuilder) registerWebSocketRoutes(wsRouter chi.Router) {
	wsRouter.Use(custom_middleware.AuthMiddleware)
	wsRouter.Get("/notifications/ws", b.notificationHandler.HandleWS)
}

func (b *RouterBuilder) registerProtectedRoutes(protectedRouter chi.Router) {
	protectedRouter.Use(custom_middleware.AuthMiddleware)

	registerUserRoutes(protectedRouter, b.userHandler)
	registerOrderRoutes(protectedRouter, b.orderHandler)
	registerCartRoutes(protectedRouter, b.cartHandler)
	registerReviewRoutes(protectedRouter, b.reviewHandler)
	registerFavoriteRoutes(protectedRouter, b.favoriteHandler)
	registerAttendanceRoutes(protectedRouter, b.attendanceHandler)

	protectedRouter.Route("/billing", b.registerBillingRoutes)

	// Rutas de administración
	protectedRouter.Group(b.registerAdminRoutesGroup)
}

func (b *RouterBuilder) registerBillingRoutes(billingRouter chi.Router) {
	billingRouter.Get("/wallet", b.billingHandler.GetWallet)
	billingRouter.Get("/payment-methods", b.billingHandler.GetPaymentMethods)
	billingRouter.Post("/payment-methods", b.billingHandler.AddPaymentMethod)
}

func (b *RouterBuilder) registerAdminRoutesGroup(adminRouter chi.Router) {
	adminRouter.Use(custom_middleware.AdminMiddleware)
	registerAdminRoutes(
		adminRouter,
		b.adminProductHandler,
		b.adminOrderHandler,
		b.adminUserHandler,
		b.adminCouponHandler,
		b.adminDashboardHandler,
		b.adminAttendanceHandler,
	)
}

// --- Specific Route Registrars ---

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

func registerAttendanceRoutes(router chi.Router, attendanceHandler *handlers.AttendanceHandler) {
	router.Post("/attendance/check-in", attendanceHandler.CheckIn)
	router.Post("/attendance/check-out", attendanceHandler.CheckOut)
	router.Get("/attendance/me", attendanceHandler.GetMyAttendance)
}

func registerAdminRoutes(
	router chi.Router,
	productHandler *adminhandlers.ProductHandler,
	orderHandler *adminhandlers.OrderHandler,
	userHandler *adminhandlers.UserHandler,
	couponHandler *adminhandlers.CouponHandler,
	dashboardHandler *adminhandlers.DashboardHandler,
	attendanceHandler *adminhandlers.AdminAttendanceHandler,
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

	// Trazabilidad de asistencia via QR
	router.Post("/admin/attendance/qr/generate", attendanceHandler.GenerateQR)
	router.Get("/admin/attendance", attendanceHandler.GetAll)
	router.Get("/admin/attendance/{user_id}", attendanceHandler.GetByUserID)
}

type staticFileServerHandler struct {
	rootDirectory http.FileSystem
}

func (h *staticFileServerHandler) ServeHTTP(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	if strings.Contains(httpRequest.URL.Path, "..") {
		http.Error(responseWriter, "Invalid path", http.StatusBadRequest)
		return
	}

	routeContext := chi.RouteContext(httpRequest.Context())
	pathPrefix := strings.TrimSuffix(routeContext.RoutePattern(), "/*")
	fileServerInstance := http.StripPrefix(pathPrefix, http.FileServer(h.rootDirectory))
	fileServerInstance.ServeHTTP(responseWriter, httpRequest)
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

	fileHandler := &staticFileServerHandler{rootDirectory: rootDirectory}
	applicationRouter.Method(http.MethodGet, urlPath, fileHandler)
}
