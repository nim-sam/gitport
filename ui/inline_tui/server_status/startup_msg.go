package serverstatus

// UI messages pushed by the cli package as the server changes state.

// RunningMsg reports that the server is bound and accepting connections.
type RunningMsg struct{}

// StoppedMsg reports that the server stopped. A nil Err is a clean shutdown
// and closes the TUI; a non-nil Err leaves the TUI up so the failure is
// readable.
type StoppedMsg struct {
	Err error
}
