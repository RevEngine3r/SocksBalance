#!/bin/bash
# SocksBalance GitHub Actions Release Trigger Script (Linux/macOS)
# This script deletes and re-pushes the 'release' tag to trigger the build workflow.

set -euo pipefail

TAG_NAME="release"

if [ $# -ge 1 ]; then
    TAG_NAME="release-$1"
fi

echo "🔖 Preparing to trigger GitHub Actions release build..."
echo "📌 Tag: $TAG_NAME"

# Delete the tag locally if it exists
echo "🗑️  Deleting local tag '$TAG_NAME' (if exists)..."
git tag -d "$TAG_NAME" 2>/dev/null || true

# Delete the tag remotely if it exists
echo "🗑️  Deleting remote tag '$TAG_NAME' (if exists)..."
git push origin --delete "$TAG_NAME" 2>/dev/null || true

# Create a new tag at the current HEAD
echo "✨ Creating new tag '$TAG_NAME' at current HEAD..."
git tag "$TAG_NAME"

# Push the tag to trigger GitHub Actions
echo "🚀 Pushing tag '$TAG_NAME' to remote..."
git push origin "$TAG_NAME"

echo ""
echo "✅ Done! GitHub Actions should now be building release binaries."
echo "📊 Check the status at: https://github.com/$(git remote get-url origin | sed -E 's|.*github.com[:/](.+)\.git|\1|')/actions"
