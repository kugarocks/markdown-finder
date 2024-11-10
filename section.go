package main

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/aquilax/truncate"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// default values for empty state.
const (
	defaultSectionFolder  = "default"
	defaultSectionFile    = "example.md"
	defaultSectionTitle   = "No TitleBar"
	defaultSectionContent = ""

	metaKeyCopyable = "copyable"
	metaKeyTitle    = "title"
)

// Section represents a partial content of section in markdown file.
type Section struct {
	Folder     string      `json:"folder"`
	File       string      `json:"file"`
	Title      string      `json:"title"`
	Content    string      `json:"content"`
	CodeBlocks []CodeBlock `json:"code_blocks"`
}

// CodeBlock represents a code block in a section.
type CodeBlock struct {
	Content  string
	Language string
	Meta     map[string]string
}

// defaultSection is a section with all the default values, used for when
// there are no section available.
var defaultSection = Section{
	Folder:  defaultSectionFolder,
	File:    defaultSectionFile,
	Title:   defaultSectionTitle,
	Content: defaultSectionContent,
}

// String returns the folder/file#title of the section.
func (s Section) String() string {
	return fmt.Sprintf("%s/%s#%s", s.Folder, s.File, s.Title)
}

// Path returns the path <folder>/<file>
func (s Section) Path() string {
	return filepath.Join(s.Folder, s.File)
}

// Sections is a wrapper for a sections array to implement the fuzzy.Source
// interface.
type Sections struct {
	sections []Section
}

// FilterValue is the section filter value that can be used when searching.
func (s Section) FilterValue() string {
	return s.Folder + "/" + s.File + s.Title + s.Content + "\n"
}

// sectionDelegate represents the section list item.
type sectionDelegate struct {
	pane   pane
	styles SectionsBaseStyle
	state  state
}

// Height is the number of lines the section list item takes up.
func (d sectionDelegate) Height() int {
	return 1
}

// Spacing is the number of lines to insert between list items.
func (d sectionDelegate) Spacing() int {
	return 0
}

// Update is called when the list is updated.
// We use this to update the section code view.
func (d sectionDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return func() tea.Msg {
		if m.SelectedItem() == nil {
			return nil
		}
		return updateContentMsg(m.SelectedItem().(Section))
	}
}

// Render renders the list item for the section.
func (d sectionDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	if item == nil {
		return
	}
	s, ok := item.(Section)
	if !ok {
		return
	}

	itemStyle := lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle := d.styles.SelectedItemTitle

	if d.state == copyingState && d.pane == sectionPane {
		selectedItemStyle = d.styles.CopiedItemTitle
	}

	if index == m.Index() {
		_, _ = fmt.Fprint(w, selectedItemStyle.Render("> "+truncate.Truncate(s.Title, 30, "...", truncate.PositionEnd)))
	} else {
		_, _ = fmt.Fprint(w, itemStyle.Render(truncate.Truncate(s.Title, 30, "...", truncate.PositionEnd)))
	}
}

// String returns the string of the section at the specified position i
func (s Sections) String(i int) string {
	return s.sections[i].String()
}

// Len returns the length of the sections array.
func (s Sections) Len() int {
	return len(s.sections)
}
