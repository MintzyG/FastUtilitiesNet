package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MintzyG/GoResponse/response"
)

// Test interceptor for logging
type LoggingInterceptor struct{}

func (l *LoggingInterceptor) Intercept(ctx context.Context, resp *response.Response, statusCode int) {
	log.Printf("Intercepted response with context: %d - %s", statusCode, resp.Message)
}

func (l *LoggingInterceptor) InterceptSimple(resp *response.Response, statusCode int) {
	log.Printf("Intercepted response without context: %d - %s", statusCode, resp.Message)
}

// Test interceptor for metrics
type MetricsInterceptor struct{}

func (m *MetricsInterceptor) Intercept(ctx context.Context, resp *response.Response, statusCode int) {
	stats := resp.GetResponseStats()
	log.Printf("Response stats: %+v", stats)
}

func (m *MetricsInterceptor) InterceptSimple(resp *response.Response, statusCode int) {
	stats := resp.GetResponseStats()
	log.Printf("Response stats (simple): %+v", stats)
}

// Test data structures
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	InStock     bool    `json:"in_stock"`
}

// Mock data
var users = []User{
	{ID: 1, Name: "John Doe", Email: "john@example.com", Username: "johndoe"},
	{ID: 2, Name: "Jane Smith", Email: "jane@example.com", Username: "janesmith"},
	{ID: 3, Name: "Bob Johnson", Email: "bob@example.com", Username: "bobjohnson"},
}

var products = []Product{
	{ID: 1, Name: "Laptop", Description: "High-performance laptop", Price: 999.99, InStock: true},
	{ID: 2, Name: "Mouse", Description: "Wireless mouse", Price: 29.99, InStock: true},
	{ID: 3, Name: "Keyboard", Description: "Mechanical keyboard", Price: 149.99, InStock: false},
}

func main() {
	// Setup interceptors
	if err := response.AddInterceptor(&LoggingInterceptor{}); err != nil {
		log.Fatal("Failed to add logging interceptor:", err)
	}
	if err := response.AddInterceptor(&MetricsInterceptor{}); err != nil {
		log.Fatal("Failed to add metrics interceptor:", err)
	}

	// Configure the response library
	config := response.Config{
		MaxTraceSize:         10,
		ResponseSizeLimit:    1024 * 1024, // 1MB
		MaxInterceptorAmount: 10,
		DefaultContentType:   "application/json",
		EnableSizeValidation: true,
	}
	response.SetConfig(config)

	// Register routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/config", configHandler)
	http.HandleFunc("/users", usersHandler)
	http.HandleFunc("/users/", userHandler)
	http.HandleFunc("/products", productsHandler)
	http.HandleFunc("/products/", productHandler)
	http.HandleFunc("/validation-test", validationTestHandler)
	http.HandleFunc("/error-test", errorTestHandler)
	http.HandleFunc("/trace-test", traceTestHandler)
	http.HandleFunc("/size-test", sizeTestHandler)
	http.HandleFunc("/content-type-test", contentTypeTestHandler)
	http.HandleFunc("/interceptor-test", interceptorTestHandler)
	http.HandleFunc("/status-codes", statusCodesHandler)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Home handler - basic response
func homeHandler(w http.ResponseWriter, r *http.Request) {
	response.OK().
		WithModule("main").
		WithMessage("Welcome to the Response Library Test API").
		WithData(map[string]string{
			"version": "1.0.0",
			"status":  "running",
		}).
		AppendTrace("Home endpoint accessed").
		SendWithContext(r.Context(), w)
}

// Health check handler
func healthHandler(w http.ResponseWriter, r *http.Request) {
	response.OK().
		WithMessage("Service is healthy").
		WithData(map[string]any{
			"timestamp": time.Now(),
			"uptime":    "simulated",
			"checks": map[string]bool{
				"database": true,
				"cache":    true,
				"external": true,
			},
		}).
		Send(w)
}

// Configuration handler
func configHandler(w http.ResponseWriter, r *http.Request) {
	config := response.GetConfig()
	response.OK().
		WithMessage("Current configuration").
		WithData(config).
		AppendTrace("Configuration retrieved").
		Send(w)
}

// Users handler - list all users or create new user
func usersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		response.OK().
			WithMessage("Users retrieved successfully").
			WithData(users).
			AppendTrace("Listed all users").
			SendWithContext(r.Context(), w)

	case http.MethodPost:
		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			response.BadRequest().
				WithMessage("Invalid JSON payload").
				AppendTrace("JSON decode failed", err).
				SendWithContext(r.Context(), w)
			return
		}

		// Simple validation
		if user.Name == "" || user.Email == "" {
			validationErrs := []response.ValidationErr{
				{Field: "name", Message: "Name is required", Value: user.Name},
				{Field: "email", Message: "Email is required", Value: user.Email},
			}
			response.WithValidationErrors(validationErrs).SendWithContext(r.Context(), w)
			return
		}

		user.ID = len(users) + 1
		users = append(users, user)

		response.Created().
			WithMessage("User created successfully").
			WithData(user).
			AppendTrace("User created", fmt.Sprintf("ID: %d", user.ID)).
			SendWithContext(r.Context(), w)

	default:
		response.MethodNotAllowed().
			WithMessage("Method not allowed").
			AppendTrace("Invalid method", r.Method).
			SendWithContext(r.Context(), w)
	}
}

