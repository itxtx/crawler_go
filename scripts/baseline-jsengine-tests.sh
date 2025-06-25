#!/bin/bash

# JavaScript Engine Baseline Testing Script
# This script runs comprehensive tests for JavaScript engine functionality
# and captures metrics for baseline performance analysis.

set -e

# Configuration
ARTIFACTS_DIR="test-artifacts"
DATE_SUFFIX=$(date +%Y%m%d-%H%M%S)
FLAKINESS_RUNS=10
TIMEOUT="30s"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Create artifacts directory
mkdir -p "$ARTIFACTS_DIR"

print_status "Starting JavaScript engine baseline testing..."

# 1. Basic test execution
print_status "Running basic JavaScript engine tests..."
if go test ./... -run TestJSEngine -v -timeout="$TIMEOUT" 2>&1 | tee "$ARTIFACTS_DIR/jsengine-baseline-$DATE_SUFFIX.log"; then
    print_status "✓ Basic tests passed"
else
    print_error "❌ Basic tests failed"
    exit 1
fi

# 2. Flakiness testing
print_status "Testing for flakiness with $FLAKINESS_RUNS runs..."
PASS_COUNT=0
FAIL_COUNT=0

for i in $(seq 1 $FLAKINESS_RUNS); do
    echo "=== Run $i ===" | tee -a "$ARTIFACTS_DIR/jsengine-flakiness-$DATE_SUFFIX.log"
    if go test ./... -run TestJSEngine -v -timeout="$TIMEOUT" >/dev/null 2>&1; then
        echo "PASS" | tee -a "$ARTIFACTS_DIR/jsengine-flakiness-$DATE_SUFFIX.log"
        ((PASS_COUNT++))
    else
        echo "FAIL" | tee -a "$ARTIFACTS_DIR/jsengine-flakiness-$DATE_SUFFIX.log"
        ((FAIL_COUNT++))
    fi
done

FLAKINESS_RATE=$((FAIL_COUNT * 100 / FLAKINESS_RUNS))
echo "Flakiness Results: $PASS_COUNT passes, $FAIL_COUNT failures ($FLAKINESS_RATE% failure rate)" | tee -a "$ARTIFACTS_DIR/jsengine-flakiness-$DATE_SUFFIX.log"

if [ $FAIL_COUNT -gt 0 ]; then
    print_warning "⚠️  Flakiness detected: $FLAKINESS_RATE% failure rate"
else
    print_status "✓ No flakiness detected"
fi

# 3. Memory and CPU profiling
print_status "Running tests with memory and CPU profiling..."
if go test github.com/itxtx/crawler_go/tests -run TestJSEngine -v \
    -memprofile="$ARTIFACTS_DIR/jsengine-mem-$DATE_SUFFIX.prof" \
    -cpuprofile="$ARTIFACTS_DIR/jsengine-cpu-$DATE_SUFFIX.prof" \
    -timeout="$TIMEOUT" 2>&1 | tee "$ARTIFACTS_DIR/jsengine-profiled-$DATE_SUFFIX.log"; then
    print_status "✓ Profiling completed"
else
    print_warning "⚠️  Profiling tests had issues"
fi

# 4. Benchmarking (if benchmark tests exist)
print_status "Running benchmarks..."
if go test ./... -run TestJSEngine -bench=. -benchmem -timeout="$TIMEOUT" 2>&1 | tee "$ARTIFACTS_DIR/jsengine-bench-$DATE_SUFFIX.log"; then
    print_status "✓ Benchmarking completed"
else
    print_warning "⚠️  Benchmarking had issues"
fi

# 5. Generate summary report
print_status "Generating summary report..."
cat > "$ARTIFACTS_DIR/jsengine-baseline-summary-$DATE_SUFFIX.md" << EOF
# JavaScript Engine Test Baseline Results

**Date**: $(date)
**Environment**: $(uname -a)
**Go Version**: $(go version)

## Test Results Summary

### Basic Tests
- Status: $([ $? -eq 0 ] && echo "✅ PASS" || echo "❌ FAIL")
- Log: jsengine-baseline-$DATE_SUFFIX.log

### Flakiness Assessment
- Runs: $FLAKINESS_RUNS
- Passes: $PASS_COUNT
- Failures: $FAIL_COUNT
- Failure Rate: $FLAKINESS_RATE%
- Status: $([ $FAIL_COUNT -eq 0 ] && echo "✅ STABLE" || echo "⚠️ FLAKY")

### Performance Profiles
- Memory Profile: jsengine-mem-$DATE_SUFFIX.prof
- CPU Profile: jsengine-cpu-$DATE_SUFFIX.prof
- Benchmark Results: jsengine-bench-$DATE_SUFFIX.log

## Issues to File
$([ $FAIL_COUNT -gt 0 ] && echo "- Flakiness detected: $FLAKINESS_RATE% failure rate (label: baseline)" || echo "- None")

## Artifacts Generated
$(ls -la "$ARTIFACTS_DIR" | grep "$DATE_SUFFIX" | awk '{print "- " $9}')

EOF

print_status "✓ Summary report generated: $ARTIFACTS_DIR/jsengine-baseline-summary-$DATE_SUFFIX.md"

# 6. Exit with appropriate code
if [ $FAIL_COUNT -gt 0 ]; then
    print_warning "Baseline testing completed with flakiness detected"
    exit 2  # Exit with code 2 to indicate flakiness
else
    print_status "✅ Baseline testing completed successfully"
    exit 0
fi
