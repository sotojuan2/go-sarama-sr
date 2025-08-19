#!/bin/bash

echo "🚀 Quick Kafka Performance Test"
echo "==============================="
echo ""
echo "This will start producing messages at ~100 msg/s"
echo "Press Ctrl+C to stop and see results"
echo ""

# Run the quick test
go run cmd/quick_test/main.go
