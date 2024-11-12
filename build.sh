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
  BUILD_DIR="mdf_${VERSION}_${GOOS}_${GOARCH}"
  OUTPUT="${BUILD_DIR}/mdf"

  # Build
  echo "Building for $GOOS $GOARCH..."
  GOOS=$GOOS GOARCH=$GOARCH go build -o "$OUTPUT"

  # Zip with version number in the filename
  ZIPFILE="build/mdf_${VERSION}_${GOOS}_${GOARCH}.zip"
  echo "Compressing $OUTPUT to $ZIPFILE..."
  zip "$ZIPFILE" "$OUTPUT"

  # Optionally, remove the uncompressed binary
  rm "$OUTPUT"
  rm -r "$BUILD_DIR"
}

# Build and zip for different architectures
build_and_zip darwin amd64
build_and_zip darwin arm64
build_and_zip linux amd64
build_and_zip linux arm64

echo "Build and compression complete for version $VERSION!"
