package ui

import "github.com/charmbracelet/lipgloss"

type PaletteStruct struct {
	Fg             lipgloss.Color
	FgHighlight    lipgloss.Color
	FgMuted        lipgloss.Color
	Border         lipgloss.Color
	Accent         lipgloss.Color
	SelectedBorder lipgloss.Color
	ListSelect     lipgloss.Color
	Error          lipgloss.Color
}

var Palette PaletteStruct

func init() {

	Palette = PaletteStruct{
		Fg:             lipgloss.Color("#545454"),
		FgHighlight:    lipgloss.Color("#717171"),
		FgMuted:        lipgloss.Color("#404040"),
		Border:         lipgloss.Color("#343434"),
		SelectedBorder: lipgloss.Color("#404040"),
		Accent:         lipgloss.Color("#8800ff"),
		ListSelect:     lipgloss.Color("#b690ff"),
		Error:          lipgloss.Color("#ff5555"),
	}

	_ = Palette
}
