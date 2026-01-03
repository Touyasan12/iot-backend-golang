#!/bin/bash

# Test Pagination Implementation
# This script tests all paginated endpoints

API_URL="http://localhost:8080/api/v1"

echo "🧪 Testing Pagination Implementation"
echo "===================================="
echo ""

# Test 1: History endpoint
echo "1️⃣  Testing GET /history with pagination..."
echo "   Request: GET $API_URL/history?page=1&page_size=10"
curl -s "$API_URL/history?page=1&page_size=10" | jq -c '{data_count: (.data | length), pagination}'
echo ""

# Test 2: History with filters
echo "2️⃣  Testing GET /history with filters..."
echo "   Request: GET $API_URL/history?device_type=FEEDER&page=1&page_size=5"
curl -s "$API_URL/history?device_type=FEEDER&page=1&page_size=5" | jq -c '{data_count: (.data | length), pagination}'
echo ""

# Test 3: Feeder schedules
echo "3️⃣  Testing GET /feeder/schedules with pagination..."
echo "   Request: GET $API_URL/feeder/schedules?page=1&page_size=5"
curl -s "$API_URL/feeder/schedules?page=1&page_size=5" | jq -c '{data_count: (.data | length), pagination}'
echo ""

# Test 4: UV schedules
echo "4️⃣  Testing GET /uv/schedules with pagination..."
echo "   Request: GET $API_URL/uv/schedules?page=1&page_size=5"
curl -s "$API_URL/uv/schedules?page=1&page_size=5" | jq -c '{data_count: (.data | length), pagination}'
echo ""

# Test 5: Default pagination (no params)
echo "5️⃣  Testing default pagination (no params)..."
echo "   Request: GET $API_URL/history"
curl -s "$API_URL/history" | jq -c '{data_count: (.data | length), pagination: {page: .pagination.page, page_size: .pagination.page_size}}'
echo ""

# Test 6: Page 2
echo "6️⃣  Testing page 2..."
echo "   Request: GET $API_URL/history?page=2&page_size=20"
curl -s "$API_URL/history?page=2&page_size=20" | jq -c '{data_count: (.data | length), pagination}'
echo ""

# Test 7: Large page_size (should be capped)
echo "7️⃣  Testing max page_size limit..."
echo "   Request: GET $API_URL/history?page=1&page_size=1000 (should be capped to 200)"
curl -s "$API_URL/history?page=1&page_size=1000" | jq -c '{pagination: {page: .pagination.page, page_size: .pagination.page_size}}'
echo ""

# Test 8: Invalid page (0)
echo "8️⃣  Testing invalid page (0, should default to 1)..."
echo "   Request: GET $API_URL/history?page=0&page_size=10"
curl -s "$API_URL/history?page=0&page_size=10" | jq -c '{pagination: {page: .pagination.page}}'
echo ""

echo ""
echo "✅ Pagination tests completed!"
echo ""
echo "📊 Response Format:"
echo "   All paginated endpoints return:"
echo "   {"
echo "     \"data\": [...],         // Array of items"
echo "     \"pagination\": {        // Metadata"
echo "       \"page\": 1,"
echo "       \"page_size\": 20,"
echo "       \"total\": 156,"
echo "       \"total_pages\": 8"
echo "     }"
echo "   }"
