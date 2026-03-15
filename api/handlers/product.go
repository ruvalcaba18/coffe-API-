package handlers

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/api/response"
	"coffeebase-api/internal/apperrors"
	productmodel "coffeebase-api/internal/models/product"
	"coffeebase-api/internal/store/product"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ProductHandler struct {
	productStore product.Store
}

// --- Public ---

func NewProductHandler(productStore product.Store) *ProductHandler {
	return &ProductHandler{
		productStore: productStore,
	}
}

func (productHandler *ProductHandler) GetAll(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	queryParameters := httpRequest.URL.Query()

	productFilter := productmodel.Filter{
		Query:    queryParameters.Get("q"),
		Category: queryParameters.Get("category"),
	}

	if minimumPriceString := queryParameters.Get("min_price"); minimumPriceString != "" {
		productFilter.MinPrice, _ = strconv.ParseFloat(minimumPriceString, 64)
	}
	if maximumPriceString := queryParameters.Get("max_price"); maximumPriceString != "" {
		productFilter.MaxPrice, _ = strconv.ParseFloat(maximumPriceString, 64)
	}

	productList, error := productHandler.productStore.GetAll(httpRequest.Context(), productFilter)
	if error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}
	
	response.SendJSON(responseWriter, http.StatusOK, dto.MapProductsToResponse(productList))
}

func (productHandler *ProductHandler) GetByID(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	productIDString := chi.URLParam(httpRequest, "id")
	productID, conversionError := strconv.Atoi(productIDString)
	if conversionError != nil {
		response.SendError(responseWriter, apperrors.ErrInvalidID)
		return
	}

	productInstance, error := productHandler.productStore.GetByID(httpRequest.Context(), productID)
	if error != nil {
		response.SendError(responseWriter, apperrors.ErrProductNotFound)
		return
	}
	
	response.SendJSON(responseWriter, http.StatusOK, dto.MapProductToResponse(productInstance))
}

func (productHandler *ProductHandler) GetCategories(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	categoryList, error := productHandler.productStore.GetCategories(httpRequest.Context())
	if error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}
	
	response.SendJSON(responseWriter, http.StatusOK, categoryList)
}
