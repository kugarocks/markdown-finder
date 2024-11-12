#!/bin/bash

# Check if version number is provided
if [ -z "$1" ]; then
  echo "Please provide a version number, e.g., ./build.sh 1.0.0"
  exit 1
fi

VERSION=$1

# Create output directory
mkdir -p build

# Function to build and zip
build_and_zip() {
  GOOS=$1
  GOARCH=$2
  OUTPUT="build/mdf_${GOOS}_${GOARCH}"

  # Build
  echo "Building for $GOOS $GOARCH..."
  GOOS=$GOOS GOARCH=$GOARCH go build -o "$OUTPUT"

  # Zip with version number in the filename
  ZIPFILE="build/mdf_version_${VERSION}_${GOOS}_${GOARCH}.zip"
  echo "Compressing $OUTPUT to $ZIPFILE..."
  zip "$ZIPFILE" "$OUTPUT"
}

# Build and zip for different architectures
build_and_zip darwin amd64
build_and_zip darwin arm64
build_and_zip linux amd64
build_and_zip linux arm64

echo "Build and compression complete for version $VERSION!"
