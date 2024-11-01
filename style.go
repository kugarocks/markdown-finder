package main

import (
	"github.com/charmbracelet/bubbles/list"
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
	Base               lipgloss.Style
	Title              lipgloss.Style
	TitleBar           lipgloss.Style
	SelectedSubtitle   lipgloss.Style
	UnselectedSubtitle lipgloss.Style
	SelectedTitle      lipgloss.Style
	UnselectedTitle    lipgloss.Style
	CopiedTitleBar     lipgloss.Style
	CopiedTitle        lipgloss.Style
	CopiedSubtitle     lipgloss.Style
}

// SectionsBaseStyle holds the necessary styling for the sections pane of
// the application.
type SectionsBaseStyle struct {
	Base               lipgloss.Style
	Title              lipgloss.Style
	TitleBar           lipgloss.Style
	SelectedSubtitle   lipgloss.Style
	UnselectedSubtitle lipgloss.Style
	SelectedTitle      lipgloss.Style
	UnselectedTitle    lipgloss.Style
	CopiedTitleBar     lipgloss.Style
	CopiedTitle        lipgloss.Style
}

// ContentBaseStyle holds the necessary styling for the content pane of the
// application.
type ContentBaseStyle struct {
	Code           lipgloss.Style
	Title          lipgloss.Style
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
}

var helpStyle = lipgloss.NewStyle().Margin(0, 0, 0, 1)

