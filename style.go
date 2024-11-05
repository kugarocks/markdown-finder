package main

import (
	"github.com/charmbracelet/bubbles/list"
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
	EmptyHint      lipgloss.Style
	EmptyHintKey   lipgloss.Style
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
	white := lipgloss.Color(config.WhiteColor)
	gray := lipgloss.Color(config.GrayColor)
	brightBlack := lipgloss.Color(config.BlackColor)
	green := lipgloss.Color(config.GreenColor)
	brightBlue := lipgloss.Color(config.PrimaryColor)
	blue := lipgloss.Color(config.PrimaryColorSubdued)

	snippetBase := lipgloss.NewStyle().
		Width(SnippetTitleBarWidth + 3).
		MarginTop(config.MarginTop)

	snippetFocusedTitleBar := lipgloss.NewStyle().
		Width(SnippetTitleBarWidth).
		Margin(SnippetBarMargin...).
		Padding(TitlePadding...).
		Background(lipgloss.Color(FocusedBarBgColor)).
		Foreground(lipgloss.Color(FocusedBarFgColor))

	snippetBlurredTitleBar := snippetFocusedTitleBar
	snippetBlurredTitleBar = snippetBlurredTitleBar.
		Background(lipgloss.Color(BlurredBarBgColor)).
		Foreground(lipgloss.Color(BlurredBarFgColor))

	snippetSelectedItem := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		MarginLeft(SnippetListMarginLeft).
		Padding(0, 0, 0, 1).
		Foreground(lipgloss.Color(SelectedItemFgColor)).
		BorderForeground(lipgloss.Color(SelectedItemFgColor))

	snippetUnselectedItem := lipgloss.NewStyle().
		MarginLeft(SnippetListMarginLeft).
		Padding(0, 0, 0, 2).
		Foreground(lipgloss.Color(UnselectedItemFgColor))

	snippetCopiedTitleBar := lipgloss.NewStyle().
		Width(SnippetTitleBarWidth).
		Margin(SnippetBarMargin...).
		Padding(TitlePadding...).
		Background(lipgloss.Color(CopiedBarBgColor)).
		Foreground(lipgloss.Color(CopiedBarFgColor))

	snippetCopiedItem := snippetSelectedItem
	snippetCopiedItem = snippetCopiedItem.
		Foreground(lipgloss.Color(CopiedItemFgColor)).
		BorderForeground(lipgloss.Color(CopiedItemFgColor))

	sectionSelectedTitle := lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color(SelectedItemFgColor))
	sectionUnselectedTitle := lipgloss.NewStyle().PaddingLeft(4)
	sectionCopiedTitle := list.NewDefaultItemStyles().SelectedTitle.Foreground(lipgloss.Color(CopiedItemFgColor)).MarginLeft(1).BorderLeft(false)

	// custom glamour style
	glamourDarkStyle := styles.DarkStyleConfig
	glamourDarkStyle.H1 = glamourDarkStyle.H2
	glamourDarkStyle.CodeBlock.Margin = &CodeBlockMarginZero
	glamourDarkStyle.CodeBlock.StylePrimitive.BlockPrefix = CodeBlockPrefix + "\n"
	glamourDarkStyle.CodeBlock.StylePrimitive.BlockSuffix = CodeBlockSuffix + "\n"

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
				Base:                lipgloss.NewStyle().Width(SectionTitleBarWidth + 3).MarginTop(config.MarginTop),
				TitleBar:            lipgloss.NewStyle().Background(lipgloss.Color(FocusedBarBgColor)).Width(SectionTitleBarWidth).Margin(SectionBarMargin...).Padding(TitlePadding...).Foreground(lipgloss.Color(FocusedBarFgColor)),
				SelectedItemTitle:   sectionSelectedTitle,
				UnselectedItemTitle: sectionUnselectedTitle,
				CopiedTitleBar:      lipgloss.NewStyle().Background(green).Width(SectionTitleBarWidth).Margin(SectionBarMargin...).Padding(TitlePadding...).Foreground(white),
				CopiedItemTitle:     sectionCopiedTitle,
			},
			Blurred: SectionsBaseStyle{
				Base:                lipgloss.NewStyle().Width(SectionTitleBarWidth + 3).MarginTop(config.MarginTop),
				TitleBar:            lipgloss.NewStyle().Background(blue).Width(SectionTitleBarWidth).Margin(SectionBarMargin...).Padding(TitlePadding...).Foreground(white),
				SelectedItemTitle:   sectionSelectedTitle,
				UnselectedItemTitle: sectionUnselectedTitle,
				CopiedTitleBar:      lipgloss.NewStyle().Background(green).Width(SectionTitleBarWidth).Margin(SectionBarMargin...).Padding(TitlePadding...),
				CopiedItemTitle:     sectionCopiedTitle,
			},
		},
		Content: ContentStyle{
			Focused: ContentBaseStyle{
				Code:           lipgloss.NewStyle().Margin(ContentCodeMargin...),
				TitleBar:       lipgloss.NewStyle().Background(lipgloss.Color(FocusedBarBgColor)).Width(ContentTitleBarWidth).Margin(config.MarginTop, 0, 0, 0).Padding(TitlePadding...).Foreground(lipgloss.Color(FocusedBarFgColor)),
				LineNumber:     lipgloss.NewStyle().Foreground(brightBlack).MarginTop(1),
				EmptyHint:      lipgloss.NewStyle().Foreground(gray),
				EmptyHintKey:   lipgloss.NewStyle().Foreground(brightBlue),
				CopiedTitleBar: lipgloss.NewStyle().Background(green).Width(ContentTitleBarWidth).Margin(config.MarginTop, 0, 0, 0).Padding(TitlePadding...).Foreground(white),
			},
			Blurred: ContentBaseStyle{
				Code:           lipgloss.NewStyle().Margin(ContentCodeMargin...),
				TitleBar:       lipgloss.NewStyle().Background(blue).Width(ContentTitleBarWidth).Margin(config.MarginTop, 0, 0, 0).Padding(TitlePadding...).Foreground(white),
				LineNumber:     lipgloss.NewStyle().Foreground(brightBlack).MarginTop(1),
				EmptyHint:      lipgloss.NewStyle().Foreground(gray),
				EmptyHintKey:   lipgloss.NewStyle().Foreground(brightBlue),
				CopiedTitleBar: lipgloss.NewStyle().Background(green).Width(ContentTitleBarWidth).Margin(config.MarginTop, 0, 0, 0).Padding(TitlePadding...).Foreground(white),
			},
		},
		Glamour: map[string]ansi.StyleConfig{
			"dark": glamourDarkStyle,
		},
	}
}
