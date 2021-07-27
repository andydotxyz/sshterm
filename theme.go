package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// termTheme is copied from the terminal repo. Hopefully we can add real theme selection in the future.
type termTheme struct {
	fyne.Theme
}

func newTermTheme() fyne.Theme {
	return &termTheme{theme.DarkTheme()}
}

// Color fixes a bug < 2.1 where theme.DarkTheme() would not override user preference.
func (t *termTheme) Color(n fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return t.Theme.Color(n, theme.VariantDark)
}

func (t *termTheme) Size(n fyne.ThemeSizeName) float32 {
	if n == theme.SizeNameText {
		return 12
	}

	return t.Theme.Size(n)
}