// DefaultStyles is the default implementation of the styles struct for all
// styling in the application.
func DefaultStyles(config Config) Styles {
	white := lipgloss.Color(config.WhiteColor)
	gray := lipgloss.Color(config.GrayColor)
	//black := lipgloss.Color(config.BackgroundColor)
	brightBlack := lipgloss.Color(config.BlackColor)
	green := lipgloss.Color(config.GreenColor)
	brightGreen := lipgloss.Color(config.BrightGreenColor)
	brightBlue := lipgloss.Color(config.PrimaryColor)
	blue := lipgloss.Color(config.PrimaryColorSubdued)
	
	defaultItemStyle := list.NewDefaultItemStyles()
	sectionBarWidth := 36
	contentBarWidth := 86
	snippetListMarginLeft := 1
	
	sectionSelectedTitle := lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	sectionUnselectedTitle := lipgloss.NewStyle().PaddingLeft(4)
	sectionCopiedTitle := list.NewDefaultItemStyles().SelectedTitle.Foreground(brightGreen).MarginLeft(1).BorderLeft(false)
	
	return Styles{
		Snippets: SnippetsStyle{
			Focused: SnippetsBaseStyle{
				Base:               lipgloss.NewStyle().Width(sectionBarWidth).MarginTop(config.MarginTop),
				TitleBar:           lipgloss.NewStyle().Background(lipgloss.Color("62")).Width(35-2).Margin(0, 1, 1, 2).Padding(0, 1).Foreground(lipgloss.Color("230")),
				SelectedSubtitle:   defaultItemStyle.SelectedDesc.MarginLeft(snippetListMarginLeft),
				UnselectedSubtitle: defaultItemStyle.DimmedDesc.MarginLeft(snippetListMarginLeft),
				SelectedTitle:      defaultItemStyle.SelectedTitle.MarginLeft(snippetListMarginLeft),
				UnselectedTitle:    defaultItemStyle.DimmedTitle.MarginLeft(snippetListMarginLeft),
				CopiedTitleBar:     lipgloss.NewStyle().Background(green).Width(35-2).Margin(0, 1, 1, 2).Padding(0, 1).Foreground(white),
				CopiedTitle:        list.NewDefaultItemStyles().SelectedTitle.Foreground(brightGreen).BorderLeftForeground(brightGreen).MarginLeft(snippetListMarginLeft),
				CopiedSubtitle:     list.NewDefaultItemStyles().SelectedDesc.Foreground(brightGreen).BorderLeftForeground(brightGreen).MarginLeft(snippetListMarginLeft),
			},
			Blurred: SnippetsBaseStyle{
				Base:               lipgloss.NewStyle().Width(sectionBarWidth).MarginTop(config.MarginTop),
				TitleBar:           lipgloss.NewStyle().Background(blue).Width(35-2).Margin(0, 1, 1, 2).Padding(0, 1).Foreground(white),
				SelectedSubtitle:   defaultItemStyle.SelectedDesc.MarginLeft(snippetListMarginLeft),
				UnselectedSubtitle: defaultItemStyle.DimmedDesc.MarginLeft(snippetListMarginLeft),
				SelectedTitle:      defaultItemStyle.SelectedTitle.MarginLeft(snippetListMarginLeft),
				UnselectedTitle:    defaultItemStyle.DimmedTitle.MarginLeft(snippetListMarginLeft),
				CopiedTitleBar:     lipgloss.NewStyle().Background(green).Width(35-2).Margin(0, 1, 1, 2).Padding(0, 1),
				CopiedTitle:        list.NewDefaultItemStyles().DimmedTitle.MarginLeft(snippetListMarginLeft),
				CopiedSubtitle:     list.NewDefaultItemStyles().DimmedDesc.MarginLeft(snippetListMarginLeft),
			},
		},
		Sections: SectionsStyle{
			Focused: SectionsBaseStyle{
				Base:            lipgloss.NewStyle().Width(sectionBarWidth).MarginTop(config.MarginTop),
				TitleBar:        lipgloss.NewStyle().Background(lipgloss.Color("62")).Width(35-2).Margin(0, 1, 1, 1).Padding(0, 1).Foreground(lipgloss.Color("230")),
				SelectedTitle:   sectionSelectedTitle,
				UnselectedTitle: sectionUnselectedTitle,
				CopiedTitleBar:  lipgloss.NewStyle().Background(green).Width(35-2).Margin(0, 1, 1, 1).Padding(0, 1).Foreground(white),
				CopiedTitle:     sectionCopiedTitle,
			},
			Blurred: SectionsBaseStyle{
				Base:            lipgloss.NewStyle().Width(sectionBarWidth).MarginTop(config.MarginTop),
				TitleBar:        lipgloss.NewStyle().Background(blue).Width(35-2).Margin(0, 1, 1, 1).Padding(0, 1).Foreground(white),
				SelectedTitle:   sectionSelectedTitle,
				UnselectedTitle: sectionUnselectedTitle,
				CopiedTitleBar:  lipgloss.NewStyle().Background(green).Width(35-2).Margin(0, 1, 1, 1).Padding(0, 1),
				CopiedTitle:     sectionCopiedTitle,
			},
		},
		Content: ContentStyle{
			Focused: ContentBaseStyle{
				Code:           lipgloss.NewStyle().Margin(1, 0),
				Title:          lipgloss.NewStyle().Background(lipgloss.Color("62")).Width(contentBarWidth).Margin(config.MarginTop, 0, 0, 0).Padding(0, 1).Foreground(lipgloss.Color("230")),
				LineNumber:     lipgloss.NewStyle().Foreground(brightBlack).MarginTop(1),
				EmptyHint:      lipgloss.NewStyle().Foreground(gray),
				EmptyHintKey:   lipgloss.NewStyle().Foreground(brightBlue),
				CopiedTitleBar: lipgloss.NewStyle().Background(green).Width(contentBarWidth).Margin(config.MarginTop, 0, 0, 0).Padding(0, 1).Foreground(white),
			},
			Blurred: ContentBaseStyle{
				Code:           lipgloss.NewStyle().Margin(1, 0),
				Title:          lipgloss.NewStyle().Background(blue).Width(contentBarWidth).Margin(config.MarginTop, 0, 0, 0).Padding(0, 1).Foreground(white),
				LineNumber:     lipgloss.NewStyle().Foreground(brightBlack).MarginTop(1),
				EmptyHint:      lipgloss.NewStyle().Foreground(gray),
				EmptyHintKey:   lipgloss.NewStyle().Foreground(brightBlue),
				CopiedTitleBar: lipgloss.NewStyle().Background(green).Width(contentBarWidth).Margin(config.MarginTop, 0, 0, 0).Padding(0, 1).Foreground(white),
			},
		},
	}
}
