#!/bin/bash
# Test script for file picker feature

echo "🧪 Testing File Picker Feature"
echo ""

# Clean up any existing test files
rm -f test*.json demo*.json 2>/dev/null

# Create test files
echo "📝 Creating test project files..."
cat > test1.json << 'EOF'
{
  "Chapters": [
    {
      "Title": "The Beginning",
      "Content": "Once upon a time...",
      "Notes": "Opening scene",
      "Target": 1000
    }
  ],
  "Wiki": [
    {
      "Title": "Main Character",
      "Content": "Name: Alice\nAge: 25"
    }
  ]
}
EOF

cat > test2.json << 'EOF'
{
  "Chapters": [
    {
      "Title": "Chapter One",
      "Content": "It was a dark and stormy night...",
      "Notes": "Set the mood",
      "Target": 1500
    }
  ],
  "Wiki": [
    {
      "Title": "Setting",
      "Content": "Victorian London"
    }
  ]
}
EOF

cat > demo_novel.json << 'EOF'
{
  "Chapters": [
    {
      "Title": "Prologue",
      "Content": "The story begins here...",
      "Notes": "Hook the reader",
      "Target": 500
    }
  ],
  "Wiki": [
    {
      "Title": "Plot",
      "Content": "Three act structure"
    }
  ]
}
EOF

echo "✅ Created 3 test files:"
ls -lh *.json | awk '{print "   - " $9 " (" $5 ")"}'
echo ""

# Build the application
echo "🔨 Building application..."
go build -o gowrite gowrite.go
if [ $? -eq 0 ]; then
    echo "✅ Build successful"
else
    echo "❌ Build failed"
    exit 1
fi
echo ""

# Run tests
echo "🧪 Running unit tests..."
go test -v -run TestCalculateReadability 2>&1 | grep -E "PASS|FAIL"
echo ""

echo "📋 File Picker Test Instructions:"
echo ""
echo "1. Run the application:"
echo "   ./gowrite"
echo ""
echo "2. Test the file picker:"
echo "   - Press Ctrl+E to open command palette"
echo "   - Type: open"
echo "   - Press Enter"
echo "   - You should see a list of 3 .json files"
echo ""
echo "3. Navigate the file picker:"
echo "   - Use ↑↓ arrow keys to select a file"
echo "   - Press Enter to open the selected file"
echo "   - Or press Esc to cancel"
echo ""
echo "4. Test direct open (existing behavior):"
echo "   - Press Ctrl+E"
echo "   - Type: open test1"
echo "   - Press Enter"
echo "   - File should open directly"
echo ""
echo "5. Test help documentation:"
echo "   - Press F1 to open help"
echo "   - Press Enter twice to see Commands page"
echo "   - Verify 'open' command shows file picker info"
echo ""

echo "🎯 Expected Behavior:"
echo "   ✓ File picker shows all .json files"
echo "   ✓ Arrow keys navigate the list"
echo "   ✓ Enter opens the selected file"
echo "   ✓ Esc cancels and returns to editor"
echo "   ✓ Direct 'open <filename>' still works"
echo ""

echo "🧹 Cleanup (after testing):"
echo "   rm test*.json demo*.json"
echo ""

echo "Ready to test! Run: ./gowrite"
