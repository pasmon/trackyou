package main

import (
	"fyne.io/fyne/v2"

	"trackyou/assets"
)

// appID should stay aligned with the macOS CFBundleIdentifier in scripts/build-dmg.sh.
const appID = "com.pasmon.trackyou"

func configureApplication(myApp fyne.App) {
	myApp.SetIcon(assets.AppIcon)
	setPlatformApplicationIcon(assets.AppIcon.Content())
}
