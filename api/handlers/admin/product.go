package admin

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/api/response"
	"coffeebase-api/internal/apperrors"
	"coffeebase-api/internal/models/product"
	productstore "coffeebase-api/internal/store/product"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ProductHandler struct {
	productStore productstore.Store
}

// --- Public ---

func NewProductHandler(productStore productstore.Store) *ProductHandler {
	return &ProductHandler{
		productStore: productStore,
	}
}

func (productHandler *ProductHandler) Create(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	var request dto.ProductRequest
	if error := response.DecodeJSON(httpRequest, &request); error != nil {
		response.SendError(responseWriter, error)
		return
	}

	productInstance := product.Product{
		Name:        request.Name,
		Description: request.Description,
		Price:       request.Price,
		Category:    request.Category,
	}

	if error := productHandler.productStore.Create(httpRequest.Context(), &productInstance); error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	response.SendJSON(responseWriter, http.StatusCreated, dto.MapProductToResponse(productInstance))
}

func (productHandler *ProductHandler) Update(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	productIDString := chi.URLParam(httpRequest, "id")
	productID, conversionError := strconv.Atoi(productIDString)
	if conversionError != nil {
		response.SendError(responseWriter, apperrors.ErrInvalidID)
		return
	}

	var request dto.ProductRequest
	if error := response.DecodeJSON(httpRequest, &request); error != nil {
		response.SendError(responseWriter, error)
		return
	}
	
	productInstance := product.Product{
		ID:          productID,
		Name:        request.Name,
		Description: request.Description,
		Price:       request.Price,
		Category:    request.Category,
	}

	if error := productHandler.productStore.Update(httpRequest.Context(), &productInstance); error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	response.SendJSON(responseWriter, http.StatusOK, dto.MapProductToResponse(productInstance))
}

func (productHandler *ProductHandler) Delete(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	productIDString := chi.URLParam(httpRequest, "id")
	productID, conversionError := strconv.Atoi(productIDString)
	if conversionError != nil {
		response.SendError(responseWriter, apperrors.ErrInvalidID)
		return
	}

	if error := productHandler.productStore.Delete(httpRequest.Context(), productID); error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	responseWriter.WriteHeader(http.StatusNoContent)
}

func (productHandler *ProductHandler) CreateBulk(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	var requests []dto.ProductRequest
	if error := response.DecodeJSON(httpRequest, &requests); error != nil {
		response.SendError(responseWriter, error)
		return
	}

	var productList []product.Product
	for _, productRequest := range requests {
		productList = append(productList, product.Product{
			Name:        productRequest.Name,
			Description: productRequest.Description,
			Price:       productRequest.Price,
			Category:    productRequest.Category,
		})
	}

	if error := productHandler.productStore.CreateBulk(httpRequest.Context(), productList); error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	response.SendJSON(responseWriter, http.StatusCreated, map[string]string{"message": "Bulk creation successful"})
}