// User handler - get, update, or delete specific user
func userHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/users/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest().
			WithMessage("Invalid user ID").
			AppendTrace("ID parsing failed", err).
			SendWithContext(r.Context(), w)
		return
	}

	// Find user
	var userIndex = -1
	for i, user := range users {
		if user.ID == id {
			userIndex = i
			break
		}
	}

	switch r.Method {
	case http.MethodGet:
		if userIndex == -1 {
			response.NotFound().
				WithMessage("User not found").
				AppendTrace("User lookup failed", fmt.Sprintf("ID: %d", id)).
				SendWithContext(r.Context(), w)
			return
		}

		response.OK().
			WithMessage("User retrieved successfully").
			WithData(users[userIndex]).
			AppendTrace("User found", fmt.Sprintf("ID: %d", id)).
			SendWithContext(r.Context(), w)

	case http.MethodPut:
		if userIndex == -1 {
			response.NotFound().
				WithMessage("User not found").
				SendWithContext(r.Context(), w)
			return
		}

		var updatedUser User
		if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
			response.BadRequest().
				WithMessage("Invalid JSON payload").
				AppendTrace("JSON decode failed", err).
				SendWithContext(r.Context(), w)
			return
		}

		updatedUser.ID = id
		users[userIndex] = updatedUser

		response.OK().
			WithMessage("User updated successfully").
			WithData(updatedUser).
			AppendTrace("User updated", fmt.Sprintf("ID: %d", id)).
			SendWithContext(r.Context(), w)

	case http.MethodDelete:
		if userIndex == -1 {
			response.NotFound().
				WithMessage("User not found").
				SendWithContext(r.Context(), w)
			return
		}

		users = append(users[:userIndex], users[userIndex+1:]...)

		response.NoContent().
			WithMessage("User deleted successfully").
			AppendTrace("User deleted", fmt.Sprintf("ID: %d", id)).
			SendWithContext(r.Context(), w)

	default:
		response.MethodNotAllowed().
			WithMessage("Method not allowed").
			SendWithContext(r.Context(), w)
	}
}

// Products handler
func productsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		inStockOnly := r.URL.Query().Get("in_stock") == "true"

		var filteredProducts []Product
		for _, product := range products {
			if !inStockOnly || product.InStock {
				filteredProducts = append(filteredProducts, product)
			}
		}

		response.OK().
			WithMessage("Products retrieved successfully").
			WithData(map[string]any{
				"products": filteredProducts,
				"count":    len(filteredProducts),
				"filtered": inStockOnly,
			}).
			AppendTrace("Products listed", fmt.Sprintf("Count: %d", len(filteredProducts))).
			SendWithContext(r.Context(), w)

	default:
		response.MethodNotAllowed().
			WithMessage("Method not allowed").
			SendWithContext(r.Context(), w)
	}
}

// Product handler
func productHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/products/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest().
			WithMessage("Invalid product ID").
			AppendTrace("ID parsing failed", err).
			SendWithContext(r.Context(), w)
		return
	}

	for _, product := range products {
		if product.ID == id {
			response.OK().
				WithMessage("Product found").
				WithData(product).
				AppendTrace("Product retrieved", fmt.Sprintf("ID: %d", id)).
				SendWithContext(r.Context(), w)
			return
		}
	}

	response.NotFound().
		WithMessage("Product not found").
		AppendTrace("Product lookup failed", fmt.Sprintf("ID: %d", id)).
		SendWithContext(r.Context(), w)
}

// Validation test handler
func validationTestHandler(w http.ResponseWriter, r *http.Request) {
	validationErrs := []response.ValidationErr{
		{Field: "email", Message: "Invalid email format", Value: "invalid-email"},
		{Field: "age", Message: "Age must be between 18 and 100", Value: 150},
		{Field: "password", Message: "Password too short"},
	}

	response.WithValidationErrors(validationErrs).SendWithContext(r.Context(), w)
}

