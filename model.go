package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
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

	// 添加新字段
	hideSnippetPane bool // 控制是否隐藏 snippetPane
}

// Init initialzes the application model.
func (m *Model) Init() tea.Cmd {
	m.SectionsMap = make(map[Snippet]*list.Model)
	m.updateKeyMap()

	if m.hideSnippetPane {
		m.pane = sectionPane
		m.SnippetStyle = DefaultStyles(m.config).Snippets.Blurred
		m.SectionStyle = DefaultStyles(m.config).Sections.Focused
		m.ContentStyle = DefaultStyles(m.config).Content.Blurred
	}

	return func() tea.Msg {
		return updateContentMsg(m.selectedSection())
	}
}

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
	case updateContentMsg:
		return m.updateContentView(msg)
	case changeStateMsg:
		m.Snippets().SetDelegate(snippetDelegate{m.pane, m.SnippetStyle, msg.newState})
		m.Sections().SetDelegate(sectionDelegate{m.pane, m.SectionStyle, msg.newState})

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
			if m.hideSnippetPane {
				// 当隐藏 snippetPane 时，只在 sectionPane 和 contentPane 之间切换
				if m.pane == sectionPane {
					m.pane = contentPane
				} else {
					m.pane = sectionPane
				}
			} else {
				m.nextPane()
			}
		case key.Matches(msg, m.keys.PreviousPane):
			if m.hideSnippetPane {
				// 当隐藏 snippetPane 时，只在 sectionPane 和 contentPane 之间切换
				if m.pane == sectionPane {
					m.pane = contentPane
				} else {
					m.pane = sectionPane
				}
			} else {
				m.previousPane()
			}
		case key.Matches(msg, m.keys.Quit):
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
				var content string

				switch m.pane {
				case snippetPane:
					contentBytes, err := os.ReadFile(m.selectedSnippetFilePath())
					if err != nil {
						return changeStateMsg{navigatingState}
					}
					content = string(contentBytes)
				default:
					k := msg.String()
					index := -1
					for i, copyKey := range m.keys.CopySnippet.Keys() {
						if k == copyKey {
							index = i
							break
						}
					}
					codeBlocks := m.selectedSection().CodeBlocks
					if index >= 0 && index < len(codeBlocks) {
						content = codeBlocks[index]
					} else {
						return changeStateMsg{navigatingState}
					}
				}

				_ = clipboard.WriteAll(content)
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
	return filepath.Join(m.config.getSourcePath(), m.selectedSnippet().Path())
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
	// 保存当前选中的 section 下标
	currentIndex := m.Sections().Index()

	return tea.ExecProcess(editorCmd(m.selectedSnippetFilePath()), func(err error) tea.Msg {
		m.updateSnippetSections(m.selectedSnippet())

		// 恢复之前选中的 section 位置
		sections := m.Sections()
		if currentIndex >= 0 && currentIndex < len(sections.Items()) {
			sections.Select(currentIndex)
		}

		return updateContentMsg(m.selectedSection())
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

// updateSnippetSections updates the snippet sections
func (m *Model) updateSnippetSections(snippet Snippet) {
	// init item list
	itemList := make([]list.Item, 0)
	styles := m.SectionStyle
	delegate := sectionDelegate{m.pane, styles, navigatingState}
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
		return
	}

	sourcePath := m.config.getSourcePath()
	snippetContentBytes, err := os.ReadFile(filepath.Join(sourcePath, snippet.Path()))
	if err != nil {
		m.displayKeyHint(m.noContentHints())
		return
	}
	snippetContent := strings.TrimSpace(string(snippetContentBytes))

	if snippetContent == "" {
		m.displayKeyHint(m.noContentHints())
		return
	}

	// split snippetContent to sections
	contentParts := strings.Split(snippetContent, "\n---\n")
	sectionSlice := make([]Section, 0, len(contentParts))
	for _, content := range contentParts {
		mdElem := &MarkdownElem{}
		content = strings.TrimSpace(content)
		mdElem, err = m.parseMarkdown(content)
		if err != nil {
			continue
		}
		sectionSlice = append(sectionSlice, Section{
			Folder:     snippet.Folder,
			File:       snippet.File,
			Content:    content,
			Title:      mdElem.FirstTitle,
			CodeBlocks: mdElem.CodeBlocks,
		})
	}

	for i, sec := range sectionSlice {
		sections.InsertItem(i, list.Item(sec))
	}
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
	c = strings.ReplaceAll(c, "\t", strings.Repeat(" ", tabSpaces))
	c = m.rewriteCodeBlockPrefix(c)
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
	m.Snippets().SetDelegate(snippetDelegate{m.pane, m.SnippetStyle, m.state})
	m.Sections().SetDelegate(sectionDelegate{m.pane, m.SectionStyle, m.state})

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
	m.updateSnippetSections(snippet)
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
	snippetTitleBar := m.SnippetStyle.TitleBar.Render("Snippets")
	sectionTitleBar := m.SectionStyle.TitleBar.Render(selectedSnippet.Name)
	contentTitleBar := m.ContentStyle.Title.Render("Content")

	if m.hideSnippetPane {
		detailTitle := fmt.Sprintf("%s / %s", selectedSnippet.Folder, selectedSnippet.Name)
		sectionTitleBar = m.SectionStyle.TitleBar.Render(detailTitle)
	}

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
	} else if m.pane == contentPane {
		if m.state == copyingState {
			contentTitleBar = m.ContentStyle.CopiedTitleBar.Render("Copied")
		}
	}

	var components []string
	if !m.hideSnippetPane {
		components = append(components, m.SnippetStyle.Base.Render(snippetTitleBar+snippetList.View()))
	}
	components = append(components,
		m.SectionStyle.Base.Render(sectionTitleBar+sectionList.View()),
		lipgloss.JoinVertical(lipgloss.Top,
			contentTitleBar,
			lipgloss.JoinHorizontal(lipgloss.Left,
				m.ContentStyle.LineNumber.Render(m.LineNumbers.View()),
				m.ContentStyle.Code.Render(m.Code.View()),
			),
		),
	)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		lipgloss.JoinHorizontal(lipgloss.Left, components...),
		helpStyle.Render(m.help.View(m.keys)),
	)
}

