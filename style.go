package main

import (
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
)

// FoldersStyle is the style struct to handle the focusing and blurring of the
// folders pane in the application.
type FoldersStyle struct {
	Focused FoldersBaseStyle
	Blurred FoldersBaseStyle
}

// SnippetsStyle is the style struct to handle the focusing and blurring of the
// snippets pane in the application.
type SnippetsStyle struct {
	Focused SnippetsBaseStyle
	Blurred SnippetsBaseStyle
}

// SectionsStyle is the style struct to handle the focusing and blurring of the
// sections pane in the application.
type SectionsStyle struct {
	Focused SectionsBaseStyle
	Blurred SectionsBaseStyle
}

// ContentStyle is the style struct to handle the focusing and blurring of the
// content pane in the application.
type ContentStyle struct {
	Focused ContentBaseStyle
	Blurred ContentBaseStyle
}

// FoldersBaseStyle holds the necessary styling for the folders pane of
// the application.
type FoldersBaseStyle struct {
	Base       lipgloss.Style
	Title      lipgloss.Style
	TitleBar   lipgloss.Style
	Selected   lipgloss.Style
	Unselected lipgloss.Style
}

// SnippetsBaseStyle holds the necessary styling for the snippets pane of
// the application.
type SnippetsBaseStyle struct {
	Base                lipgloss.Style
	Title               lipgloss.Style
	TitleBar            lipgloss.Style
	SelectedItemTitle   lipgloss.Style
	SelectedItemDesc    lipgloss.Style
	UnselectedItemTitle lipgloss.Style
	UnselectedItemDesc  lipgloss.Style
	CopiedTitleBar      lipgloss.Style
	CopiedItemTitle     lipgloss.Style
	CopiedItemDesc      lipgloss.Style
}

// SectionsBaseStyle holds the necessary styling for the sections pane of
// the application.
type SectionsBaseStyle struct {
	Base                lipgloss.Style
	Title               lipgloss.Style
	TitleBar            lipgloss.Style
	SelectedItemTitle   lipgloss.Style
	SelectedItemDesc    lipgloss.Style
	UnselectedItemTitle lipgloss.Style
	UnselectedItemDesc  lipgloss.Style
	CopiedTitleBar      lipgloss.Style
	CopiedItemTitle     lipgloss.Style
}

// ContentBaseStyle holds the necessary styling for the content pane of the
// application.
type ContentBaseStyle struct {
	Code           lipgloss.Style
	TitleBar       lipgloss.Style
	Separator      lipgloss.Style
	LineNumber     lipgloss.Style
	CopiedTitleBar lipgloss.Style
}

// Styles is the struct of all styles for the application.
type Styles struct {
	Folders  FoldersStyle
	Snippets SnippetsStyle
	Sections SectionsStyle
	Content  ContentStyle
	Glamour  map[string]ansi.StyleConfig
}

var helpStyle = lipgloss.NewStyle().Margin(0, 0, 0, 1)

