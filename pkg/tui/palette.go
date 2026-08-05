package tui

import "github.com/charmbracelet/lipgloss"

var palette Palette

type Palette struct {
	border     lipgloss.Color
	accent     lipgloss.Color
	listSelect lipgloss.Color
}

func init() {
	palette = Palette{
		border:     lipgloss.Color("#303030"),
		accent:     lipgloss.Color("#8800ff"),
		listSelect: lipgloss.Color("#b690ff"),
	}
	_ = palette
}
