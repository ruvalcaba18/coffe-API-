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
	ProductStore *productstore.Store
}

func (productHandler *ProductHandler) Create(responseWriter http.ResponseWriter, request *http.Request) {
	var productRequest dto.ProductRequest
	if decodeError := json.NewDecoder(request.Body).Decode(&productRequest); decodeError != nil {
		http.Error(responseWriter, "Invalid request body", http.StatusBadRequest)
		return
	}

	productInstance := product.Product{
		Name:        productRequest.Name,
		Description: productRequest.Description,
		Price:       productRequest.Price,
		Category:    productRequest.Category,
	}

	if createError := productHandler.ProductStore.Create(&productInstance); createError != nil {
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
		return
	}

	responseWriter.WriteHeader(http.StatusCreated)
	json.NewEncoder(responseWriter).Encode(dto.MapProductToResponse(productInstance))
}

func (productHandler *ProductHandler) Update(responseWriter http.ResponseWriter, request *http.Request) {
	productIdentifier, _ := strconv.Atoi(chi.URLParam(request, "id"))
	var productRequest dto.ProductRequest
	if decodeError := json.NewDecoder(request.Body).Decode(&productRequest); decodeError != nil {
		http.Error(responseWriter, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	productInstance := product.Product{
		ID:          productIdentifier,
		Name:        productRequest.Name,
		Description: productRequest.Description,
		Price:       productRequest.Price,
		Category:    productRequest.Category,
	}

	if updateError := productHandler.ProductStore.Update(&productInstance); updateError != nil {
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(responseWriter).Encode(dto.MapProductToResponse(productInstance))
}

func (productHandler *ProductHandler) Delete(responseWriter http.ResponseWriter, request *http.Request) {
	productIdentifier, _ := strconv.Atoi(chi.URLParam(request, "id"))

	if deletionError := productHandler.ProductStore.Delete(productIdentifier); deletionError != nil {
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
		return
	}

	responseWriter.WriteHeader(http.StatusNoContent)
}

func (productHandler *ProductHandler) CreateBulk(responseWriter http.ResponseWriter, request *http.Request) {
	var productRequests []dto.ProductRequest
	if decodeError := json.NewDecoder(request.Body).Decode(&productRequests); decodeError != nil {
		http.Error(responseWriter, "Invalid request body", http.StatusBadRequest)
		return
	}

	var productList []product.Product
	for _, productRequest := range productRequests {
		productList = append(productList, product.Product{
			Name:        productRequest.Name,
			Description: productRequest.Description,
			Price:       productRequest.Price,
			Category:    productRequest.Category,
		})
	}

	if bulkCreateError := productHandler.ProductStore.CreateBulk(productList); bulkCreateError != nil {
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
		return
	}

	responseWriter.WriteHeader(http.StatusCreated)
	json.NewEncoder(responseWriter).Encode(map[string]string{"message": "Bulk creation successful"})
}
