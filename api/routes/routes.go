package routes

import (
	"coffeebase-api/api/handlers"
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
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Serve static files for uploads
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "uploads"))
	FileServer(r, "/uploads", filesDir)

	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Post("/auth/register", ah.Register)
		r.Post("/auth/login", ah.Login)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(custom_middleware.AuthMiddleware)
			
			// User Profile
			r.Get("/me", uh.GetProfile)
			r.Put("/me/language", uh.UpdateLanguage)
			r.Put("/me/avatar", uh.UploadAvatar)
			// Products
			r.Get("/products", ph.GetAll)
			r.Get("/products/{id}", ph.GetByID)
			
			// Orders & Checkout
			r.Post("/orders", oh.Create)
			r.Post("/orders/checkout", oh.Checkout)
			r.Get("/orders/history", oh.GetHistory)
			
			// Shopping Cart
			r.Get("/cart", ch.GetCart)
			r.Put("/cart", ch.UpdateItem)
			
			// Reviews
			r.Post("/products/{id}/reviews", rh.Create)
			r.Get("/products/{id}/reviews", rh.GetByProduct)
			
			// Favorites
			r.Post("/products/{id}/favorite", fh.Add)
			r.Delete("/products/{id}/favorite", fh.Remove)
			r.Get("/favorites", fh.GetUserFavorites)
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