func (m *Model) rewriteCodeBlockPrefix(code string) string {
	oldPrefix := "------------- CodeBlock -------------"
	for _, k := range m.keys.CopySnippet.Keys() {
		prefixFormat := "---------- Press %s to copy ----------"
		newPrefix := fmt.Sprintf(prefixFormat, strings.ToUpper(k))
		code = strings.Replace(code, oldPrefix, newPrefix, 1)
	}
	return code
}

type MarkdownElem struct {
	FirstTitle string
	CodeBlocks []string
}

func (m *Model) parseMarkdown(source string) (*MarkdownElem, error) {
	mdElem := &MarkdownElem{
		CodeBlocks: make([]string, 0),
	}

	// create parser
	md := goldmark.New()
	reader := text.NewReader([]byte(source))
	doc := md.Parser().Parse(reader)

	// get first title and code block
	firstTitle := ""
	var walker = func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch node := n.(type) {
		case *ast.Heading:
			if firstTitle == "" {
				title := string(node.Text(reader.Source()))
				firstTitle = title
				mdElem.FirstTitle = title
			}
		case *ast.FencedCodeBlock:
			var content bytes.Buffer
			lines := node.Lines()
			for i := 0; i < lines.Len(); i++ {
				line := lines.At(i)
				content.Write(line.Value(reader.Source()))
			}
			codeBlock := strings.TrimSuffix(content.String(), "\n")
			mdElem.CodeBlocks = append(mdElem.CodeBlocks, codeBlock)
		}
		return ast.WalkContinue, nil
	}
	if err := ast.Walk(doc, walker); err != nil {
		return nil, err
	}

	return mdElem, nil
}
