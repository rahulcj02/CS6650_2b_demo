package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

type Product struct {
	ProductID    int    `json:"product_id"`
	SKU          string `json:"sku"`
	Manufacturer string `json:"manufacturer"`
	CategoryID   int    `json:"category_id"`
	Weight       int    `json:"weight"`
	SomeOtherID  int    `json:"some_other_id"`
}

type productPayload struct {
	ProductID    *int   `json:"product_id"`
	SKU          string `json:"sku"`
	Manufacturer string `json:"manufacturer"`
	CategoryID   *int   `json:"category_id"`
	Weight       *int   `json:"weight"`
	SomeOtherID  *int   `json:"some_other_id"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

var (
	productsMu sync.RWMutex
	products   = seedProducts()
)

func seedProducts() map[int]Product {
	return map[int]Product{
		1: {
			ProductID:    1,
			SKU:          "ABC-123-XYZ",
			Manufacturer: "Acme Corporation",
			CategoryID:   100,
			Weight:       1250,
			SomeOtherID:  500,
		},
		2: {
			ProductID:    2,
			SKU:          "FOO-222-BAR",
			Manufacturer: "Globex",
			CategoryID:   101,
			Weight:       900,
			SomeOtherID:  501,
		},
	}
}

func resetProducts() {
	productsMu.Lock()
	products = seedProducts()
	productsMu.Unlock()
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/products/:productId", getProduct)
	router.POST("/products/:productId/details", addProductDetails)
	return router
}

func main() {
	router := setupRouter()
	_ = router.Run(":8080")
}

func getProduct(c *gin.Context) {
	productID, err := parseProductID(c.Param("productId"))
	if err != nil {
		respondError(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found", err.Error())
		return
	}

	productsMu.RLock()
	product, ok := products[productID]
	productsMu.RUnlock()
	if !ok {
		respondError(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found", fmt.Sprintf("No product with id %d", productID))
		return
	}

	c.JSON(http.StatusOK, product)
}

func addProductDetails(c *gin.Context) {
	productID, err := parseProductID(c.Param("productId"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_INPUT", "The provided input data is invalid", err.Error())
		return
	}

	var payload productPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_INPUT", "The provided input data is invalid", "Request body must be valid JSON")
		return
	}

	validated, detailErr := validateProductPayload(payload)
	if detailErr != nil {
		respondError(c, http.StatusBadRequest, "INVALID_INPUT", "The provided input data is invalid", detailErr.Error())
		return
	}

	if validated.ProductID != productID {
		respondError(c, http.StatusBadRequest, "INVALID_INPUT", "The provided input data is invalid", "Path productId must match body product_id")
		return
	}

	productsMu.Lock()
	if _, ok := products[productID]; !ok {
		productsMu.Unlock()
		respondError(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found", fmt.Sprintf("No product with id %d", productID))
		return
	}
	products[productID] = validated
	productsMu.Unlock()

	c.Status(http.StatusNoContent)
}

func parseProductID(raw string) (int, error) {
	productID, err := strconv.Atoi(raw)
	if err != nil || productID < 1 {
		return 0, fmt.Errorf("Product ID must be a positive integer")
	}
	return productID, nil
}

func validateProductPayload(p productPayload) (Product, error) {
	if p.ProductID == nil {
		return Product{}, fmt.Errorf("product_id is required")
	}
	if *p.ProductID < 1 {
		return Product{}, fmt.Errorf("product_id must be >= 1")
	}

	sku := strings.TrimSpace(p.SKU)
	if sku == "" {
		return Product{}, fmt.Errorf("sku is required")
	}
	if len(sku) > 100 {
		return Product{}, fmt.Errorf("sku length must be <= 100")
	}

	manufacturer := strings.TrimSpace(p.Manufacturer)
	if manufacturer == "" {
		return Product{}, fmt.Errorf("manufacturer is required")
	}
	if len(manufacturer) > 200 {
		return Product{}, fmt.Errorf("manufacturer length must be <= 200")
	}

	if p.CategoryID == nil {
		return Product{}, fmt.Errorf("category_id is required")
	}
	if *p.CategoryID < 1 {
		return Product{}, fmt.Errorf("category_id must be >= 1")
	}

	if p.Weight == nil {
		return Product{}, fmt.Errorf("weight is required")
	}
	if *p.Weight < 0 {
		return Product{}, fmt.Errorf("weight must be >= 0")
	}

	if p.SomeOtherID == nil {
		return Product{}, fmt.Errorf("some_other_id is required")
	}
	if *p.SomeOtherID < 1 {
		return Product{}, fmt.Errorf("some_other_id must be >= 1")
	}

	return Product{
		ProductID:    *p.ProductID,
		SKU:          sku,
		Manufacturer: manufacturer,
		CategoryID:   *p.CategoryID,
		Weight:       *p.Weight,
		SomeOtherID:  *p.SomeOtherID,
	}, nil
}

func respondError(c *gin.Context, status int, code, message, details string) {
	resp := ErrorResponse{
		Error:   code,
		Message: message,
		Details: details,
	}
	c.JSON(status, resp)
}
