package main

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	folderTitleStyle        = lipgloss.NewStyle().MarginLeft(2)
	folderItemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	folderSelectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	folderPaginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	folderHelpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	folderQuitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

// Folder represents a group of snippets in a directory.
type Folder string

// FilterValue is the searchable value for the folder.
func (f Folder) FilterValue() string {
	return string(f)
}

// folderDelegate represents a folder list item.
type folderDelegate struct{ styles FoldersBaseStyle }

// Height is the number of lines the folder list item takes up.
func (d folderDelegate) Height() int {
	return 1
}

// Spacing is the number of lines to insert between folder items.
func (d folderDelegate) Spacing() int {
	return 0
}

// Update is what is called when the folder selection is updated.
// TODO: Update the filter search for the snippets with the folder name.
func (d folderDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

// Render renders a folder list item.
func (d folderDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	f, ok := item.(Folder)
	if !ok {
		return
	}
	_, _ = fmt.Fprint(w, "  ")
	if index == m.Index() {
		_, _ = fmt.Fprint(w, d.styles.Selected.Render("→ "+string(f)))
		return
	}
	_, _ = fmt.Fprint(w, d.styles.Unselected.Render("  "+string(f)))
}

// folderSelectDelegate represents a folder list item.
type folderSelectDelegate struct{}

func (d folderSelectDelegate) Height() int                             { return 1 }
func (d folderSelectDelegate) Spacing() int                            { return 0 }
func (d folderSelectDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d folderSelectDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Folder)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, string(i))

	fn := folderItemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return folderSelectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type folderSelectModel struct {
	list    list.Model
	choice  string
	config  *Config
	quiting bool
}

func (m folderSelectModel) Init() tea.Cmd {
	return nil
}

func (m folderSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quiting = true
			return m, tea.Quit
		case "enter":
			if i, ok := m.list.SelectedItem().(Folder); ok {
				m.choice = string(i)
				m.config.FolderName = string(i)
				err := m.config.writeConfig()
				if err != nil {
					fmt.Printf("write config failed: %v\n", err)
				}
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m folderSelectModel) View() string {
	if m.choice != "" {
		return folderQuitTextStyle.Render(fmt.Sprintf("Switched to folder: %s", m.choice))
	}
	if m.quiting {
		return ""
	}
	return "\n" + m.list.View()
}

// getFolders returns a sorted list of unique folders from snippets
func getFolders(snippets []Snippet) []string {
	// collect all folders and sort
	folderSet := make(map[string]struct{})
	for _, snippet := range snippets {
		folderSet[snippet.Folder] = struct{}{}
	}

	folders := make([]string, 0, len(folderSet))
	for folder := range folderSet {
		folders = append(folders, folder)
	}
	slices.Sort(folders)

	return folders
}

func setFolder(config *Config, snippets []Snippet) error {
	folders := getFolders(snippets)

	// convert to list items
	var items []list.Item
	currentIndex := 0
	for i, folder := range folders {
		items = append(items, Folder(folder))
		if folder == config.FolderName {
			currentIndex = i
		}
	}

	// create list
	l := list.New(items, folderSelectDelegate{}, 30, 14)
	l.Title = "Choose a folder"
	l.SetShowStatusBar(false)
	l.Styles.Title = folderTitleStyle
	l.Styles.PaginationStyle = folderPaginationStyle
	l.Styles.HelpStyle = folderHelpStyle

	// set current selected item
	l.Select(currentIndex)

	// run interactive program
	p := tea.NewProgram(folderSelectModel{
		list:    l,
		config:  config,
		quiting: false,
	})

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("run interactive program failed: %w", err)
	}

	return nil
}

func listFolders(config Config, snippets []Snippet) error {
	folders := getFolders(snippets)

	// show folders list, current selected folder with arrow mark
	for _, folder := range folders {
		if folder == config.FolderName {
			fmt.Printf("%s\n", folderSelectedItemStyle.Render("> "+folder))
		} else {
			fmt.Printf("%s\n", folderItemStyle.Render(folder))
		}
	}
	return nil
}