// DefaultStyles is the default implementation of the styles struct for all
// styling in the application.
func DefaultStyles(config Config) Styles {
	// snippets

	snippetBase := lipgloss.NewStyle().
		Width(config.SnippetTitleBarWidth + 3).
		MarginTop(config.BaseMarginTop)

	snippetFocusedTitleBar := lipgloss.NewStyle().
		Width(config.SnippetTitleBarWidth).
		Margin(SnippetBarMargin...).
		Padding(TitlePadding...).
		Background(lipgloss.Color(config.FocusedBarBgColor)).
		Foreground(lipgloss.Color(config.FocusedBarFgColor))

	snippetBlurredTitleBar := snippetFocusedTitleBar
	snippetBlurredTitleBar = snippetBlurredTitleBar.
		Background(lipgloss.Color(config.BlurredBarBgColor)).
		Foreground(lipgloss.Color(config.BlurredBarFgColor))

	snippetSelectedItem := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		MarginLeft(config.SnippetListMarginLeft).
		Padding(0, 0, 0, 1).
		Foreground(lipgloss.Color(config.SelectedItemFgColor)).
		BorderForeground(lipgloss.Color(config.SelectedItemFgColor))

	snippetUnselectedItem := lipgloss.NewStyle().
		MarginLeft(config.SnippetListMarginLeft).
		Padding(0, 0, 0, 2).
		Foreground(lipgloss.Color(config.UnselectedItemFgColor))

	snippetCopiedTitleBar := lipgloss.NewStyle().
		Width(config.SnippetTitleBarWidth).
		Margin(SnippetBarMargin...).
		Padding(TitlePadding...).
		Background(lipgloss.Color(config.CopiedBarBgColor)).
		Foreground(lipgloss.Color(config.CopiedBarFgColor))

	snippetCopiedItem := snippetSelectedItem
	snippetCopiedItem = snippetCopiedItem.
		Foreground(lipgloss.Color(config.CopiedItemFgColor)).
		BorderForeground(lipgloss.Color(config.CopiedItemFgColor))

	// sections

	sectionBase := lipgloss.NewStyle().
		Width(config.SectionTitleBarWidth + 3).
		MarginTop(config.BaseMarginTop)

	sectionFocusedTitleBar := lipgloss.NewStyle().
		Width(config.SectionTitleBarWidth).
		Margin(SectionBarMargin...).
		Padding(TitlePadding...).
		Background(lipgloss.Color(config.FocusedBarBgColor)).
		Foreground(lipgloss.Color(config.FocusedBarFgColor))

	sectionBlurredTitleBar := sectionFocusedTitleBar
	sectionBlurredTitleBar = sectionBlurredTitleBar.
		Background(lipgloss.Color(config.BlurredBarBgColor)).
		Foreground(lipgloss.Color(config.BlurredBarFgColor))

	sectionSelectedItem := lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(lipgloss.Color(config.SelectedItemFgColor))

	sectionUnselectedItem := lipgloss.NewStyle().
		PaddingLeft(4).
		Foreground(lipgloss.Color(config.UnselectedItemFgColor))

	sectionCopiedTitleBar := lipgloss.NewStyle().
		Width(config.SectionTitleBarWidth).
		Margin(SectionBarMargin...).
		Padding(TitlePadding...).
		Background(lipgloss.Color(config.CopiedBarBgColor)).
		Foreground(lipgloss.Color(config.CopiedBarFgColor))

	sectionCopiedItem := sectionSelectedItem
	sectionCopiedItem = sectionCopiedItem.
		Foreground(lipgloss.Color(config.CopiedItemFgColor)).
		BorderForeground(lipgloss.Color(config.CopiedItemFgColor))

	// content

	contentCode := lipgloss.NewStyle().Margin(ContentCodeMargin...)

	contentFocusedTitleBar := lipgloss.NewStyle().
		Width(config.ContentTitleBarWidth).
		Margin(config.BaseMarginTop, 0, 0, 0).
		Padding(TitlePadding...).
		Background(lipgloss.Color(config.FocusedBarBgColor)).
		Foreground(lipgloss.Color(config.FocusedBarFgColor))

	contentBlurredTitleBar := contentFocusedTitleBar
	contentBlurredTitleBar = contentBlurredTitleBar.
		Background(lipgloss.Color(config.BlurredBarBgColor)).
		Foreground(lipgloss.Color(config.BlurredBarFgColor))

	contentLineNumber := lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.ContentLineNumberFgColor)).
		MarginTop(1)

	contentCopiedTitleBar := lipgloss.NewStyle().
		Width(config.ContentTitleBarWidth).
		Margin(config.BaseMarginTop, 0, 0, 0).
		Padding(TitlePadding...).
		Background(lipgloss.Color(config.CopiedBarBgColor)).
		Foreground(lipgloss.Color(config.CopiedBarFgColor))

	// custom glamour style
	glamourDarkStyle := styles.DarkStyleConfig
	glamourDarkStyle.H1 = glamourDarkStyle.H2
	glamourDarkStyle.CodeBlock.Margin = &CodeBlockMarginZero
	glamourDarkStyle.CodeBlock.StylePrimitive.BlockPrefix = config.CodeBlockPrefix + "\n"
	glamourDarkStyle.CodeBlock.StylePrimitive.BlockSuffix = config.CodeBlockSuffix + "\n"

	return Styles{
		Snippets: SnippetsStyle{
			Focused: SnippetsBaseStyle{
				Base:                snippetBase,
				TitleBar:            snippetFocusedTitleBar,
				SelectedItemTitle:   snippetSelectedItem,
				SelectedItemDesc:    snippetSelectedItem,
				UnselectedItemTitle: snippetUnselectedItem,
				UnselectedItemDesc:  snippetUnselectedItem,
				CopiedTitleBar:      snippetCopiedTitleBar,
				CopiedItemTitle:     snippetCopiedItem,
				CopiedItemDesc:      snippetCopiedItem,
			},
			Blurred: SnippetsBaseStyle{
				Base:                snippetBase,
				TitleBar:            snippetBlurredTitleBar,
				SelectedItemTitle:   snippetSelectedItem,
				SelectedItemDesc:    snippetSelectedItem,
				UnselectedItemTitle: snippetUnselectedItem,
				UnselectedItemDesc:  snippetUnselectedItem,
				CopiedTitleBar:      snippetCopiedTitleBar,
				CopiedItemTitle:     snippetCopiedItem,
				CopiedItemDesc:      snippetCopiedItem,
			},
		},
		Sections: SectionsStyle{
			Focused: SectionsBaseStyle{
				Base:                sectionBase,
				TitleBar:            sectionFocusedTitleBar,
				SelectedItemTitle:   sectionSelectedItem,
				UnselectedItemTitle: sectionUnselectedItem,
				CopiedTitleBar:      sectionCopiedTitleBar,
				CopiedItemTitle:     sectionCopiedItem,
			},
			Blurred: SectionsBaseStyle{
				Base:                sectionBase,
				TitleBar:            sectionBlurredTitleBar,
				SelectedItemTitle:   sectionSelectedItem,
				UnselectedItemTitle: sectionUnselectedItem,
				CopiedTitleBar:      sectionCopiedTitleBar,
				CopiedItemTitle:     sectionCopiedItem,
			},
		},
		Content: ContentStyle{
			Focused: ContentBaseStyle{
				Code:           contentCode,
				TitleBar:       contentFocusedTitleBar,
				LineNumber:     contentLineNumber,
				CopiedTitleBar: contentCopiedTitleBar,
			},
			Blurred: ContentBaseStyle{
				Code:           contentCode,
				TitleBar:       contentBlurredTitleBar,
				LineNumber:     contentLineNumber,
				CopiedTitleBar: contentCopiedTitleBar,
			},
		},
		Glamour: map[string]ansi.StyleConfig{
			"dark": glamourDarkStyle,
		},
	}
}
