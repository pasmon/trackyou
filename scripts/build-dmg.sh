#!/bin/bash
set -e

# Version and Artifact Name arguments
VERSION=$1
ARCH=$2
BINARY_PATH=$3

APP_NAME="TrackYou"
DMG_NAME="trackyou_${VERSION}_darwin_${ARCH}.dmg"
ICON_SOURCE="assets/app_icon.png"
ICON_NAME="trackyou.icns"

echo "Creating DMG for ${ARCH}..."

run_or_fail() {
  local message=$1
  shift
  "$@" || {
    echo "${message}" >&2
    exit 1
  }
}

# Create a temporary directory for the DMG content
mkdir -p "dist/dmg-${ARCH}/${APP_NAME}.app/Contents/MacOS"
mkdir -p "dist/dmg-${ARCH}/${APP_NAME}.app/Contents/Resources"

# Copy the binary
cp "${BINARY_PATH}" "dist/dmg-${ARCH}/${APP_NAME}.app/Contents/MacOS/trackyou"
chmod +x "dist/dmg-${ARCH}/${APP_NAME}.app/Contents/MacOS/trackyou"

TMP_ICONSET="dist/dmg-${ARCH}/${APP_NAME}.app/Contents/Resources/AppIcon.iconset"
mkdir -p "${TMP_ICONSET}"

# These are the standard base sizes used to assemble a macOS .icns file,
# along with their generated @2x retina counterparts. The 512@2x asset
# provides the 1024px representation used by modern macOS icons.
for SIZE in 16 32 128 256 512; do
  run_or_fail "Failed to generate ${SIZE}x${SIZE} icon asset" \
    sips -z "${SIZE}" "${SIZE}" "${ICON_SOURCE}" --out "${TMP_ICONSET}/icon_${SIZE}x${SIZE}.png" >/dev/null
  DOUBLE_SIZE=$((SIZE * 2))
  run_or_fail "Failed to generate ${SIZE}x${SIZE}@2x icon asset" \
    sips -z "${DOUBLE_SIZE}" "${DOUBLE_SIZE}" "${ICON_SOURCE}" --out "${TMP_ICONSET}/icon_${SIZE}x${SIZE}@2x.png" >/dev/null
done

run_or_fail "Failed to build macOS icns file" \
  iconutil -c icns "${TMP_ICONSET}" -o "dist/dmg-${ARCH}/${APP_NAME}.app/Contents/Resources/${ICON_NAME}"
rm -rf "${TMP_ICONSET}"

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
    <key>CFBundleDisplayName</key>
    <string>${APP_NAME}</string>
    <key>CFBundleIconFile</key>
    <string>${ICON_NAME}</string>
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
