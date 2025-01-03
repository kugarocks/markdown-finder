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
	bkey "github.com/charmbracelet/bubbles/key"
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
	// default is true
	hideSnippetPane bool
}

// Init initialzes the application model.
func (m *Model) Init() tea.Cmd {
	m.SectionsMap = make(map[Snippet]*list.Model)
	m.pane = m.defaultPane()
	m.keys = m.config.newKeyMap()
	m.updateStyleByPane()
	m.updateKeyMap()

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
		m.height = msg.Height - 4 - m.config.BaseMarginTop
		for _, li := range m.SnippetsMap {
			li.SetHeight(m.height)
		}
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
		case bkey.Matches(msg, m.keys.NextPane):
			m.nextPane()
		case bkey.Matches(msg, m.keys.PrevPane):
			m.previousPane()
		case bkey.Matches(msg, m.keys.Quit):
			m.state = quittingState
			return m, tea.Quit
		case bkey.Matches(msg, m.keys.MoveSnippetDown):
			m.moveSnippetDown()
		case bkey.Matches(msg, m.keys.MoveSnippetUp):
			m.moveSnippetUp()
		case bkey.Matches(msg, m.keys.ToggleHelp):
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
		case bkey.Matches(msg, m.keys.CopyContent):
			if m.config.ExitAfterCopy {
				content, ok := m.getContentToCopy(msg)
				if ok {
					_ = clipboard.WriteAll(content)
				}
				m.state = quittingState
				return m, tea.Quit
			}
			return m, func() tea.Msg {
				content, ok := m.getContentToCopy(msg)
				if !ok {
					return changeStateMsg{navigatingState}
				}
				_ = clipboard.WriteAll(content)
				return changeStateMsg{copyingState}
			}
		case bkey.Matches(msg, m.keys.CopyContentExit):
			content, ok := m.getContentToCopy(msg)
			if ok {
				_ = clipboard.WriteAll(content)
			}
			m.state = quittingState
			return m, tea.Quit
		case bkey.Matches(msg, m.keys.EditSnippet):
			return m, m.editSnippet()
		case bkey.Matches(msg, m.keys.Search):
			//m.pane = sectionPane
		case bkey.Matches(msg, m.keys.ToggleSnippetPane):
			m.hideSnippetPane = !m.hideSnippetPane
			return m, nil
		}
	}

	m.updateKeyMap()
	cmd := m.updateActivePane(teaMsg)
	return m, cmd
}

// getContentToCopy
func (m *Model) getContentToCopy(msg tea.KeyMsg) (string, bool) {
	switch m.pane {
	case snippetPane:
		// copy snippet
		contentBytes, err := os.ReadFile(m.selectedSnippetFilePath())
		if err != nil {
			return "", false
		}
		return string(contentBytes), true
	default:
		// copy section code block
		key := strings.ToLower(msg.String())
		keyIndex := -1
		for i, copyKey := range m.keys.CopyContent.Keys() {
			if key == strings.ToLower(copyKey) {
				keyIndex = i
				break
			}
		}

		copyCount := 0
		codeBlocks := m.selectedSection().CodeBlocks
		for _, codeBlock := range codeBlocks {
			_, copyable := codeBlock.Meta[metaKeyCopyable]
			if copyable {
				copyCount++
				if keyIndex+1 == copyCount {
					return codeBlock.Content, true
				}
			}
		}
	}
	return "", false
}

// selectedSnippetFilePath returns the file path of the snippet that is
// currently selected.
func (m *Model) selectedSnippetFilePath() string {
	return filepath.Join(m.config.getRepoPath(), m.selectedSnippet().Path())
}

// nextPane sets the next pane to be active.
func (m *Model) nextPane() {
	m.pane = (m.pane + 1) % maxPane
	if m.hideSnippetPane && m.pane == snippetPane {
		m.pane = sectionPane
	}
}

