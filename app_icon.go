package main

import (
	"fyne.io/fyne/v2"

	"trackyou/assets"
)

const appID = "com.pasmon.trackyou"

func configureApplication(myApp fyne.App) {
	myApp.SetIcon(assets.AppIcon)
	setPlatformApplicationIcon(assets.AppIcon.Content())
}
