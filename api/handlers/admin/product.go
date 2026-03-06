package admin

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/internal/models/product"
	productstore "coffeebase-api/internal/store/product"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ProductHandler struct {
	Store *productstore.Store
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	p := product.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
	}

	if err := h.Store.Create(&p); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.MapProductToResponse(p))
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req dto.ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	p := product.Product{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
	}

	if err := h.Store.Update(&p); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(dto.MapProductToResponse(p))
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	if err := h.Store.Delete(id); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) CreateBulk(w http.ResponseWriter, r *http.Request) {
	var reqs []dto.ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var products []product.Product
	for _, req := range reqs {
		products = append(products, product.Product{
			Name:        req.Name,
			Description: req.Description,
			Price:       req.Price,
			Category:    req.Category,
		})
	}

	if err := h.Store.CreateBulk(products); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Bulk creation successful"})
}

