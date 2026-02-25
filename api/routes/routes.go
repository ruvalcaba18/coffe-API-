package routes

import (
	"coffeebase-api/api/handlers"
	adminhandlers "coffeebase-api/api/handlers/admin"
	custom_middleware "coffeebase-api/internal/middleware"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	ah *handlers.AuthHandler,
	ph *handlers.ProductHandler,
	oh *handlers.OrderHandler,
	rh *handlers.ReviewHandler,
	fh *handlers.FavoriteHandler,
	uh *handlers.UserHandler,
	ch *handlers.CartHandler,
	aph *adminhandlers.ProductHandler,
	aoh *adminhandlers.OrderHandler,
	nh *handlers.NotificationHandler,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Serve static files for uploads
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "uploads"))
	FileServer(r, "/uploads", filesDir)

	r.Route("/api/v1", func(r chi.Router) {
		// Public/Session routes
		r.Post("/users", ah.Register)   // POST /users is the REST way to register
		r.Post("/tokens", ah.Login)    // POST /tokens is the REST way to login/get token

		// WebSocket Notification Route
		r.Group(func(r chi.Router) {
			r.Use(custom_middleware.AuthMiddleware)
			r.Get("/notifications/ws", nh.HandleWS)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(custom_middleware.AuthMiddleware)
			
			// Account/Profile
			r.Get("/profile", uh.GetProfile)
			r.Patch("/profile", uh.UpdateLanguage) // Using PATCH for partial updates
			r.Post("/profile/avatar", uh.UploadAvatar) // Resource: Avatar

			// Products
			r.Get("/products", ph.GetAll)
			r.Get("/products/{id}", ph.GetByID)
			
			// Orders
			r.Post("/orders", oh.Checkout) // Default POST creates an order from cart
			r.Get("/orders", oh.GetHistory) // Default GET lists your orders
			
			// Shopping Cart
			r.Get("/cart", ch.GetCart)
			r.Patch("/cart", ch.UpdateItem)
			
			// Reviews
			r.Post("/reviews", rh.Create) // product_id in body
			r.Get("/reviews", rh.GetByProduct) // product_id in query param
			
			// Favorites
			r.Get("/favorites", fh.GetUserFavorites)
			r.Post("/favorites", fh.Add) // product_id in body
			r.Delete("/favorites/{id}", fh.Remove)

			// Admin Section (Resource-based)
			r.Group(func(r chi.Router) {
				r.Use(custom_middleware.AdminMiddleware)
				
				// Admin Products management
				r.Post("/admin/products", aph.Create)
				r.Put("/admin/products/{id}", aph.Update)
				r.Delete("/admin/products/{id}", aph.Delete)
				
				// Admin Orders management
				r.Get("/admin/orders", aoh.GetAll)
				r.Patch("/admin/orders/{id}", aoh.UpdateStatus) // PATCH to update status only
			})
		})
	})

	return r
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
