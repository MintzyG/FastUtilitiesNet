#!/bin/bash

# HTTP Response Library Test Script
# This script comprehensively tests all endpoints of the test application

BASE_URL="http://localhost:8080"
VERBOSE=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_header() {
    echo -e "${BLUE}============================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}============================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# Function to make curl request and format output
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4

    echo -e "\n${YELLOW}Testing:${NC} $description"
    echo -e "${YELLOW}Request:${NC} $method $endpoint"

    if [ "$VERBOSE" = true ]; then
        if [ -n "$data" ]; then
            curl -X $method \
                 -H "Content-Type: application/json" \
                 -d "$data" \
                 -w "\n\nStatus Code: %{http_code}\nTime: %{time_total}s\nSize: %{size_download} bytes\n" \
                 -v \
                 "$BASE_URL$endpoint"
        else
            curl -X $method \
                 -w "\n\nStatus Code: %{http_code}\nTime: %{time_total}s\nSize: %{size_download} bytes\n" \
                 -v \
                 "$BASE_URL$endpoint"
        fi
    else
        if [ -n "$data" ]; then
            curl -X $method \
                 -H "Content-Type: application/json" \
                 -d "$data" \
                 -w "\nStatus: %{http_code} | Time: %{time_total}s | Size: %{size_download}b\n" \
                 -s \
                 "$BASE_URL$endpoint" \
            | jq 'if (tostring | length) > 2000 then "LargeBody" else . end' 2>/dev/null || true
        else
            curl -X $method \
                 -w "\nStatus: %{http_code} | Time: %{time_total}s | Size: %{size_download}b\n" \
                 -s \
                 "$BASE_URL$endpoint" \
            | jq 'if (tostring | length) > 2000 then "LargeBody" else . end' 2>/dev/null || true
        fi
    fi

    echo -e "\n${BLUE}---${NC}"
}


# Check if server is running
check_server() {
    print_header "CHECKING SERVER STATUS"
    
    if curl -s "$BASE_URL/health" > /dev/null; then
        print_success "Server is running at $BASE_URL"
        return 0
    else
        print_error "Server is not running at $BASE_URL"
        print_info "Please start the server first: go run main.go"
        exit 1
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -u|--url)
            BASE_URL="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  -v, --verbose    Enable verbose output"
            echo "  -u, --url        Set base URL (default: http://localhost:8080)"
            echo "  -h, --help       Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option $1"
            exit 1
            ;;
    esac
done

# Start testing
print_header "HTTP Response Library Comprehensive Test"
echo "Testing against: $BASE_URL"
echo "Verbose mode: $VERBOSE"

check_server

# Test 1: Basic Endpoints
print_header "BASIC ENDPOINTS"

make_request "GET" "/" "" "Home endpoint - basic response with data"

make_request "GET" "/health" "" "Health check - service status"

make_request "GET" "/config" "" "Configuration - library settings"

# Test 2: Users CRUD Operations
print_header "USERS CRUD OPERATIONS"

make_request "GET" "/users" "" "List all users"

make_request "GET" "/users/1" "" "Get specific user (should exist)"

make_request "GET" "/users/999" "" "Get non-existent user (should return 404)"

make_request "POST" "/users" '{"name":"Test User","email":"test@example.com","username":"testuser"}' "Create new user"

make_request "POST" "/users" '{"name":"","email":""}' "Create user with validation errors"

make_request "POST" "/users" '{"invalid":"json"}' "Create user with invalid data" # fix: this json is valid, add a field int to user and send it as string or bool

make_request "PUT" "/users/1" '{"name":"Updated User","email":"updated@example.com","username":"updateduser"}' "Update existing user"

make_request "PUT" "/users/999" '{"name":"Non-existent","email":"none@example.com"}' "Update non-existent user"

make_request "DELETE" "/users/2" "" "Delete user"

make_request "DELETE" "/users/999" "" "Delete non-existent user"

# Test 3: Products
print_header "PRODUCTS ENDPOINTS"

