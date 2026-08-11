package tui

import "github.com/charmbracelet/lipgloss"

var palette Palette

type Palette struct {
	fg             lipgloss.Color
	fgHighlight    lipgloss.Color
	fgMuted        lipgloss.Color
	border         lipgloss.Color
	accent         lipgloss.Color
	selectedBorder lipgloss.Color
	listSelect     lipgloss.Color
}

func init() {
	palette = Palette{
		fg:             lipgloss.Color("#545454"),
		fgHighlight:    lipgloss.Color("#717171"),
		fgMuted:        lipgloss.Color("#343434"),
		border:         lipgloss.Color("#2a2a2a"),
		selectedBorder: lipgloss.Color("#404040"),
		accent:         lipgloss.Color("#8800ff"),
		listSelect:     lipgloss.Color("#b690ff"),
	}
	_ = palette
}
