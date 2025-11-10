#!/bin/bash
set -e

# Check if the GitHub token is provided
if [ -z "$GITHUB_TOKEN" ]; then
  echo "Please set the GITHUB_TOKEN environment variable"
  echo "Example: GITHUB_TOKEN=your_github_token ./build-local.sh"
  exit 1
fi

echo "Creating git credentials file..."
# Create a credentials file with both formats
cat > git-credentials << EOF
https://x-access-token:${GITHUB_TOKEN}@github.com
https://x-access-token:${GITHUB_TOKEN}@github.com/soundcloud
EOF
chmod 600 git-credentials

echo "Building Docker image..."
DOCKER_BUILDKIT=1 docker build \
  --progress=plain \
  --secret id=git_credentials,src=git-credentials \
  -t user-settings:latest .

# Clean up credentials
rm git-credentials

echo "Build completed successfully!"