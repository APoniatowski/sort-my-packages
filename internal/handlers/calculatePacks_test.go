package handlers

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCalculatePacks tests the CalculatePacks function for various scenarios.
func TestCalculatePacks(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    PackRequest
		expectedStatus int
		expectedBody   PackResponse
	}{
		// Basic scenarios
		{"Single pack of 250", PackRequest{Quantity: 250}, http.StatusOK, PackResponse{Packs: map[int]int{250: 1}, TotalPacks: 1}},
		{"Single pack of 500", PackRequest{Quantity: 500}, http.StatusOK, PackResponse{Packs: map[int]int{500: 1}, TotalPacks: 1}},
		{"Single pack of 1000", PackRequest{Quantity: 1000}, http.StatusOK, PackResponse{Packs: map[int]int{1000: 1}, TotalPacks: 1}},
		{"Mixed packs: 750 (500 + 250)", PackRequest{Quantity: 750}, http.StatusOK, PackResponse{Packs: map[int]int{500: 1, 250: 1}, TotalPacks: 2}},

		// Edge cases
		{"Small quantity requiring multiple packs: 251", PackRequest{Quantity: 251}, http.StatusOK, PackResponse{Packs: map[int]int{500: 1}, TotalPacks: 1}},
		{"Complex quantity: 1200 (1000 + 250)", PackRequest{Quantity: 1200}, http.StatusOK, PackResponse{Packs: map[int]int{1000: 1, 250: 1}, TotalPacks: 2}},

		// Boundary conditions
		{"Quantity 0", PackRequest{Quantity: 0}, http.StatusBadRequest, PackResponse{}},
		{"Negative quantity", PackRequest{Quantity: -100}, http.StatusBadRequest, PackResponse{}},

		// Large quantity triggering greedy fallback
		{"Large quantity triggering greedy fallback", PackRequest{Quantity: 100000000}, http.StatusOK, PackResponse{Packs: map[int]int{5000: 20000}, TotalPacks: 20000}},

		// Quantity exceeding maximum limit
		{"Quantity exceeding maximum limit", PackRequest{Quantity: math.MaxInt32 + 1}, http.StatusBadRequest, PackResponse{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal request body
			requestBody, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			// Create a new request
			req := httptest.NewRequest(http.MethodPost, "/calculate-packs", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Create a response recorder
			rr := httptest.NewRecorder()

			// Call the handler function
			handler := http.HandlerFunc(CalculatePacks)
			handler.ServeHTTP(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Check response body if status is OK
			if tt.expectedStatus == http.StatusOK {
				var responseBody PackResponse
				err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, responseBody)
			}
		})
	}
}

// TestCalculatePacksInvalidInputs tests the CalculatePacks function for invalid inputs.
func TestCalculatePacksInvalidInputs(t *testing.T) {
	invalidJSONPayloads := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{"Non-integer quantity", `{"quantity": "abc"}`, http.StatusBadRequest},
		{"Incorrect JSON key", `{"quant": 100}`, http.StatusBadRequest},
		{"Missing quantity field", `{}`, http.StatusBadRequest},
		{"Empty payload", ``, http.StatusBadRequest},
	}

	for _, tt := range invalidJSONPayloads {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new request with invalid JSON
			req := httptest.NewRequest(http.MethodPost, "/calculate-packs", bytes.NewBuffer([]byte(tt.requestBody)))
			req.Header.Set("Content-Type", "application/json")

			// Create a response recorder
			rr := httptest.NewRecorder()

			// Call the handler function
			handler := http.HandlerFunc(CalculatePacks)
			handler.ServeHTTP(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

// TestCalculatePacksInvalidMethod tests the CalculatePacks function for invalid HTTP methods.
func TestCalculatePacksInvalidMethod(t *testing.T) {
	// Create a new request with an invalid HTTP method
	req := httptest.NewRequest(http.MethodGet, "/calculate-packs", nil)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler function
	handler := http.HandlerFunc(CalculatePacks)
	handler.ServeHTTP(rr, req)

	// Check status code
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

// TestCalculatePacksInvalidContentType tests the CalculatePacks function for invalid Content-Type.
func TestCalculatePacksInvalidContentType(t *testing.T) {
	// Create a new request with an invalid Content-Type
	req := httptest.NewRequest(http.MethodPost, "/calculate-packs", bytes.NewBuffer([]byte(`{"quantity": 100}`)))
	req.Header.Set("Content-Type", "text/plain")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler function
	handler := http.HandlerFunc(CalculatePacks)
	handler.ServeHTTP(rr, req)

	// Check status code
	assert.Equal(t, http.StatusUnsupportedMediaType, rr.Code)
}

// TestSetPackSizes tests the SetPackSizes function.
func TestSetPackSizes(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    PackSizesRequest
		expectedStatus int
		expectedBody   string
		origin         string
		authToken      string
	}{
		// Valid pack sizes and authorization
		{"Valid Pack Sizes", PackSizesRequest{PackSizes: []int{100, 200, 300}}, http.StatusOK, "Pack sizes updated successfully", AllowedOrigin, "Bearer " + AuthToken},

		// Invalid scenarios
		{"Empty Pack Sizes", PackSizesRequest{PackSizes: []int{}}, http.StatusBadRequest, "", AllowedOrigin, "Bearer " + AuthToken},
		{"Negative Pack Size", PackSizesRequest{PackSizes: []int{100, -200, 300}}, http.StatusBadRequest, "", AllowedOrigin, "Bearer " + AuthToken},
		{"Zero Pack Size", PackSizesRequest{PackSizes: []int{100, 0, 300}}, http.StatusBadRequest, "", AllowedOrigin, "Bearer " + AuthToken},

		// Invalid Origin and Auth Token
		{"Invalid Origin", PackSizesRequest{PackSizes: []int{100, 200, 300}}, http.StatusForbidden, "", "http://invalid-origin.com", "Bearer " + AuthToken},
		{"Invalid Auth Token", PackSizesRequest{PackSizes: []int{100, 200, 300}}, http.StatusUnauthorized, "", AllowedOrigin, "Bearer invalid_token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/set-pack-sizes", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Origin", tt.origin)
			req.Header.Set("Authorization", tt.authToken)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(SetPackSizes)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Check response body if status is OK
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, tt.expectedBody, rr.Body.String())
			}
		})
	}
}
