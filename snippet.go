package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/aquilax/truncate"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"
)

// default values for empty state.
const (
	defaultSnippetFolder   = "folder"
	defaultLanguage        = "md"
	defaultSnippetName     = "Example"
	defaultSnippetFileName = defaultSnippetName + "." + defaultLanguage

	Day   = 24 * time.Hour
	Week  = 7 * Day
	Month = 30 * Day
	Year  = 12 * Month
)

// defaultSnippet is a snippet with all the default values, used for when
// there are no snippets available.
var defaultSnippet = Snippet{
	Name:     defaultSnippetName,
	Folder:   defaultSnippetFolder,
	Language: defaultLanguage,
	File:     defaultSnippetFileName,
	Date:     time.Now(),
}

// Snippet represents a snippet of code in a language.
// It is nested within a folder and can be tagged with metadata.
type Snippet struct {
	Folder   string    `json:"folder"`
	Date     time.Time `json:"date"`
	Name     string    `json:"title"`
	File     string    `json:"file"`
	Language string    `json:"language"`
}

// SnippetsWrapper represents the root JSON structure that contains the snippet list
type SnippetsWrapper struct {
	SnippetList []Snippet `json:"snippet_list"`
}

// String returns the folder/name.ext of the snippet.
func (s Snippet) String() string {
	return fmt.Sprintf("%s/%s.%s", s.Folder, s.Name, s.Language)
}

// LegacyPath returns the legacy path <folder>-<file>
func (s Snippet) LegacyPath() string {
	return s.File
}

// Path returns the path <folder>/<file>
func (s Snippet) Path() string {
	return filepath.Join(s.Folder, s.File)
}

// Content returns the snippet contents.
func (s Snippet) Content(highlight bool) string {
	config := readConfig()
	file := filepath.Join(config.Home, s.Path())
	content, err := os.ReadFile(file)
	if err != nil {
		return ""
	}

	if !highlight {
		return string(content)
	}

	var b bytes.Buffer
	err = quick.Highlight(&b, string(content), s.Language, "terminal16m", config.CodeBlockTheme)
	if err != nil {
		return string(content)
	}
	return b.String()
}

// FilterValue is the snippet filter value that can be used when searching.
func (s Snippet) FilterValue() string {
	return s.Folder + "/" + s.Name + "\n"
	//return s.Folder + "/" + s.Name + "\n" + "+" + strings.Join(s.Tags, "+") + "\n" + s.Language
}

// snippetDelegate represents the snippet list item.
type snippetDelegate struct {
	pane   pane
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

	titleStyle := d.styles.SelectedItemTitle
	descStyle := d.styles.SelectedItemDesc
	if d.state == copyingState && d.pane == snippetPane {
		titleStyle = d.styles.CopiedItemTitle
		descStyle = d.styles.CopiedItemDesc
	}

	if index == m.Index() {
		_, _ = fmt.Fprintln(w, "  "+titleStyle.Render(truncate.Truncate(s.Name, 30, "...", truncate.PositionEnd)))
		_, _ = fmt.Fprint(w, "  "+descStyle.Render(s.Folder+" • "+humanizeTime(s.Date)))
		return
	}
	_, _ = fmt.Fprintln(w, "  "+d.styles.UnselectedItemTitle.Render(truncate.Truncate(s.Name, 30, "...", truncate.PositionEnd)))
	_, _ = fmt.Fprint(w, "  "+d.styles.UnselectedItemDesc.Render(s.Folder+" • "+humanizeTime(s.Date)))
}

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

// Snippets is a wrapper for a snippets array to implement the fuzzy.Source
// interface.
type Snippets struct {
	snippets []Snippet
}

// String returns the string of the snippet at the specified position i
func (s Snippets) String(i int) string {
	return s.snippets[i].String()
}

// Len returns the length of the snippets array.
func (s Snippets) Len() int {
	return len(s.snippets)
}
