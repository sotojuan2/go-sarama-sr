#!/bin/bash

echo "🚀 Kafka Performance Test Suite"
echo "==============================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if .env file exists
if [ ! -f .env ]; then
    echo -e "${RED}❌ .env file not found!${NC}"
    echo "Please create a .env file with your Kafka configuration."
    exit 1
fi

echo -e "${BLUE}🔧 Building performance test...${NC}"
cd cmd/performance_test
go build -o performance_test main.go

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Build failed!${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Build successful!${NC}"
echo -e "${YELLOW}⚡ Starting performance tests...${NC}"
echo ""
echo "This will run 4 different performance tests:"
echo "1. Conservative: ~10 msg/s"  
echo "2. Moderate: ~40 msg/s"
echo "3. Aggressive: ~400 msg/s"
echo "4. Maximum: ~8000 msg/s"
echo ""
echo "Press Ctrl+C at any time to stop."
echo ""

# Run the performance test
./performance_test

echo -e "${GREEN}🎉 Performance testing completed!${NC}"
