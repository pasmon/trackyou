package assets

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed app_icon.png
var appIconContent []byte

var AppIcon = fyne.NewStaticResource("app_icon.png", appIconContent)
