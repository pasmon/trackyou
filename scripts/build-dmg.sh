#!/bin/bash
set -e

# Version and Artifact Name arguments
VERSION=$1
ARCH=$2
BINARY_PATH=$3

APP_NAME="TrackYou"
DMG_NAME="trackyou_${VERSION}_darwin_${ARCH}.dmg"

echo "Creating DMG for ${ARCH}..."

# Create a temporary directory for the DMG content
mkdir -p "dist/dmg-${ARCH}/${APP_NAME}.app/Contents/MacOS"
mkdir -p "dist/dmg-${ARCH}/${APP_NAME}.app/Contents/Resources"

# Copy the binary
cp "${BINARY_PATH}" "dist/dmg-${ARCH}/${APP_NAME}.app/Contents/MacOS/trackyou"
chmod +x "dist/dmg-${ARCH}/${APP_NAME}.app/Contents/MacOS/trackyou"

# Create a minimal Info.plist (Required for a valid .app bundle)
cat <<EOF > "dist/dmg-${ARCH}/${APP_NAME}.app/Contents/Info.plist"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>trackyou</string>
    <key>CFBundleIdentifier</key>
    <string>com.pasmon.trackyou</string>
    <key>CFBundleName</key>
    <string>${APP_NAME}</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>${VERSION}</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.13</string>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
EOF

# Create the DMG
create-dmg \
  --volname "${APP_NAME}" \
  --window-pos 200 120 \
  --window-size 600 400 \
  --app-drop-link 450 175 \
  "dist/${DMG_NAME}" \
  "dist/dmg-${ARCH}/"
