package handlers

import (
	"coffeebase-api/api/dto"
	productmodel "coffeebase-api/internal/models/product"
	"encoding/json"
	webServer "net/http"
	numberParsing "strconv"

	"github.com/go-chi/chi/v5"
)

/**
 * ProductRepository defines the expected behavior for product data persistence.
 * Refactored to follow strictly declarative naming.
 */
type ProductRepository interface {
	GetAll(filter productmodel.Filter) ([]productmodel.Product, error)
	GetByID(productID int) (productmodel.Product, error)
	GetCategories() ([]string, error)
}

/**
 * ProductHandler manages product-related HTTP requests.
 * Refactored to eliminate all shorthands and follow strictly declarative naming.
 */
type ProductHandler struct {
	Store ProductRepository
}

func (handler *ProductHandler) GetAll(responseWriter webServer.ResponseWriter, httpRequest *webServer.Request) {
	queryParameters := httpRequest.URL.Query()

	productFilter := productmodel.Filter{
		Query:    queryParameters.Get("q"),
		Category: queryParameters.Get("category"),
	}

	if minimumPrice := queryParameters.Get("min_price"); minimumPrice != "" {
		productFilter.MinPrice, _ = numberParsing.ParseFloat(minimumPrice, 64)
	}
	if maximumPrice := queryParameters.Get("max_price"); maximumPrice != "" {
		productFilter.MaxPrice, _ = numberParsing.ParseFloat(maximumPrice, 64)
	}

	productList, fetchError := handler.Store.GetAll(productFilter)
	if fetchError != nil {
		webServer.Error(responseWriter, "Internal server error", webServer.StatusInternalServerError)
		return
	}
	responseWriter.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseWriter).Encode(dto.MapProductsToResponse(productList))
}

func (handler *ProductHandler) GetByID(responseWriter webServer.ResponseWriter, httpRequest *webServer.Request) {
	productIDString := chi.URLParam(httpRequest, "id")
	productID, _ := numberParsing.Atoi(productIDString)

	productInstance, fetchError := handler.Store.GetByID(productID)
	if fetchError != nil {
		webServer.Error(responseWriter, "Product not found", webServer.StatusNotFound)
		return
	}
	responseWriter.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseWriter).Encode(dto.MapProductToResponse(productInstance))
}

func (handler *ProductHandler) GetCategories(responseWriter webServer.ResponseWriter, httpRequest *webServer.Request) {
	categoryList, fetchError := handler.Store.GetCategories()
	if fetchError != nil {
		webServer.Error(responseWriter, "Internal server error", webServer.StatusInternalServerError)
		return
	}
	
	responseWriter.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseWriter).Encode(categoryList)
}

