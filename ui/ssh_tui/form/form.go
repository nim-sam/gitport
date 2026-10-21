package form

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nim-sam/gitport/internal/model"
	"github.com/nim-sam/gitport/pkg/selector"
	"github.com/nim-sam/gitport/ui"
)

// inputGutter is the horizontal space reserved around a text input inside
// the form (field labels, container padding), so the input's viewport never
// exceeds its dedicated area.
const inputGutter = 26

// User form model struct
type Model struct {
	Inputs        []textinput.Model
	selectedInput int
	Selector      selector.Model
	Width         int
	Height        int
	buttonMsg     string
	title         string
	errMsg        string

	// requireAll rejects submission while any text field is empty. It is set
	// for the creation form so a new user always has a name, email and key.
	requireAll bool

	// accessLocked pins the Access selector. It is used when the Owner edits
	// their own record, whose access level may never change; lockedAccess is the
	// value submitted in that case.
	accessLocked bool
	lockedAccess model.AccessLevel
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {

	var cmds tea.BatchMsg

	// Adjust length of text inputs
	m.syncInputWidths()

	switch msg := msg.(type) {
	// Handle keystrokes
	case tea.KeyMsg:
		// Any keystroke dismisses a previous validation error.
		m.errMsg = ""

		switch msg.String() {
		case "up":
			m.selectedInput = m.prevFocusIndex()
		case "down":
			m.selectedInput = m.nextFocusIndex()
		case "enter":

			// Pass UserFormConfirmationMsg if "enter" is pressed
			// while the confirm/submit button is selected
			if m.confirmButtonSelected() {

				if m.requireAll {
					if err := m.missingFields(); err != "" {
						m.errMsg = err
						break
					}
				}

				user := model.User{
					Name:        m.Inputs[0].Value(),
					Email:       m.Inputs[1].Value(),
					SshKey:      strings.ReplaceAll(m.Inputs[2].Value(), "\n", ""),
					AccessLevel: m.selectedAccess(),
				}

				cmd := func() tea.Msg { return UserFormConfirmationMsg{user: user} }
				cmds = append(cmds, cmd)
			}
			m.selectedInput = m.nextFocusIndex()
		}
	}

	// Update form field selection
	for i := range len(m.Inputs) {
		if i == m.selectedInput {
			m.Inputs[i].Focus()
		} else {
			m.Inputs[i].Blur()
		}
	}

	if m.selectedInput == len(m.Inputs) && !m.accessLocked {
		m.Selector.Focused = true
	} else {
		m.Selector.Focused = false
	}

	// Update form text field values
	for i := range m.Inputs {
		m.Inputs[i], _ = m.Inputs[i].Update(msg)
	}

	// Update selector selection, unless it is locked
	if !m.accessLocked {
		m.Selector, _ = m.Selector.Update(msg)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {

	var formFieldsView []string

	// Styling form title
	titleStyle := lipgloss.NewStyle().
		Width(m.Width-2).
		Bold(true).
		Foreground(ui.Palette.FgHighlight).
		AlignHorizontal(lipgloss.Center).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(ui.Palette.Border)

	// Styling text input prompts
	promptStyle := lipgloss.NewStyle().
		Foreground(ui.Palette.Fg).
		Bold(true).
		Width(15)

	// Styling text input models
	for _, model := range m.Inputs {
		formFieldsView = append(formFieldsView, styleForView(model).View())
	}

	// Styling submit button
	focusedButton := lipgloss.NewStyle().Foreground(ui.Palette.Accent)
	blurredButton := lipgloss.NewStyle().Foreground(ui.Palette.FgMuted)

	var button lipgloss.Style
	if m.confirmButtonSelected() {
		button = focusedButton
	} else {
		button = blurredButton
	}

	// Styling Selector carrousel
	focusedSelector := lipgloss.NewStyle().Foreground(ui.Palette.Fg)
	blurredSelector := lipgloss.NewStyle().Foreground(ui.Palette.FgMuted).Italic(true)

	var Selector lipgloss.Style
	if m.selectedInput == len(m.Inputs) && !m.accessLocked {
		Selector = focusedSelector
	} else {
		Selector = blurredSelector
	}

	m.Selector.Styles.Carrousel = Selector

	// Styling Selector and text input prompt
	formPromptStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.Palette.FgHighlight)
	m.Selector.Styles.Prompt = formPromptStyle

	// The Access row renders as static text when locked, so the Owner cannot
	// cycle their own access level.
	accessView := m.Selector.View()
	if m.accessLocked {
		accessView = blurredSelector.Render(accessLevelLabel(m.lockedAccess))
	}

	for i, _ := range m.Inputs {
		m.Inputs[i].PromptStyle = formPromptStyle
	}

	// Composing the form
	titleBar := lipgloss.NewStyle().Width(m.Width).AlignHorizontal(lipgloss.Center)
	composition := lipgloss.NewStyle().Width(m.Width-2).Height(m.Height-5).Padding(0, 1).AlignVertical(lipgloss.Center).AlignHorizontal(lipgloss.Center)
	inputsContainer := lipgloss.NewStyle().Width(m.Width-8-promptStyle.GetWidth()).Height(3).Padding(0, 1)
	buttonContainer := lipgloss.NewStyle().Width(m.Width).Height(3).Align(lipgloss.Center, lipgloss.Center)

	// Submit button, with the validation error stacked above it when a
	// required field is missing. Both lines share a width so they centre
	// together within the button container.
	buttonView := button.Render(m.buttonMsg)
	if m.errMsg != "" {
		lineStyle := lipgloss.NewStyle().Width(m.Width).AlignHorizontal(lipgloss.Center)
		buttonView = lipgloss.JoinVertical(lipgloss.Top,
			lineStyle.Foreground(ui.Palette.FgMuted).Italic(true).Render(m.errMsg),
			lineStyle.Render(buttonView),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Top,
		titleBar.Render(titleStyle.Render(m.title)),
		composition.Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				promptStyle.Render(
					lipgloss.JoinVertical(
						lipgloss.Top,
						"Username",
						"Email",
						"Ssh-key",
						"Access",
					),
				),
				inputsContainer.Render(
					lipgloss.JoinVertical(lipgloss.Top,
						strings.Join(formFieldsView, "\n"),
						accessView,
					),
				),
			),
		),
		buttonContainer.Render(buttonView),
	)
}

// styleForView applies the palette to a copy of a text input so its value,
// placeholder, prompt and cursor render in the theme's colors.
func styleForView(model textinput.Model) textinput.Model {
	model.TextStyle = model.TextStyle.Foreground(ui.Palette.Fg)
	model.PlaceholderStyle = model.PlaceholderStyle.Foreground(ui.Palette.FgMuted).Italic(true)
	model.Cursor.Style = model.Cursor.Style.Foreground(ui.Palette.FgHighlight)
	// The char under a blurred cursor renders via Cursor.TextStyle, which
	// bubbles leaves at its zero value; without this it shows the terminal's
	// default color instead of the theme's text color.
	model.Cursor.TextStyle = model.TextStyle
	model.PromptStyle = model.PromptStyle.Foreground(ui.Palette.FgHighlight)
	return model
}

// SetSize applies the given dimensions to the form and propagates the input
// viewport width to every text input. Values injected via SetUserFields before
// the first Update (e.g. when the popup opens) are then windowed correctly
// even while the input is blurred, since bubbles' textinput only recomputes
// its overflow window in Update when focused.
func (m *Model) SetSize(width, height int) {
	m.Width = width
	m.Height = height
	m.syncInputWidths()
}

// syncInputWidths keeps every text input's viewport within the form layout.
func (m *Model) syncInputWidths() {
	for i := range m.Inputs {
		m.Inputs[i].Width = max(1, m.Width-inputGutter)
	}
}

// Populates input fields from user data
func (m *Model) SetUserFields(u model.User) {
	m.syncInputWidths()
	m.Inputs[0].SetValue(u.Name)
	m.Inputs[1].SetValue(u.Email)
	m.Inputs[2].SetValue(u.SshKey)

	// SetValue leaves the cursor at the end, which scrolls an overflowing
	// value (the ssh key) to its tail; start at the head instead.
	for i := range m.Inputs {
		m.Inputs[i].CursorStart()
	}

	// The selector only carries the assignable levels; the Owner level is not
	// selectable in the dashboard, so leave the selector untouched for it.
	if u.AccessLevel <= model.Admin {
		m.Selector.SetValue(int(u.AccessLevel))
	}
}

// SetAccessLocked pins the Access selector to level when locked, preventing the
// user from changing it. Pass locked=false to restore normal selection.
func (m *Model) SetAccessLocked(locked bool, level model.AccessLevel) {
	m.accessLocked = locked
	m.lockedAccess = level
}

// nextFocusIndex advances the focused field, skipping the Access selector while
// it is locked.
func (m Model) nextFocusIndex() int {
	count := len(m.Inputs) + 2 // inputs + Access selector + submit button
	next := (m.selectedInput + 1) % count
	if m.accessLocked && next == len(m.Inputs) {
		next = (next + 1) % count
	}
	return next
}

func (m Model) prevFocusIndex() int {
	count := len(m.Inputs) + 2
	prev := (m.selectedInput + count - 1) % count
	if m.accessLocked && prev == len(m.Inputs) {
		prev = (prev - 1) % count
	}
	return prev
}

// selectedAccess returns the access level the form will submit, honouring the
// lock applied to the Owner's own record.
func (m Model) selectedAccess() model.AccessLevel {
	if m.accessLocked {
		return m.lockedAccess
	}
	return model.AccessLevel(m.Selector.SelectedItem)
}

// accessLevelLabel renders an access level the way the selector does.
func accessLevelLabel(level model.AccessLevel) string {
	switch level {
	case model.ReadOnly:
		return "ReadOnly"
	case model.ReadWrite:
		return "ReadWrite"
	case model.Admin:
		return "Admin"
	case model.Owner:
		return "Owner"
	}
	return ""
}

// Returns true if the form's confirm/submit button is selected
func (m Model) confirmButtonSelected() bool {
	return m.selectedInput == len(m.Inputs)+1
}

// missingFields reports which required text fields are still empty, as a
// human-readable message, or "" when all are populated. Whitespace-only input
// counts as empty.
func (m Model) missingFields() string {
	labels := []string{"Username", "Email", "Ssh-key"}

	var missing []string
	for i, in := range m.Inputs {
		if strings.TrimSpace(in.Value()) == "" {
			missing = append(missing, labels[i])
		}
	}

	if len(missing) == 0 {
		return ""
	}
	return "Required: " + strings.Join(missing, ", ")
}

// Resets selection to first field of the form
func (m *Model) ResetSelection() {
	m.selectedInput = 0
}

func NewUserCreationFormModel() Model {

	s := selector.New("ReadOnly", "ReadWrite", "Admin")

	u := Model{
		Inputs:     make([]textinput.Model, 3),
		buttonMsg:  "[ SUBMIT ]",
		title:      "Create New User",
		Selector:   s,
		requireAll: true,
	}

	for i := range u.Inputs {

		t := textinput.New()

		switch i {
		case 0:
			t.Prompt = ""
			t.CharLimit = 32
			t.Placeholder = "Max. 32 char"
			// t.Width = u.width - 4
			t.Focus()
		case 1:
			t.Prompt = ""
			t.CharLimit = 128
			t.Placeholder = "Max. 128 char"
		case 2:
			t.Prompt = ""
			t.CharLimit = 512
			t.Placeholder = "Paste with Ctrl+v"
		}

		u.Inputs[i] = t

	}

	return u

}

func NewUserModificationFormModel() Model {

	s := selector.New("ReadOnly", "ReadWrite", "Admin")

	s.Prompt = ""

	u := Model{
		Inputs:    make([]textinput.Model, 3),
		buttonMsg: "[ CONFIRM ]",
		title:     "Edit User Information",
		Selector:  s,
	}

	for i := range u.Inputs {

		t := textinput.New()

		switch i {
		case 0:
			t.Prompt = ""
			t.CharLimit = 32
			t.Placeholder = "Username"
			// t.Width = u.width - 4
			t.Focus()
		case 1:
			t.Prompt = ""
			t.CharLimit = 128
			t.Placeholder = "Email address"
		case 2:
			t.Prompt = ""
			t.CharLimit = 512
			t.Placeholder = "SSH Key"
		}

		u.Inputs[i] = t
	}

	return u
}

func (m *Model) ClearInputs() {

	m.syncInputWidths()

	for i := range m.Inputs {
		m.Inputs[i].SetValue("")
	}

}
