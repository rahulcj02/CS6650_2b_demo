package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetProductSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetProducts()
	router := setupRouter()

	req := httptest.NewRequest(http.MethodGet, "/products/1", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}

	var got Product
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got.ProductID != 1 {
		t.Fatalf("expected product_id 1, got %d", got.ProductID)
	}
}

func TestGetProductNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetProducts()
	router := setupRouter()

	req := httptest.NewRequest(http.MethodGet, "/products/999", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.Code)
	}
}

func TestPostProductDetailsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetProducts()
	router := setupRouter()

	body := []byte(`{
		"product_id": 1,
		"sku": "NEW-123",
		"manufacturer": "Updated Manufacturer",
		"category_id": 321,
		"weight": 111,
		"some_other_id": 654
	}`)

	req := httptest.NewRequest(http.MethodPost, "/products/1/details", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/products/1", nil)
	getResp := httptest.NewRecorder()
	router.ServeHTTP(getResp, getReq)
	if getResp.Code != http.StatusOK {
		t.Fatalf("expected 200 after update, got %d", getResp.Code)
	}

	var got Product
	if err := json.Unmarshal(getResp.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal get response: %v", err)
	}
	if got.SKU != "NEW-123" {
		t.Fatalf("expected updated SKU, got %q", got.SKU)
	}
}

func TestPostProductDetailsInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetProducts()
	router := setupRouter()

	body := []byte(`{
		"product_id": 1,
		"manufacturer": "Acme",
		"category_id": 100,
		"weight": 100,
		"some_other_id": 200
	}`)

	req := httptest.NewRequest(http.MethodPost, "/products/1/details", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

func TestPostProductDetailsMismatchID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetProducts()
	router := setupRouter()

	body := []byte(`{
		"product_id": 2,
		"sku": "ABC-123",
		"manufacturer": "Acme",
		"category_id": 100,
		"weight": 100,
		"some_other_id": 200
	}`)

	req := httptest.NewRequest(http.MethodPost, "/products/1/details", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

func TestPostProductDetailsProductNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetProducts()
	router := setupRouter()

	body := []byte(`{
		"product_id": 999,
		"sku": "ABC-123",
		"manufacturer": "Acme",
		"category_id": 100,
		"weight": 100,
		"some_other_id": 200
	}`)

	req := httptest.NewRequest(http.MethodPost, "/products/999/details", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.Code)
	}
}
