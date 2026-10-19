package navigator

// NavigatorMsg carries the index of the currently selected tab
type NavigatorMsg struct {
	currentTab int
}

// Retrieves the currentTab index from NavigatorMsg m
func (m NavigatorMsg) GetCurrentTab() int {
	return m.currentTab
}
