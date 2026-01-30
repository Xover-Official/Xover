#!/bin/bash
set -e

echo "🚀 Preparing Talos Guardian Extension Release v0.9.0..."

# Navigate to extension directory
cd vscode-extension

# 1. Install dependencies
echo "📦 Installing dependencies..."
npm install

# 2. Compile and Package
echo "🔨 Building extension..."
npm run package

# 3. Create VSIX package
echo "📦 Creating .vsix package..."
# Check if vsce is installed
if ! command -v vsce &> /dev/null; then
    echo "⚠️ 'vsce' not found. Installing globally..."
    npm install -g @vscode/vsce
fi

vsce package --out talos-guardian-0.9.0.vsix

echo "✅ Build Complete: talos-guardian-0.9.0.vsix"
echo "👉 To publish to Marketplace: vsce publish"