make_request "GET" "/products" "" "Get all products"

make_request "GET" "/products?in_stock=true" "" "Get only in-stock products"

make_request "GET" "/products/1" "" "Get specific product"

make_request "GET" "/products/999" "" "Get non-existent product"

make_request "POST" "/products" "" "Method not allowed for products"

# Test 4: Error Testing
print_header "ERROR RESPONSE TESTING"

make_request "GET" "/error-test" "" "Error test info"

make_request "GET" "/error-test?type=400" "" "400 Bad Request"

make_request "GET" "/error-test?type=401" "" "401 Unauthorized"

make_request "GET" "/error-test?type=403" "" "403 Forbidden"

make_request "GET" "/error-test?type=404" "" "404 Not Found"

make_request "GET" "/error-test?type=409" "" "409 Conflict"

make_request "GET" "/error-test?type=422" "" "422 Unprocessable Entity"

make_request "GET" "/error-test?type=429" "" "429 Too Many Requests"

make_request "GET" "/error-test?type=500" "" "500 Internal Server Error"

make_request "GET" "/error-test?type=502" "" "502 Bad Gateway"

make_request "GET" "/error-test?type=503" "" "503 Service Unavailable"

# Test 5: Validation Testing
print_header "VALIDATION TESTING"

make_request "GET" "/validation-test" "" "Validation errors response"

# Test 6: Trace Testing
print_header "TRACE FUNCTIONALITY TESTING"

make_request "GET" "/trace-test" "" "Trace accumulation and limits"

# Test 7: Size Testing
print_header "RESPONSE SIZE TESTING"

make_request "GET" "/size-test" "" "Size test info"

make_request "GET" "/size-test?size=small" "" "Small response"

make_request "GET" "/size-test?size=large" "" "Large response (may hit size limits)"

# Test 8: Content Type Testing
print_header "CONTENT TYPE TESTING"

make_request "GET" "/content-type-test" "" "Default JSON content type"

make_request "GET" "/content-type-test?type=xml" "" "XML content type"

make_request "GET" "/content-type-test?type=text" "" "Plain text content type"

# Test 9: Interceptor Testing
print_header "INTERCEPTOR TESTING"

make_request "GET" "/interceptor-test" "" "Test interceptor functionality"

# Test 10: Status Codes Overview
print_header "STATUS CODES OVERVIEW"

make_request "GET" "/status-codes" "" "Available status code methods"

# Test 11: Method Testing
print_header "HTTP METHOD TESTING"

make_request "GET" "/users" "" "GET method (allowed)"

make_request "POST" "/users" '{}' "POST method (allowed)"

make_request "PATCH" "/users" "" "PATCH method (not allowed)"

make_request "DELETE" "/users" "" "DELETE method (not allowed)"

# Test 12: Edge Cases
print_header "EDGE CASES AND STRESS TESTING"

make_request "GET" "/nonexistent" "" "Non-existent endpoint (404)"

make_request "GET" "/users/abc" "" "Invalid ID format"

make_request "POST" "/users" '{"malformed json"}' "Malformed JSON"

# Multiple rapid requests to test concurrency
print_header "CONCURRENCY TESTING (5 rapid requests)"

for i in {1..5}; do
    print_info "Concurrent request $i"
    curl -s "$BASE_URL/" | jq -r '.message' &
done

wait
print_success "Concurrency test completed"

# Final summary
print_header "TEST SUMMARY"

print_success "All tests completed successfully!"
print_info "Check the application logs to see interceptor output"
print_info "Review response sizes, trace counts, and timing information"

echo -e "\n${YELLOW}Additional manual testing suggestions:${NC}"
echo "1. Test with different payload sizes to trigger size limits"
echo "2. Add more interceptors to test the interceptor limit"
echo "3. Modify configuration values and test impacts"
echo "4. Test concurrent high-load scenarios"
echo "5. Verify memory usage under load"

# Performance test section
print_header "SIMPLE PERFORMANCE TEST"
print_info "Running 50 requests to measure average response time..."

total_time=0
