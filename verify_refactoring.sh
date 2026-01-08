#!/bin/bash
# Verification script for gowrite refactoring steps
# This script checks the current state of refactoring and reports progress

set -e
set +e  # Temporarily disable exit on error for arithmetic operations

echo "🔍 Verifying gowrite refactoring progress..."
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track refactoring completion
COMPLETED_STEPS=0
TOTAL_STEPS=10

# Step 0: Check basic build
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Basic Verification"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "📦 Building application..."
if go build -o gowrite gowrite.go 2>/dev/null; then
    echo -e "  ${GREEN}✅ Build successful${NC} ($(ls -lh gowrite | awk '{print $5}'))"
else
    echo -e "  ${RED}❌ Build failed${NC}"
    exit 1
fi
echo ""

echo "🧪 Running tests..."
if go test ./... > /dev/null 2>&1; then
    TEST_COUNT=$(go test ./... -v 2>/dev/null | grep -c "^--- PASS" || echo 0)
    echo -e "  ${GREEN}✅ All tests passed${NC} ($TEST_COUNT tests)"
else
    echo -e "  ${RED}❌ Tests failed${NC}"
    exit 1
fi
echo ""

# Check refactoring steps
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Refactoring Progress"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Step 1: State package
echo "Step 0-2: Core Packages (state, persistence, commands)"
if [ -d "state" ] && [ -f "state/state.go" ]; then
    echo -e "  ${GREEN}✅ state/ package exists${NC}"
    ((COMPLETED_STEPS++))
else
    echo -e "  ${YELLOW}⏸️  state/ package not found${NC}"
fi

if [ -d "persistence" ] && [ -f "persistence/persistence.go" ]; then
    echo -e "  ${GREEN}✅ persistence/ package exists${NC}"
    ((COMPLETED_STEPS++))
else
    echo -e "  ${YELLOW}⏸️  persistence/ package not found${NC}"
fi

if [ -d "commands" ] && [ -f "commands/commands.go" ]; then
    echo -e "  ${GREEN}✅ commands/ package exists${NC}"
    ((COMPLETED_STEPS++))
else
    echo -e "  ${YELLOW}⏸️  commands/ package not found${NC}"
fi
echo ""

# Step 3: Analysis package
echo "Step 3: Analysis Package"
if [ -d "analysis" ] && [ -f "analysis/analysis.go" ]; then
    echo -e "  ${GREEN}✅ analysis/ package exists${NC}"
    if grep -q "func CalculateReadability" analysis/analysis.go 2>/dev/null; then
        echo -e "  ${GREEN}✅ CalculateReadability moved to package${NC}"
    fi
    if grep -q "func AnalyzeTextForHemingway" analysis/analysis.go 2>/dev/null; then
        echo -e "  ${GREEN}✅ AnalyzeTextForHemingway moved to package${NC}"
    fi
    ((COMPLETED_STEPS++))
else
    echo -e "  ${YELLOW}⏸️  analysis/ package not created yet${NC}"
    if grep -q "^func CalculateReadability" gowrite.go; then
        echo -e "  ${YELLOW}→  CalculateReadability still in gowrite.go${NC}"
    fi
    if grep -q "^func AnalyzeTextForHemingway" gowrite.go; then
        echo -e "  ${YELLOW}→  AnalyzeTextForHemingway still in gowrite.go${NC}"
    fi
fi
echo ""

# Step 4: Spell check package
echo "Step 4: Spell Check Package"
if [ -d "spellcheck" ] && [ -f "spellcheck/spellcheck.go" ]; then
    echo -e "  ${GREEN}✅ spellcheck/ package exists${NC}"
    ((COMPLETED_STEPS++))
else
    echo -e "  ${YELLOW}⏸️  spellcheck/ package not created yet${NC}"
    echo -e "  ${YELLOW}→  Spell check logic still in gowrite.go${NC}"
fi
echo ""

# Step 5: UI Theme
echo "Step 5: UI Theme Package"
if [ -d "ui" ] && [ -f "ui/theme.go" ]; then
    echo -e "  ${GREEN}✅ ui/theme.go exists${NC}"
    ((COMPLETED_STEPS++))
else
    echo -e "  ${YELLOW}⏸️  ui/theme.go not created yet${NC}"
fi
echo ""

# Step 6: Modals
echo "Step 6: Modal/Dialog Package"
if [ -d "ui" ] && [ -f "ui/modals.go" ]; then
    echo -e "  ${GREEN}✅ ui/modals.go exists${NC}"
    ((COMPLETED_STEPS++))
else
    echo -e "  ${YELLOW}⏸️  ui/modals.go not created yet${NC}"
fi
echo ""

