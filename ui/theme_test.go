package ui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func TestMaterialTheme_Color(t *testing.T) {
	darkTheme := NewMaterialTheme(theme.VariantDark)
	lightTheme := NewMaterialTheme(theme.VariantLight)

	tests := []struct {
		name     string
		thm      fyne.Theme
		colorKey fyne.ThemeColorName
		variant  fyne.ThemeVariant
		want     color.RGBA
	}{
		{
			name:     "Dark Mode Disabled Text",
			thm:      darkTheme,
			colorKey: theme.ColorNameDisabled,
			variant:  theme.VariantDark,
			want:     color.RGBA{R: 0xAB, G: 0xAB, B: 0xAB, A: 0xFF},
		},
		{
			name:     "Light Mode Disabled Text",
			thm:      lightTheme,
			colorKey: theme.ColorNameDisabled,
			variant:  theme.VariantLight,
			want:     color.RGBA{R: 0xAD, G: 0xAD, B: 0xAD, A: 0xFF},
		},
		{
			name:     "Dark Mode Disabled Button",
			thm:      darkTheme,
			colorKey: theme.ColorNameDisabledButton,
			variant:  theme.VariantDark,
			want:     color.RGBA{R: 0x42, G: 0x42, B: 0x42, A: 0xFF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.thm.Color(tt.colorKey, tt.variant)
			r, g, b, a := got.RGBA()
			wr, wg, wb, wa := tt.want.RGBA()

			if r != wr || g != wg || b != wb || a != wa {
				t.Errorf("Color() = %v, want %v", got, tt.want)
			}
		})
	}
}
