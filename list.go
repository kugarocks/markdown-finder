package main

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"io"
	"time"
	
	"github.com/aquilax/truncate"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"
)

// FilterValue is the snippet filter value that can be used when searching.
func (s Snippet) FilterValue() string {
	return s.Folder + "/" + s.Name + "\n"
	//return s.Folder + "/" + s.Name + "\n" + "+" + strings.Join(s.Tags, "+") + "\n" + s.Language
}

// snippetDelegate represents the snippet list item.
type snippetDelegate struct {
	styles SnippetsBaseStyle
	state  state
}

// Height is the number of lines the snippet list item takes up.
func (d snippetDelegate) Height() int {
	return 2
}

// Spacing is the number of lines to insert between list items.
func (d snippetDelegate) Spacing() int {
	return 1
}

// Update is called when the list is updated.
// We use this to update the section view.
func (d snippetDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
	//return func() tea.Msg {
	//	if m.SelectedItem() == nil {
	//		return nil
	//	}
	//	return updateSectionMsg(m.SelectedItem().(Snippet))
	//}
}

// Render renders the list item for the snippet which includes the title,
// folder, and date.
func (d snippetDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	if item == nil {
		return
	}
	s, ok := item.(Snippet)
	if !ok {
		return
	}
	
	titleStyle := d.styles.SelectedTitle
	subtitleStyle := d.styles.SelectedSubtitle
	if d.state == copyingState {
		titleStyle = d.styles.CopiedTitle
		subtitleStyle = d.styles.CopiedSubtitle
	}
	
	if index == m.Index() {
		fmt.Fprintln(w, "  "+titleStyle.Render(truncate.Truncate(s.Name, 30, "...", truncate.PositionEnd)))
		fmt.Fprint(w, "  "+subtitleStyle.Render("grep -V foobar..."))
		//fmt.Fprint(w, "  "+subtitleStyle.Render(s.Folder+" • "+humanizeTime(s.Date)))
		return
	}
	fmt.Fprintln(w, "  "+d.styles.UnselectedTitle.Render(truncate.Truncate(s.Name, 30, "...", truncate.PositionEnd)))
	fmt.Fprint(w, "  "+d.styles.UnselectedSubtitle.Render("grep -V foobar..."))
	//fmt.Fprint(w, "  "+d.styles.UnselectedSubtitle.Render(s.Folder+" • "+humanizeTime(s.Date)))
}

// FilterValue is the section filter value that can be used when searching.
func (s Section) FilterValue() string {
	return s.Folder + "/" + s.File + s.Title + "\n"
}

// sectionDelegate represents the section list item.
type sectionDelegate struct {
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
	selectedItemStyle := lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	
	//titleStyle := d.styles.SelectedTitle
	//if d.state == copyingState {
	//	titleStyle = d.styles.CopiedTitle
	//}
	
	if index == m.Index() {
		//fmt.Fprintln(w, "> "+titleStyle.Render(truncate.Truncate(s.Title, 30, "...", truncate.PositionEnd)))
		fmt.Fprint(w, selectedItemStyle.Render("> "+truncate.Truncate(s.Title, 30, "...", truncate.PositionEnd)))
		return
	}
	fmt.Fprint(w, itemStyle.Render(truncate.Truncate(s.Title, 30, "...", truncate.PositionEnd)))
	//fmt.Fprintln(w, "  "+d.styles.UnselectedTitle.Render(truncate.Truncate(s.Title, 30, "...", truncate.PositionEnd)))
}

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
	fmt.Fprint(w, "  ")
	if index == m.Index() {
		fmt.Fprint(w, d.styles.Selected.Render("→ "+string(f)))
		return
	}
	fmt.Fprint(w, d.styles.Unselected.Render("  "+string(f)))
}

const (
	Day   = 24 * time.Hour
	Week  = 7 * Day
	Month = 30 * Day
	Year  = 12 * Month
)

var magnitudes = []humanize.RelTimeMagnitude{
	{D: 5 * time.Second, Format: "just now", DivBy: time.Second},
	{D: time.Minute, Format: "moments ago", DivBy: time.Second},
	{D: time.Hour, Format: "%dm %s", DivBy: time.Minute},
	{D: 2 * time.Hour, Format: "1h %s", DivBy: 1},
	{D: Day, Format: "%dh %s", DivBy: time.Hour},
	{D: 2 * Day, Format: "1d %s", DivBy: 1},
	{D: Week, Format: "%dd %s", DivBy: Day},
	{D: 2 * Week, Format: "1w %s", DivBy: 1},
	{D: Month, Format: "%dw %s", DivBy: Week},
	{D: 2 * Month, Format: "1mo %s", DivBy: 1},
	{D: Year, Format: "%dmo %s", DivBy: Month},
	{D: 18 * Month, Format: "1y %s", DivBy: 1},
	{D: 2 * Year, Format: "2y %s", DivBy: 1},
}

func humanizeTime(t time.Time) string {
	return humanize.CustomRelTime(t, time.Now(), "ago", "from now", magnitudes)
}
