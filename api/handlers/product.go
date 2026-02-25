package handlers

import (
	"coffeebase-api/api/dto"
	productmodel "coffeebase-api/internal/models/product"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ProductRepository interface {
	GetAll(f productmodel.Filter) ([]productmodel.Product, error)
	GetByID(id int) (productmodel.Product, error)
}

type ProductHandler struct {
	Store ProductRepository
}


func (h *ProductHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	filter := productmodel.Filter{
		Query:    q.Get("q"),
		Category: q.Get("category"),
	}

	if min := q.Get("min_price"); min != "" {
		filter.MinPrice, _ = strconv.ParseFloat(min, 64)
	}
	if max := q.Get("max_price"); max != "" {
		filter.MaxPrice, _ = strconv.ParseFloat(max, 64)
	}

	products, err := h.Store.GetAll(filter)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.MapProductsToResponse(products))
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)

	p, err := h.Store.GetByID(id)
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.MapProductToResponse(p))
}
