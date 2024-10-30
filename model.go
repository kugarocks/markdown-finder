package main

import (
	"fmt"
	"github.com/atotto/clipboard"
	"os"
	"path/filepath"
	"strings"
	"time"
	
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

const maxPane = 3

type pane int

const (
	snippetPane pane = iota
	sectionPane
	contentPane
)

type state int

const (
	navigatingState state = iota
	copyingState
	quittingState
	editingState
)

// Model represents the state of the application.
// It contains all the snippets organized in folders.
type Model struct {
	// the config map.
	config Config
	// the key map.
	keys KeyMap
	// the help model.
	help help.Model
	// the height of the terminal.
	height int
	// the working directory.
	Workdir string
	// the map of Sections to display to the user.
	SectionsMap map[Snippet]*list.Model
	// the map of Snippets to display to the user.
	SnippetsMap map[Folder]*list.Model
	// the list of Folders to display to the user.
	Folders list.Model
	// the viewport of the Code snippet.
	Code        viewport.Model
	LineNumbers viewport.Model
	// the current active pane of focus.
	pane pane
	// the current state / action of the application.
	state state
	// stying for components
	SnippetStyle SnippetsBaseStyle
	SectionStyle SectionsBaseStyle
	ContentStyle ContentBaseStyle
	
	// markdown render
	mdRender *glamour.TermRenderer
}

// Init initialzes the application model.
func (m *Model) Init() tea.Cmd {
	m.SectionsMap = make(map[Snippet]*list.Model)
	m.updateKeyMap()
	
	return func() tea.Msg {
		return updateContentMsg(m.selectedSection())
	}
}

// updateSectionMsg tells the application to update the section view with the
// given snippet.
type updateSectionMsg Snippet

// updateContentMsg tells the application to update the content view with the
// given section.
type updateContentMsg Section

// updateContent instructs the application to fetch the latest contents of the
// snippet file.
//
// This is useful after a Paste or Edit.
func (m *Model) updateContent() tea.Cmd {
	return func() tea.Msg {
		return updateContentMsg(m.selectedSection())
	}
}

// changeStateMsg tells the application to enter a different state.
type changeStateMsg struct{ newState state }

// changeState returns a Cmd to enter a different state.
func changeState(newState state) tea.Cmd {
	return func() tea.Msg {
		return changeStateMsg{newState}
	}
}

// Update updates the model based on user interaction.
func (m *Model) Update(teaMsg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := teaMsg.(type) {
	case updateSectionMsg:
		return m.updateSectionView(msg)
	case updateContentMsg:
		return m.updateContentView(msg)
	case changeStateMsg:
		if m.pane == snippetPane {
			m.Snippets().SetDelegate(snippetDelegate{m.SnippetStyle, msg.newState})
		}
		if m.pane == sectionPane {
			m.Sections().SetDelegate(sectionDelegate{m.SectionStyle, msg.newState})
		}
		
		var cmd tea.Cmd
		
		if m.state == msg.newState {
			break
		}
		
		m.state = msg.newState
		m.updateKeyMap()
		m.updateActivePane(msg)
		
		switch msg.newState {
		case copyingState:
			m.state = copyingState
			m.updateActivePane(msg)
			cmd = tea.Tick(time.Second, func(t time.Time) tea.Msg {
				return changeStateMsg{navigatingState}
			})
		default:
			// do nothing
		}
		
		m.updateKeyMap()
		m.updateActivePane(msg)
		return m, cmd
	case tea.WindowSizeMsg:
		m.height = msg.Height - 4 - m.config.MarginTop
		for _, li := range m.SnippetsMap {
			li.SetHeight(m.height)
		}
		//m.Folders.SetHeight(m.height)
		m.Code.Height = m.height
		m.LineNumbers.Height = m.height
		m.Code.Width = msg.Width - m.Snippets().Width() - m.Folders.Width() - 20
		m.LineNumbers.Width = 5
		return m, nil
	case tea.KeyMsg:
		if m.Snippets().FilterState() == list.Filtering {
			break
		}
		if m.Sections().FilterState() == list.Filtering {
			break
		}
		
		if m.state == copyingState {
			return m, changeState(navigatingState)
		}
		
		switch {
		case key.Matches(msg, m.keys.NextPane):
			m.nextPane()
		case key.Matches(msg, m.keys.PreviousPane):
			m.previousPane()
		case key.Matches(msg, m.keys.Quit):
			m.saveState()
			m.state = quittingState
			return m, tea.Quit
		case key.Matches(msg, m.keys.MoveSnippetDown):
			m.moveSnippetDown()
		case key.Matches(msg, m.keys.MoveSnippetUp):
			m.moveSnippetUp()
		case key.Matches(msg, m.keys.ToggleHelp):
			m.help.ShowAll = !m.help.ShowAll
			
			var newHeight int
			if m.help.ShowAll {
				newHeight = m.height - 1
			} else {
				newHeight = m.height
			}
			m.Snippets().SetHeight(newHeight)
			m.Folders.SetHeight(newHeight)
			m.Code.Height = newHeight
			m.LineNumbers.Height = newHeight
		case key.Matches(msg, m.keys.CopySnippet):
			return m, func() tea.Msg {
				content, err := os.ReadFile(m.selectedSnippetFilePath())
				if err != nil {
					return changeStateMsg{navigatingState}
				}
				_ = clipboard.WriteAll(string(content))
				return changeStateMsg{copyingState}
			}
		case key.Matches(msg, m.keys.EditSnippet):
			return m, m.editSnippet()
		case key.Matches(msg, m.keys.Search):
			//m.pane = sectionPane
		}
	}
	
	m.updateKeyMap()
	cmd := m.updateActivePane(teaMsg)
	return m, cmd
}

// selectedSnippetFilePath returns the file path of the snippet that is
// currently selected.
func (m *Model) selectedSnippetFilePath() string {
	return filepath.Join(m.config.Home, m.selectedSnippet().Path())
}

// nextPane sets the next pane to be active.
func (m *Model) nextPane() {
	m.pane = (m.pane + 1) % maxPane
}

// previousPane sets the previous pane to be active.
func (m *Model) previousPane() {
	m.pane--
	if m.pane < 0 {
		m.pane = maxPane - 1
	}
}

// editSnippet opens the editor with the selected snippet file path.
func (m *Model) editSnippet() tea.Cmd {
	return tea.ExecProcess(editorCmd(m.selectedSnippetFilePath()), func(err error) tea.Msg {
		return updateSectionMsg(m.selectedSnippet())
	})
}

func (m *Model) noContentHints() []keyHint {
	return []keyHint{
		{m.keys.EditSnippet, "edit contents"},
	}
}

func (m *Model) noSnippetHints() []keyHint {
	return []keyHint{
		{m.keys.Quit, "no snippet"},
	}
}

// updateSectionView updates the section view with the correct section based on
// the active snippet or display the appropriate error message / hint message.
func (m *Model) updateSectionView(msg updateSectionMsg) (tea.Model, tea.Cmd) {
	snippet := Snippet(msg)
	
	// init item list
	itemList := make([]list.Item, 0)
	styles := m.SectionStyle
	delegate := sectionDelegate{styles, navigatingState}
	sections := list.New(itemList, delegate, 25, 20)
	sections.SetShowHelp(false)
	sections.SetShowFilter(false)
	sections.SetShowTitle(false)
	sections.Styles.StatusBar = lipgloss.NewStyle().Margin(1, 2).Foreground(lipgloss.Color("240")).MaxWidth(35 - 2)
	sections.Styles.NoItems = lipgloss.NewStyle().Margin(0, 2).Foreground(lipgloss.Color("8")).MaxWidth(35 - 2)
	sections.FilterInput.Prompt = "Find: "
	sections.FilterInput.PromptStyle = styles.Title
	sections.SetStatusBarItemName("Section", "Sections")
	sections.DisableQuitKeybindings()
	sections.Styles.Title = styles.Title
	sections.Styles.TitleBar = styles.TitleBar
	
	m.SectionsMap[snippet] = &sections
	
	if len(m.Snippets().Items()) <= 0 {
		m.displayKeyHint(m.noSnippetHints())
		return m, nil
	}
	
	contentBytes, err := os.ReadFile(filepath.Join(m.config.Home, snippet.Path()))
	if err != nil {
		m.displayKeyHint(m.noContentHints())
		return m, nil
	}
	content := strings.TrimSpace(string(contentBytes))
	
	if content == "" {
		m.displayKeyHint(m.noContentHints())
		return m, nil
	}
	
	// split content to sections
	contentParts := strings.Split(content, "---")
	sectionSlice := make([]Section, 0, len(contentParts))
	for _, subContent := range contentParts {
		subContent = strings.TrimSpace(subContent)
		title := getMarkdownFirstTitle(subContent)
		if title == "" {
			title = "No Title"
		}
		sectionSlice = append(sectionSlice, Section{
			Folder:  snippet.Folder,
			File:    snippet.File,
			Title:   title,
			Content: subContent,
		})
	}
	
	for i, sec := range sectionSlice {
		sections.InsertItem(i, list.Item(sec))
	}
	
	return m, m.updateContent()
}

// updateContentView updates the content view with the correct content based on
// the active section or display the appropriate error message / hint message.
func (m *Model) updateContentView(msg updateContentMsg) (tea.Model, tea.Cmd) {
	if len(m.Snippets().Items()) <= 0 {
		m.displayKeyHint(m.noSnippetHints())
		return m, nil
	}
	
	section := Section(msg)
	c, _ := m.mdRender.Render(section.Content)
	c = strings.TrimPrefix(c, "\n")
	m.writeLineNumbers(lipgloss.Height(c))
	m.Code.SetContent(c)
	
	return m, nil
}

type keyHint struct {
	binding key.Binding
	help    string
}

// displayKeyHint updates the content viewport with instructions on the
// relevent key binding that the user should most likely press.
func (m *Model) displayKeyHint(hints []keyHint) {
	m.LineNumbers.SetContent(strings.Repeat("  ~ \n", len(hints)))
	var s strings.Builder
	for _, hint := range hints {
		s.WriteString(
			fmt.Sprintf("%s %s\n",
				m.ContentStyle.EmptyHintKey.Render(hint.binding.Help().Key),
				m.ContentStyle.EmptyHint.Render("• "+hint.help),
			))
	}
	m.Code.SetContent(s.String())
}

// displayError updates the content viewport with the error message provided.
func (m *Model) displayError(error string) {
	m.LineNumbers.SetContent(" ~ ")
	m.Code.SetContent(fmt.Sprintf("%s",
		m.ContentStyle.EmptyHint.Render(error),
	))
}

// writeLineNumbers writes the number of line numbers to the line number
// viewport.
func (m *Model) writeLineNumbers(n int) {
	var lineNumbers strings.Builder
	for i := 1; i < n; i++ {
		lineNumbers.WriteString(fmt.Sprintf("%3d\n", i))
	}
	m.LineNumbers.SetContent(lineNumbers.String() + "  ~\n")
}

const tabSpaces = 4

// updateActivePane updates the currently active pane.
func (m *Model) updateActivePane(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch m.pane {
	case snippetPane:
		m.SnippetStyle = DefaultStyles(m.config).Snippets.Focused
		m.SectionStyle = DefaultStyles(m.config).Sections.Blurred
		m.ContentStyle = DefaultStyles(m.config).Content.Blurred
		*m.Snippets(), cmd = (*m.Snippets()).Update(msg)
		cmds = append(cmds, cmd, m.updateContent())
	case sectionPane:
		m.SnippetStyle = DefaultStyles(m.config).Snippets.Blurred
		m.SectionStyle = DefaultStyles(m.config).Sections.Focused
		m.ContentStyle = DefaultStyles(m.config).Content.Blurred
		*m.Sections(), cmd = (*m.Sections()).Update(msg)
		cmds = append(cmds, cmd)
	case contentPane:
		m.SnippetStyle = DefaultStyles(m.config).Snippets.Blurred
		m.SectionStyle = DefaultStyles(m.config).Sections.Blurred
		m.ContentStyle = DefaultStyles(m.config).Content.Focused
		m.Code, cmd = m.Code.Update(msg)
		cmds = append(cmds, cmd)
		m.LineNumbers, cmd = m.LineNumbers.Update(msg)
		cmds = append(cmds, cmd)
	}
	//m.Snippets().SetDelegate(snippetDelegate{m.SnippetStyle, m.state})
	//m.Sections().SetDelegate(sectionDelegate{m.SectionStyle, m.state})
	
	return tea.Batch(cmds...)
}

// updateKeyMap disables or enables the keys based on the current state of the
// snippet list.
func (m *Model) updateKeyMap() {
	hasItems := len(m.Snippets().VisibleItems()) > 0
	isFiltering := m.Snippets().FilterState() == list.Filtering
	isEditing := m.state == editingState
	m.keys.EditSnippet.SetEnabled(hasItems && !isFiltering && !isEditing)
}

// selected folder returns the currently selected folder.
func (m *Model) selectedFolder() Folder {
	item := m.Folders.SelectedItem()
	if item == nil {
		return "misc"
	}
	return item.(Folder)
}

// selectedSnippet returns the currently selected snippet.
func (m *Model) selectedSnippet() Snippet {
	item := m.Snippets().SelectedItem()
	if item == nil {
		return defaultSnippet
	}
	return item.(Snippet)
}

// selectedSection returns the currently selected section.
func (m *Model) selectedSection() Section {
	item := m.Sections().SelectedItem()
	if item == nil {
		return defaultSection
	}
	return item.(Section)
}

// Snippets returns the active list.
func (m *Model) Snippets() *list.Model {
	return m.SnippetsMap[m.selectedFolder()]
}

// Sections returns the active list.
func (m *Model) Sections() *list.Model {
	snippet := m.selectedSnippet()
	if sections, ok := m.SectionsMap[snippet]; ok {
		return sections
	}
	m.updateSectionView(updateSectionMsg(snippet))
	return m.SectionsMap[snippet]
}

func (m *Model) moveSnippetDown() {
	currentPosition := m.Snippets().Index()
	currentItem := m.Snippets().SelectedItem()
	m.Snippets().InsertItem(currentPosition+2, currentItem)
	m.Snippets().RemoveItem(currentPosition)
	m.Snippets().CursorDown()
}

func (m *Model) moveSnippetUp() {
	currentPosition := m.Snippets().Index()
	currentItem := m.Snippets().SelectedItem()
	m.Snippets().RemoveItem(currentPosition)
	m.Snippets().InsertItem(currentPosition-1, currentItem)
	m.Snippets().CursorUp()
}

// View returns the view string for the application model.
func (m *Model) View() string {
	if m.state == quittingState {
		return ""
	}
	
	snippetList := m.Snippets()
	sectionList := m.Sections()
	selectedSnippet := m.selectedSnippet()
	selectedSection := m.selectedSection()
	snippetTitleBar := m.SnippetStyle.TitleBar.Render("Snippets")
	sectionTitleBar := m.SectionStyle.TitleBar.Render(selectedSnippet.Name)
	contentTitleBar := m.ContentStyle.Title.Render(selectedSection.Title)
	
	if m.pane == snippetPane {
		if m.state == copyingState {
			snippetTitleBar = m.SnippetStyle.CopiedTitleBar.Render("Copied")
		} else if snippetList.SettingFilter() {
			snippetTitleBar = m.SnippetStyle.TitleBar.Render(snippetList.FilterInput.View())
		}
	} else if m.pane == sectionPane {
		if m.state == copyingState {
			sectionTitleBar = m.SectionStyle.CopiedTitleBar.Render("Copied")
		} else if sectionList.SettingFilter() {
			sectionTitleBar = m.SectionStyle.TitleBar.Render(sectionList.FilterInput.View())
		}
	}
	
	return lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			m.SnippetStyle.Base.Render(snippetTitleBar+snippetList.View()),
			m.SectionStyle.Base.Render(sectionTitleBar+sectionList.View()),
			lipgloss.JoinVertical(lipgloss.Top,
				contentTitleBar,
				lipgloss.JoinHorizontal(lipgloss.Left,
					m.ContentStyle.LineNumber.Render(m.LineNumbers.View()),
					m.ContentStyle.Code.Render(strings.ReplaceAll(m.Code.View(), "\t", strings.Repeat(" ", tabSpaces))),
				),
			),
		),
		helpStyle.Render(m.help.View(m.keys)),
	)
}

func (m *Model) saveState() {
	s := State{
		CurrentFolder:  string(m.selectedFolder()),
		CurrentSnippet: m.selectedSnippet().File,
	}
	err := s.Save()
	if err != nil {
		panic(err.Error())
	}
}
