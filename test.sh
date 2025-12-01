#!/bin/bash

# Test script for Open-Telemorph-Prime

echo "🚀 Testing Open-Telemorph-Prime..."

# Check if the binary exists
if [ ! -f "./open-telemorph-prime" ]; then
    echo "❌ Binary not found. Building..."
    go build -o open-telemorph-prime .
    if [ $? -ne 0 ]; then
        echo "❌ Build failed"
        exit 1
    fi
fi

# Start the service in background
echo "🔄 Starting service..."
./open-telemorph-prime &
SERVICE_PID=$!

# Wait for service to start
echo "⏳ Waiting for service to start..."
sleep 3

# Test health endpoint
echo "🏥 Testing health endpoint..."
curl -s http://localhost:8080/health | jq '.' || echo "❌ Health check failed"

# Test readiness endpoint
echo "✅ Testing readiness endpoint..."
curl -s http://localhost:8080/ready | jq '.' || echo "❌ Readiness check failed"

# Test services endpoint
echo "📊 Testing services endpoint..."
curl -s http://localhost:8080/api/v1/services | jq '.' || echo "❌ Services endpoint failed"

# Test metrics endpoint
echo "📈 Testing metrics endpoint..."
curl -s http://localhost:8080/api/v1/metrics | jq '.' || echo "❌ Metrics endpoint failed"

# Test traces endpoint
echo "🔍 Testing traces endpoint..."
curl -s http://localhost:8080/api/v1/traces | jq '.' || echo "❌ Traces endpoint failed"

# Test logs endpoint
echo "📝 Testing logs endpoint..."
curl -s http://localhost:8080/api/v1/logs | jq '.' || echo "❌ Logs endpoint failed"

# Test web UI
echo "🌐 Testing web UI..."
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/ | grep -q "200" && echo "✅ Web UI accessible" || echo "❌ Web UI failed"

# Send sample trace data
echo "📤 Sending sample trace data..."
curl -X POST http://localhost:8080/v1/traces \
  -H "Content-Type: application/json" \
  -d '{
    "resourceSpans": [{
      "resource": {
        "attributes": [{
          "key": "service.name",
          "value": {"stringValue": "test-service"}
        }]
      },
      "scopeSpans": [{
        "spans": [{
          "traceId": "12345678901234567890123456789012",
          "spanId": "1234567890123456",
          "name": "test-operation",
          "startTimeUnixNano": "2024-01-01T00:00:00.000000000Z",
          "endTimeUnixNano": "2024-01-01T00:00:01.000000000Z",
          "status": {"code": "OK"},
          "attributes": [{
            "key": "http.method",
            "value": {"stringValue": "GET"}
          }]
        }]
      }]
    }]
  }' | jq '.' || echo "❌ Trace ingestion failed"

# Wait a moment for data to be processed
sleep 1

# Check if trace was stored
echo "🔍 Checking stored traces..."
curl -s http://localhost:8080/api/v1/traces | jq '.data | length' || echo "❌ No traces found"

# Cleanup
echo "🧹 Cleaning up..."
kill $SERVICE_PID 2>/dev/null
wait $SERVICE_PID 2>/dev/null

echo "✅ Test completed!"


