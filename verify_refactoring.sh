#!/bin/bash
# Verification script for gowrite refactoring steps

set -e

echo "🔍 Verifying gowrite refactoring..."
echo ""

# Step 1: Build check
echo "✓ Step 1: Building application..."
go build -o gowrite gowrite.go
if [ -f gowrite ]; then
    echo "  ✅ Build successful ($(ls -lh gowrite | awk '{print $5}'))"
else
    echo "  ❌ Build failed"
    exit 1
fi
echo ""

# Step 2: Run tests
echo "✓ Step 2: Running unit tests..."
go test -v
if [ $? -eq 0 ]; then
    echo "  ✅ All tests passed"
else
    echo "  ❌ Tests failed"
    exit 1
fi
echo ""

# Step 3: Run benchmarks
echo "✓ Step 3: Running benchmarks..."
go test -bench=. -benchmem | grep -E "Benchmark|PASS"
echo "  ✅ Benchmarks completed"
echo ""

# Step 4: Test coverage
echo "✓ Step 4: Checking test coverage..."
go test -cover
echo ""

# Step 5: Verify extracted functions exist
echo "✓ Step 5: Verifying extracted functions..."
if grep -q "^func CalculateReadability" gowrite.go; then
    echo "  ✅ CalculateReadability extracted"
else
    echo "  ❌ CalculateReadability not found"
    exit 1
fi

if grep -q "^func AnalyzeTextForHemingway" gowrite.go; then
    echo "  ✅ AnalyzeTextForHemingway extracted"
else
    echo "  ❌ AnalyzeTextForHemingway not found"
    exit 1
fi
echo ""

# Step 6: Verify constants extracted
echo "✓ Step 6: Verifying constants..."
if grep -q "^const TargetWidth" gowrite.go; then
    echo "  ✅ TargetWidth constant extracted"
else
    echo "  ❌ TargetWidth not found"
    exit 1
fi

if grep -q "ViewMain = iota" gowrite.go; then
    echo "  ✅ View constants extracted"
else
    echo "  ❌ View constants not found"
    exit 1
fi
echo ""

echo "🎉 All verification checks passed!"
echo ""
echo "Summary:"
echo "  - Constants extracted: ViewMain, ViewNotes, ViewAnalyze, ViewWiki, TargetWidth"
echo "  - Functions extracted: CalculateReadability, AnalyzeTextForHemingway"
echo "  - Tests created: $(grep -c "^func Test" gowrite_test.go) test functions"
echo "  - Benchmarks created: $(grep -c "^func Benchmark" gowrite_test.go) benchmark functions"
echo ""
echo "Next steps:"
echo "  1. Test the application manually: ./gowrite"
echo "  2. Proceed with Step 3: Extract File I/O Functions"