// previousPane sets the previous pane to be active.
func (m *Model) previousPane() {
	m.pane--
	if m.pane < 0 {
		m.pane = maxPane - 1
	}
	if m.hideSnippetPane && m.pane == snippetPane {
		m.pane = contentPane
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
		return
	}

	repoPath := m.config.getRepoPath()
	snippetContentBytes, err := os.ReadFile(filepath.Join(repoPath, snippet.Path()))
	if err != nil {
		return
	}
	snippetContent := strings.TrimSpace(string(snippetContentBytes))

	if snippetContent == "" {
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
		return m, nil
	}

	section := Section(msg)
	c, _ := m.mdRender.Render(section.Content)
	c = strings.TrimPrefix(c, "\n")
	c = strings.ReplaceAll(c, "\t", strings.Repeat(" ", tabSpaces))
	c = m.handleCodeBlockBorder(c, section)
	m.writeLineNumbers(lipgloss.Height(c))
	m.Code.SetContent(c)

	return m, nil
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

	m.updateStyleByPane()

	switch m.pane {
	case snippetPane:
		*m.Snippets(), cmd = (*m.Snippets()).Update(msg)
		cmds = append(cmds, cmd, m.updateContent())
	case sectionPane:
		*m.Sections(), cmd = (*m.Sections()).Update(msg)
		cmds = append(cmds, cmd)
	case contentPane:
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
	contentTitleBar := m.ContentStyle.TitleBar.Render("Content")

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

func (m *Model) handleCodeBlockBorder(content string, section Section) string {
	s := content

	defaultBorder := m.config.CodeBlockBorderDefault
	copyKeys := m.config.CopyContentKeys
	copyKeysIndex := 0

	// handle prefix
	for _, codeBlock := range section.CodeBlocks {
		prefix := ""

		// handle copy title
		_, copyable := codeBlock.Meta[metaKeyCopyable]
		if copyable {
			if copyKeysIndex < len(copyKeys) {
				key := strings.ToUpper(copyKeys[copyKeysIndex])
				title := strings.ReplaceAll(m.config.CodeBlockTitleCopy, "{key}", key)
				prefix = m.paddingBorderWithTitle(title)
				copyKeysIndex++
			} else {
				prefix = defaultBorder
			}
			s = strings.Replace(s, m.config.CodeBlockPrefixTemp, prefix, 1)
			continue
		}

		// handle normal title
		title, titleExists := codeBlock.Meta[metaKeyTitle]
		title = strings.TrimSpace(title)
		if titleExists && len(title) > 0 {
			prefix = m.paddingBorderWithTitle(title)
		} else {
			// no title
			prefix = defaultBorder
		}

		s = strings.Replace(s, m.config.CodeBlockPrefixTemp, prefix, 1)
	}

	// handle suffix
	s = strings.ReplaceAll(s, m.config.CodeBlockSuffixTemp, defaultBorder)

	return s
}

func (m *Model) paddingBorderWithTitle(title string) string {
	maxTitleLen := m.config.CodeBlockBorderLength - 4
	if maxTitleLen <= 0 {
		return m.config.CodeBlockBorderDefault
	}
	if len(title) > maxTitleLen {
		title = title[:maxTitleLen]
	}

	paddingNum := m.config.CodeBlockBorderLength - len(title) - 2
	leftPaddingNum := paddingNum / 2
	rightPaddingNum := paddingNum - leftPaddingNum

	leftPadding := strings.Repeat(m.config.CodeBlockBorderPadding, leftPaddingNum)
	rightPadding := strings.Repeat(m.config.CodeBlockBorderPadding, rightPaddingNum)
	border := fmt.Sprintf("%s %s %s", leftPadding, title, rightPadding)

	return border
}

type MarkdownElem struct {
	FirstTitle string
	CodeBlocks []CodeBlock
}

func (m *Model) parseMarkdown(source string) (*MarkdownElem, error) {
	mdElem := &MarkdownElem{
		CodeBlocks: make([]CodeBlock, 0),
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

			// 获取代码内容
			codeContent := strings.TrimSuffix(content.String(), "\n")

			// 解析信息字符串
			info := string(node.Info.Text(reader.Source()))
			language, meta := parseCodeBlockInfo(info)

			// 创建新的代码块
			codeBlock := CodeBlock{
				Content:  codeContent,
				Language: language,
				Meta:     meta,
			}

			mdElem.CodeBlocks = append(mdElem.CodeBlocks, codeBlock)
		}
		return ast.WalkContinue, nil
	}
	if err := ast.Walk(doc, walker); err != nil {
		return nil, err
	}

	return mdElem, nil
}

// parseCodeBlockInfo Created by claude-3.5-sonnet
func parseCodeBlockInfo(info string) (string, map[string]string) {
	meta := make(map[string]string)

	// If the input is empty, return immediately
	info = strings.TrimSpace(info)
	if info == "" {
		return "", meta
	}

	// Split the language and metadata parts
	parts := strings.SplitN(info, " ", 2)
	language := parts[0]

	// If there is only a language without metadata, return immediately
	if len(parts) == 1 {
		return language, meta
	}

	// Process the metadata part
	metaStr := strings.TrimSpace(parts[1])
	metaStr = strings.Trim(metaStr, "{}")

	// Use a state machine to parse metadata
	var key, value string
	var inQuote bool
	var quoteChar rune
	var isKey = true
	var buf strings.Builder

	for _, ch := range metaStr + " " { // Add space to handle the last value
		switch {
		case ch == '"' || ch == '\'':
			if !inQuote {
				inQuote = true
				quoteChar = ch
			} else if ch == quoteChar {
				inQuote = false
			} else {
				buf.WriteRune(ch)
			}
		case ch == '=' && isKey && !inQuote:
			key = strings.TrimSpace(buf.String())
			buf.Reset()
			isKey = false
		case ch == ' ' && !inQuote:
			if !isKey {
				value = strings.TrimSpace(buf.String())
				if key != "" {
					meta[key] = value
				}
			} else if buf.Len() > 0 {
				key = strings.TrimSpace(buf.String())
				meta[key] = "true"
			}
			buf.Reset()
			isKey = true
		default:
			buf.WriteRune(ch)
		}
	}

	return language, meta
}

func (m *Model) updateStyleByPane() {
	switch m.pane {
	case snippetPane:
		m.SnippetStyle = DefaultStyles(m.config).Snippets.Focused
		m.SectionStyle = DefaultStyles(m.config).Sections.Blurred
		m.ContentStyle = DefaultStyles(m.config).Content.Blurred
	case contentPane:
		m.SnippetStyle = DefaultStyles(m.config).Snippets.Blurred
		m.SectionStyle = DefaultStyles(m.config).Sections.Blurred
		m.ContentStyle = DefaultStyles(m.config).Content.Focused
	case sectionPane:
		m.SnippetStyle = DefaultStyles(m.config).Snippets.Blurred
		m.SectionStyle = DefaultStyles(m.config).Sections.Focused
		m.ContentStyle = DefaultStyles(m.config).Content.Blurred
	}
}

func (m *Model) defaultPane() pane {
	switch m.config.DefaultPane {
	case "content":
		return contentPane
	case "snippet":
		return snippetPane
	default:
		return sectionPane
	}
}
