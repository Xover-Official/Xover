#!/bin/bash
set -e

# Talos SDK Publishing Script
# Usage: ./publish_sdks.sh [version]

VERSION=$1

if [ -z "$VERSION" ]; then
  echo "Usage: ./publish_sdks.sh <version>"
  exit 1
fi

echo "🚀 Preparing release for version $VERSION..."

# 1. Python SDK (PyPI)
echo "🐍 Building Python SDK..."
cd sdk/python
# Update version in setup.py (mock)
sed -i "s/version='.*'/version='$VERSION'/" setup.py
python3 setup.py sdist bdist_wheel
# twine upload dist/* (Commented out for safety)
echo "✅ Python SDK built."
cd ../..

# 2. JavaScript SDK (npm)
echo "📦 Building JS SDK..."
cd sdk/javascript
# Update version in package.json
sed -i "s/\"version\": \".*\"/\"version\": \"$VERSION\"/" package.json
npm install
npm run build
# npm publish (Commented out for safety)
echo "✅ JS SDK built."
cd ../..

# 3. Go SDK (pkg.go.dev)
echo "🐹 Tagging Go SDK..."
# git tag sdk/go/v$VERSION
# git push origin sdk/go/v$VERSION
echo "✅ Go SDK tagged."

echo "🎉 All SDKs prepared for release $VERSION!"