// Error test handler - test different error types
func errorTestHandler(w http.ResponseWriter, r *http.Request) {
	errorType := r.URL.Query().Get("type")

	switch errorType {
	case "400":
		response.BadRequest().
			WithMessage("This is a bad request error").
			AppendTrace("Bad request triggered").
			SendWithContext(r.Context(), w)
	case "401":
		response.Unauthorized().
			WithMessage("Authentication required").
			AppendTrace("Unauthorized access attempt").
			SendWithContext(r.Context(), w)
	case "403":
		response.Forbidden().
			WithMessage("Access denied").
			AppendTrace("Forbidden resource access").
			SendWithContext(r.Context(), w)
	case "404":
		response.NotFound().
			WithMessage("Resource not found").
			AppendTrace("Resource lookup failed").
			SendWithContext(r.Context(), w)
	case "409":
		response.Conflict().
			WithMessage("Resource conflict").
			AppendTrace("Conflict detected").
			SendWithContext(r.Context(), w)
	case "422":
		response.UnprocessableEntity().
			WithMessage("Unprocessable entity").
			AppendTrace("Entity processing failed").
			SendWithContext(r.Context(), w)
	case "429":
		response.TooManyRequests().
			WithMessage("Rate limit exceeded").
			AppendTrace("Too many requests").
			SendWithContext(r.Context(), w)
	case "500":
		response.InternalServerError().
			WithMessage("Internal server error").
			AppendTrace("Simulated internal error").
			SendWithContext(r.Context(), w)
	case "502":
		response.BadGateway().
			WithMessage("Bad gateway").
			AppendTrace("Gateway error").
			SendWithContext(r.Context(), w)
	case "503":
		response.ServiceUnavailable().
			WithMessage("Service unavailable").
			AppendTrace("Service down").
			SendWithContext(r.Context(), w)
	default:
		response.OK().
			WithMessage("Error test endpoint - use ?type=400|401|403|404|409|422|429|500|502|503").
			SendWithContext(r.Context(), w)
	}
}

// Trace test handler - test trace functionality
func traceTestHandler(w http.ResponseWriter, r *http.Request) {
	resp := response.OK().
		WithMessage("Trace test response").
		AppendTrace("Step 1: Request received").
		AppendTrace("Step 2: Processing started").
		AppendTrace("Step 3: Data validation passed").
		AppendTrace("Step 4: Business logic executed").
		AppendTrace("Step 5: Database operation completed").
		AppendTrace("Step 6: Response prepared").
		AppendTrace("Step 7: Additional trace entry").
		AppendTrace("Step 8: Another trace entry").
		AppendTrace("Step 9: More trace information").
		AppendTrace("Step 10: Even more traces").
		AppendTrace("Step 11: This should test trace limits").
		AppendTrace("Step 12: Final trace entry")

	resp.SendWithContext(r.Context(), w)
}

// Size test handler - test response size limits
func sizeTestHandler(w http.ResponseWriter, r *http.Request) {
	sizeParam := r.URL.Query().Get("size")

	switch sizeParam {
	case "large":
		// Create a large response
		largeData := make(map[string]any)
		for i := 0; i < 1000; i++ {
			largeData[fmt.Sprintf("key_%d", i)] = strings.Repeat("data", 100)
		}

		response.OK().
			WithMessage("Large response test").
			WithData(largeData).
			AppendTrace("Generated large response").
			SendWithContext(r.Context(), w)

	case "small":
		response.OK().
			WithMessage("Small response test").
			WithData("small data").
			SendWithContext(r.Context(), w)

	default:
		response.OK().
			WithMessage("Size test endpoint - use ?size=large|small").
			SendWithContext(r.Context(), w)
	}
}

// Content type test handler
func contentTypeTestHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.URL.Query().Get("type")

	switch contentType {
	case "xml":
		response.OK().
			WithContentType("application/xml").
			WithMessage("XML content type test").
			WithData("This would be XML if properly formatted").
			SendWithContext(r.Context(), w)

	case "text":
		response.OK().
			WithContentType("text/plain").
			WithMessage("Plain text content type test").
			SendWithContext(r.Context(), w)

	default:
		response.OK().
			WithContentType("application/json").
			WithMessage("Default JSON content type").
			SendWithContext(r.Context(), w)
	}
}

// Interceptor test handler
func interceptorTestHandler(w http.ResponseWriter, r *http.Request) {
	// This handler specifically tests interceptor functionality
	response.OK().
		WithMessage("This response will be intercepted by all registered interceptors").
		WithData(map[string]any{
			"interceptors_count": len(response.GetInterceptors()),
			"test_type":          "interceptor",
		}).
		AppendTrace("Interceptor test initiated").
		SendWithContext(r.Context(), w)
}

// Status codes handler - showcase all available status codes
func statusCodesHandler(w http.ResponseWriter, r *http.Request) {
	statusCodes := map[string]string{
		"200": "OK",
		"201": "Created",
		"202": "Accepted",
		"204": "NoContent",
		"400": "BadRequest",
		"401": "Unauthorized",
		"402": "PaymentRequired",
		"403": "Forbidden",
		"404": "NotFound",
		"405": "MethodNotAllowed",
		"409": "Conflict",
		"422": "UnprocessableEntity",
		"429": "TooManyRequests",
		"500": "InternalServerError",
		"501": "NotImplemented",
		"502": "BadGateway",
		"503": "ServiceUnavailable",
	}

	response.OK().
		WithMessage("Available status code methods").
		WithData(statusCodes).
		AppendTrace("Status codes listed").
		SendWithContext(r.Context(), w)
}
