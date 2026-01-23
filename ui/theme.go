package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type materialTheme struct {
	variant fyne.ThemeVariant
}

var _ fyne.Theme = (*materialTheme)(nil)

// NewMaterialTheme creates a new custom material design theme
// variant: theme.VariantLight or theme.VariantDark.
func NewMaterialTheme(variant fyne.ThemeVariant) fyne.Theme {
	return &materialTheme{variant: variant}
}

func (t *materialTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	// Force the variant based on initialization
	isDark := t.variant == theme.VariantDark

	switch name {
	case theme.ColorNameBackground:
		if isDark {
			return color.RGBA{R: 0x12, G: 0x12, B: 0x12, A: 0xFF}
		}
		return color.RGBA{R: 0xFA, G: 0xFA, B: 0xFA, A: 0xFF} // Material Light Background (almost white)
	case theme.ColorNameForeground:
		if isDark {
			return color.RGBA{R: 0xE0, G: 0xE0, B: 0xE0, A: 0xFF}
		}
		return color.RGBA{R: 0x21, G: 0x21, B: 0x21, A: 0xFF}
	case theme.ColorNamePrimary:
		return color.RGBA{R: 0x62, G: 0x00, B: 0xEE, A: 0xFF} // Purple 500
	case theme.ColorNameButton:
		// Default button background
		if isDark {
			return color.RGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xFF}
		}
		return color.White
	case theme.ColorNameInputBackground:
		if isDark {
			return color.RGBA{R: 0x2C, G: 0x2C, B: 0x2C, A: 0xFF}
		}
		return color.RGBA{R: 0xF5, G: 0xF5, B: 0xF5, A: 0xFF} // Grey 100 for inputs
	case theme.ColorNamePlaceHolder:
		if isDark {
			return color.RGBA{R: 0xB0, G: 0xB0, B: 0xB0, A: 0xFF}
		}
		return color.RGBA{R: 0x9E, G: 0x9E, B: 0x9E, A: 0xFF} // Grey 500
	case theme.ColorNameSeparator:
		if isDark {
			return color.RGBA{R: 0x42, G: 0x42, B: 0x42, A: 0xFF}
		}
		return color.RGBA{R: 0xE0, G: 0xE0, B: 0xE0, A: 0xFF}
	case theme.ColorNameScrollBar:
		if isDark {
			return color.RGBA{R: 0x55, G: 0x55, B: 0x55, A: 0xAA}
		}
		return color.RGBA{R: 0xAA, G: 0xAA, B: 0xAA, A: 0xAA}
	}

	return theme.DefaultTheme().Color(name, t.variant)
}

func (t *materialTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *materialTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *materialTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 16
	case theme.SizeNamePadding:
		return 10
	case theme.SizeNameInlineIcon:
		return 24
	case theme.SizeNameScrollBar:
		return 12
	case theme.SizeNameInputRadius:
		return 4
	}
	return theme.DefaultTheme().Size(name)
}