# Step 7: View management
echo "Step 7: View Management"
if [ -d "ui" ] && [ -f "ui/views.go" ]; then
    echo -e "  ${GREEN}✅ ui/views.go exists${NC}"
    ((COMPLETED_STEPS++))
else
    echo -e "  ${YELLOW}⏸️  ui/views.go not created yet${NC}"
fi
echo ""

# Step 8: Command extensions
echo "Step 8: Extended Command Handlers"
if [ -f "commands/commands.go" ]; then
    COMMAND_COUNT=$(grep -c "Registry\[" commands/commands.go || echo 0)
    if [ "$COMMAND_COUNT" -gt 5 ]; then
        echo -e "  ${GREEN}✅ Extended command handlers ($COMMAND_COUNT commands)${NC}"
        ((COMPLETED_STEPS++))
    else
        echo -e "  ${YELLOW}⏸️  Basic commands only ($COMMAND_COUNT commands)${NC}"
    fi
fi
echo ""

# Step 9: Chapter/Wiki operations
echo "Step 9: Chapter/Wiki Operations"
if [ -f "ui/chapter_ops.go" ] && [ -f "ui/wiki_ops.go" ]; then
    echo -e "  ${GREEN}✅ Chapter and Wiki operations extracted${NC}"
    ((COMPLETED_STEPS++))
else
    echo -e "  ${YELLOW}⏸️  Chapter/Wiki operations not extracted yet${NC}"
fi
echo ""

# Step 10: Input handling
echo "Step 10: Input Handling"
if [ -f "ui/input.go" ]; then
    echo -e "  ${GREEN}✅ Input handling extracted${NC}"
    ((COMPLETED_STEPS++))
else
    echo -e "  ${YELLOW}⏸️  Input handling not extracted yet${NC}"
fi
echo ""

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Calculate progress percentage
if [ "$TOTAL_STEPS" -gt 0 ]; then
    PROGRESS=$((COMPLETED_STEPS * 100 / TOTAL_STEPS))
else
    PROGRESS=0
fi

echo "Refactoring Progress: $COMPLETED_STEPS/$TOTAL_STEPS steps completed ($PROGRESS%)"
echo ""

# Line count analysis
if [ -f "gowrite.go" ]; then
    MAIN_LINES=$(wc -l < gowrite.go)
    echo "Main file size: $MAIN_LINES lines"
    
    if [ "$MAIN_LINES" -lt 500 ]; then
        echo -e "${GREEN}🎉 Excellent! Main file is well refactored${NC}"
    elif [ "$MAIN_LINES" -lt 800 ]; then
        echo -e "${GREEN}👍 Good! Main file size is reasonable${NC}"
    elif [ "$MAIN_LINES" -lt 1200 ]; then
        echo -e "${YELLOW}⚠️  Main file could be smaller${NC}"
    else
        echo -e "${YELLOW}⚠️  Main file needs more refactoring${NC}"
    fi
fi
echo ""

# Test statistics
if [ -f "gowrite_test.go" ]; then
    TEST_FUNCS=$(grep -c "^func Test" gowrite_test.go || echo 0)
    BENCH_FUNCS=$(grep -c "^func Benchmark" gowrite_test.go || echo 0)
    echo "Test functions: $TEST_FUNCS"
    echo "Benchmark functions: $BENCH_FUNCS"
fi
echo ""

# Package structure
echo "Package structure:"
PACKAGES=$(find . -maxdepth 1 -type d ! -name ".*" ! -name "." | sort)
for pkg in $PACKAGES; do
    if [ -f "$pkg/$(basename $pkg).go" ] || [ -f "$pkg/go.mod" ]; then
        FILE_COUNT=$(find "$pkg" -name "*.go" -type f | wc -l)
        if [ "$FILE_COUNT" -gt 0 ]; then
            echo "  📁 $(basename $pkg)/ ($FILE_COUNT Go files)"
        fi
    fi
done
echo ""

# Next steps
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Next Steps"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if [ "$COMPLETED_STEPS" -eq "$TOTAL_STEPS" ]; then
    echo -e "${GREEN}🎉 All refactoring steps completed!${NC}"
    echo ""
    echo "Consider:"
    echo "  1. Adding more comprehensive tests"
    echo "  2. Improving documentation"
    echo "  3. Performance optimization"
else
    echo "To continue refactoring:"
    echo "  1. Review REFACTORING_STEPS.md for detailed instructions"
    echo "  2. Start with the next uncompleted step"
    echo "  3. Run this script after each step to verify progress"
    echo "  4. Commit changes after each successful step"
fi
echo ""

echo "Manual testing:"
echo "  ./gowrite  # Test the application manually"
echo ""

exit 0
