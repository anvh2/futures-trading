#!/bin/bash

# E2E Test Runner for Futures Trading System
# This script runs comprehensive end-to-end tests

set -e

echo "🚀 Futures Trading System E2E Test Runner"
echo "=========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
TEST_TIMEOUT="60s"
TEST_VERBOSE="-v"

echo -e "${BLUE}📊 Running Basic Market Data Generation Tests...${NC}"
go test ${TEST_VERBOSE} -timeout ${TEST_TIMEOUT} -run TestBasicMarketDataGeneration ./tests/e2e/
echo -e "${GREEN}✅ Basic market data tests completed${NC}"
echo

echo -e "${BLUE}🎭 Running Market Scenario Tests...${NC}"
go test ${TEST_VERBOSE} -timeout ${TEST_TIMEOUT} -run TestMarketScenarios ./tests/e2e/
echo -e "${GREEN}✅ Scenario tests completed${NC}"
echo

echo -e "${BLUE}📈 Running Trading Signal Generation Tests...${NC}"
go test ${TEST_VERBOSE} -timeout ${TEST_TIMEOUT} -run TestTradingSignalGeneration ./tests/e2e/
echo -e "${GREEN}✅ Signal generation tests completed${NC}"
echo

echo -e "${BLUE}🔍 Running Data Integrity Tests...${NC}"
go test ${TEST_VERBOSE} -timeout ${TEST_TIMEOUT} -run TestCandleDataIntegrity ./tests/e2e/
echo -e "${GREEN}✅ Data integrity tests completed${NC}"
echo

echo -e "${BLUE}💥 Running High Volatility Tests...${NC}"
go test ${TEST_VERBOSE} -timeout ${TEST_TIMEOUT} -run TestHighVolatilityScenario ./tests/e2e/
echo -e "${GREEN}✅ High volatility tests completed${NC}"
echo

echo -e "${BLUE}🎯 Running Orchestrator Integration Tests...${NC}"
echo -e "${YELLOW}Note: These tests simulate the complete orchestrator flow${NC}"
go test ${TEST_VERBOSE} -timeout ${TEST_TIMEOUT} -run TestOrchestratorIntegration ./tests/e2e/
echo -e "${GREEN}✅ Orchestrator integration tests completed${NC}"
echo

echo -e "${BLUE}🔄 Running All E2E Tests Together...${NC}"
go test ${TEST_VERBOSE} -timeout 120s ./tests/e2e/
echo -e "${GREEN}✅ All E2E tests completed successfully${NC}"
echo

echo -e "${GREEN}🎉 All E2E Tests Completed Successfully!${NC}"
echo
echo -e "${BLUE}📋 Test Summary:${NC}"
echo "  ✅ Basic Market Data Generation (TestBasicMarketDataGeneration)"
echo "  ✅ Market Scenarios (TestMarketScenarios: bull_run, bear_market, sideways_chop, flash_crash)"
echo "  ✅ Trading Signal Generation (TestTradingSignalGeneration)" 
echo "  ✅ Data Integrity Validation (TestCandleDataIntegrity)"
echo "  ✅ High Volatility Scenarios (TestHighVolatilityScenario)"
echo "  ✅ Orchestrator Integration (TestOrchestratorIntegration)"
echo
echo -e "${BLUE}🎯 Trading Pipeline Verified:${NC}"
echo "  • Market Service → Analyzer Service → Signal Service"
echo "  • Signal Service → Decision Engine → Risk Engine"  
echo "  • Risk Engine → Order Executor → State Manager → Notifier"
echo "  • Complete E2E flow: Market → Analysis → Signal → Decision → Risk → Order → State → Notify"
echo
echo -e "${BLUE}✨ Key Features Tested:${NC}"
echo "  • Realistic market data generation with trends and volatility"
echo "  • Multiple market scenarios (bull, bear, sideways, crash)"
echo "  • Trading signal generation with RSI and confidence scoring"
echo "  • Data integrity validation for OHLCV candles"
echo "  • High volatility stress testing"
echo "  • Complete orchestrator flow simulation"
echo "  • Mock implementations for external dependencies"
echo
echo -e "${YELLOW}💡 To run individual test suites:${NC}"
echo "  go test -v ./tests/e2e/ -run TestBasicMarketDataGeneration"
echo "  go test -v ./tests/e2e/ -run TestMarketScenarios"
echo "  go test -v ./tests/e2e/ -run TestOrchestratorIntegration"
echo
echo -e "${YELLOW}💡 To run with race detection:${NC}"
echo "  go test -race -v ./tests/e2e/"
echo
echo -e "${YELLOW}💡 Test Data Location:${NC}"
echo "  tests/testdata/market_generator.go - Market data generation"
echo "  tests/testdata/scenarios.go - Trading scenarios"
echo "  tests/testdata/config.go - Test configuration"
echo "  tests/e2e/mocks.go - Mock implementations"
echo