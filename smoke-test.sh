#!/bin/bash

API_URL="http://localhost:8080/calculate-packs"

send_request() {
	local quantity=$1
	echo -e "\nTesting with quantity: $quantity"
	response=$(curl -s -X POST "$API_URL" -H "Content-Type: application/json" -d "{\"quantity\": $quantity}")
	echo "Response: $response"
}

echo "Starting API tests..."

# Normal use cases
send_request 1     # Minimal quantity, should return 1 x 250 pack
send_request 250   # Exact pack size, should return 1 x 250 pack
send_request 251   # Slightly above a single pack, should return 1 x 500 pack
send_request 500   # Exact pack size, should return 1 x 500 pack
send_request 750   # Should use 1 x 500 + 1 x 250
send_request 1200  # Should use 1 x 1000 + 1 x 250
send_request 12001 # Should use 2 x 5000 + 1 x 2000 + 1 x 250

# Large quantities
send_request 100000  # Large number to see if the algorithm handles it correctly
send_request 1000000 # Even larger number

# Edge cases
send_request 0                   # Invalid, should return an error or empty response
send_request -1                  # Negative number, should return an error
send_request 2147483647          # Maximum 32-bit signed int, to test large number handling
send_request 9223372036854775807 # Maximum 64-bit signed int, should test for overflow

echo -e "\nTesting with invalid JSON payloads..."
response=$(curl -s -X POST "$API_URL" -H "Content-Type: application/json" -d "{\"quantity\": \"abc\"}")
echo "Response for non-integer quantity: $response"

response=$(curl -s -X POST "$API_URL" -H "Content-Type: application/json" -d "{\"qty\": 1000}")
echo "Response for incorrect JSON key: $response"

response=$(curl -s -X POST "$API_URL" -H "Content-Type: application/json" -d "{}")
echo "Response for missing quantity field: $response"

response=$(curl -s -X POST "$API_URL" -H "Content-Type: application/json" -d "")
echo "Response for empty payload: $response"

echo -e "\nTesting with invalid HTTP method..."
response=$(curl -s -X GET "$API_URL")
echo "Response for invalid HTTP method: $response"

echo -e "\nTesting with invalid Content-Type..."
response=$(curl -s -X POST "$API_URL" -H "Content-Type: text/plain" -d "{\"quantity\": 1000}")
echo "Response for invalid Content-Type: $response"

echo -e "\nAll tests completed."
