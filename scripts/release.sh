#!/bin/bash
set -e

# Release script for lit
# Tags the lit module with the given version.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

MODULE="github.com/tracewayapp/lit/v2"

usage() {
    echo "Usage: $0 <version>"
    echo ""
    echo "Arguments:"
    echo "  version    Semantic version (e.g., v0.3.0)"
    echo ""
    echo "Tags and pushes a release for $MODULE"
    echo ""
    echo "Examples:"
    echo "  $0 v0.3.0"
    exit 1
}

# --- Step 1: Pre-flight checks ---

if [ -z "$1" ]; then
    echo -e "${RED}Error: Version argument required${NC}"
    usage
fi

VERSION="$1"

# Validate semver format (vX.Y.Z or vX.Y.Z-suffix)
if ! [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$ ]]; then
    echo -e "${RED}Error: Invalid version format '$VERSION'${NC}"
    echo "Version must be in semver format: vX.Y.Z (e.g., v1.0.0)"
    exit 1
fi

cd "$PROJECT_ROOT"

echo -e "${GREEN}=== Pre-flight checks ===${NC}"

# Check for uncommitted changes
if [ -n "$(git status --porcelain)" ]; then
    echo -e "${RED}Error: You have uncommitted changes${NC}"
    echo "Please commit all changes before releasing."
    git status --short
    exit 1
fi
echo -e "${GREEN}No uncommitted changes${NC}"

# Build and vet lit module
echo -e "${YELLOW}Building and vetting lit...${NC}"
(cd "$PROJECT_ROOT" && go build ./... && go vet ./...)
echo -e "${GREEN}lit OK${NC}"

echo ""

# --- Step 2: Tag and push ---

TAG="${VERSION}"

echo -e "${GREEN}=== Releasing $MODULE $VERSION ===${NC}"

git tag -a "$TAG" -m "Release $MODULE $VERSION"
echo -e "${GREEN}Created tag: $TAG${NC}"

git push origin "$TAG"
echo -e "${GREEN}Pushed tag: $TAG${NC}"

echo ""

# --- Step 3: Summary ---

echo -e "${GREEN}=== Release Complete ===${NC}"
echo ""
echo "Tag created: $TAG"
echo ""
echo "Install:"
echo "  go get $MODULE@$VERSION"
