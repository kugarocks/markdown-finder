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
	DeletedTitleBar    lipgloss.Style
	DeletedTitle       lipgloss.Style
	DeletedSubtitle    lipgloss.Style
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
	CopiedSubtitle     lipgloss.Style
}

// ContentBaseStyle holds the necessary styling for the content pane of the
// application.
type ContentBaseStyle struct {
	Code         lipgloss.Style
	Title        lipgloss.Style
	Separator    lipgloss.Style
	LineNumber   lipgloss.Style
	EmptyHint    lipgloss.Style
	EmptyHintKey lipgloss.Style
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
	red := lipgloss.Color(config.RedColor)
	brightRed := lipgloss.Color(config.BrightRedColor)
	
	defaultItemStyle := list.NewDefaultItemStyles()
	sectionBarWidth := 36
	contentBarWidth := 86
	
	return Styles{
		Snippets: SnippetsStyle{
			Focused: SnippetsBaseStyle{
				Base: lipgloss.NewStyle().Width(sectionBarWidth).MarginTop(config.MarginTop),
				//TitleBar:           lipgloss.NewStyle().Background(blue).Width(35-2).Margin(0, 1, 1, 4).Padding(0, 1).Foreground(white),
				TitleBar:           lipgloss.NewStyle().Background(lipgloss.Color("62")).Width(35-2).Margin(0, 1, 1, 2).Padding(0, 1).Foreground(lipgloss.Color("230")),
				SelectedSubtitle:   defaultItemStyle.SelectedDesc,
				UnselectedSubtitle: defaultItemStyle.DimmedDesc,
				SelectedTitle:      defaultItemStyle.SelectedTitle,
				UnselectedTitle:    defaultItemStyle.DimmedTitle,
				CopiedTitleBar:     lipgloss.NewStyle().Background(green).Width(35-2).Margin(0, 1, 1, 2).Padding(0, 1).Foreground(white),
				CopiedTitle:        list.NewDefaultItemStyles().SelectedTitle.Foreground(brightGreen).BorderLeftForeground(brightGreen),
				CopiedSubtitle:     list.NewDefaultItemStyles().SelectedDesc.Foreground(brightGreen).BorderLeftForeground(brightGreen),
				DeletedTitleBar:    lipgloss.NewStyle().Background(red).Width(35-2).Margin(0, 1, 1, 1).Padding(0, 1).Foreground(white),
				DeletedTitle:       lipgloss.NewStyle().Foreground(brightRed),
				DeletedSubtitle:    lipgloss.NewStyle().Foreground(red),
			},
			Blurred: SnippetsBaseStyle{
				Base: lipgloss.NewStyle().Width(sectionBarWidth).MarginTop(config.MarginTop),
				//TitleBar:           lipgloss.NewStyle().Background(black).Width(35-2).Margin(0, 1, 1, 4).Padding(0, 1).Foreground(gray),
				TitleBar:           lipgloss.NewStyle().Background(blue).Width(35-2).Margin(0, 1, 1, 2).Padding(0, 1).Foreground(white),
				SelectedSubtitle:   defaultItemStyle.SelectedDesc,
				UnselectedSubtitle: defaultItemStyle.DimmedDesc,
				SelectedTitle:      defaultItemStyle.SelectedTitle,
				UnselectedTitle:    defaultItemStyle.DimmedTitle,
				CopiedTitleBar:     lipgloss.NewStyle().Background(green).Width(35-2).Margin(0, 1, 1, 2).Padding(0, 1),
				CopiedTitle:        list.NewDefaultItemStyles().DimmedTitle,
				CopiedSubtitle:     list.NewDefaultItemStyles().DimmedDesc,
				DeletedTitleBar:    lipgloss.NewStyle().Background(red).Width(35-2).Margin(0, 1, 1, 1).Padding(0, 1),
				DeletedTitle:       lipgloss.NewStyle().Foreground(brightRed),
				DeletedSubtitle:    lipgloss.NewStyle().Foreground(red),
			},
		},
		Sections: SectionsStyle{
			Focused: SectionsBaseStyle{
				Base: lipgloss.NewStyle().Width(sectionBarWidth).MarginTop(config.MarginTop),
				//TitleBar:           lipgloss.NewStyle().Background(blue).Width(35-2).Margin(0, 1, 1, 4).Padding(0, 1).Foreground(white),
				TitleBar:           lipgloss.NewStyle().Background(lipgloss.Color("62")).Width(35-2).Margin(0, 1, 1, 2).Padding(0, 1).Foreground(lipgloss.Color("230")),
				SelectedSubtitle:   defaultItemStyle.SelectedDesc,
				UnselectedSubtitle: defaultItemStyle.DimmedDesc,
				SelectedTitle:      defaultItemStyle.SelectedTitle,
				UnselectedTitle:    defaultItemStyle.DimmedTitle,
				CopiedTitleBar:     lipgloss.NewStyle().Background(green).Width(35-2).Margin(0, 1, 1, 2).Padding(0, 1).Foreground(white),
				CopiedTitle:        list.NewDefaultItemStyles().SelectedTitle.Foreground(brightGreen).BorderLeftForeground(brightGreen),
				CopiedSubtitle:     list.NewDefaultItemStyles().SelectedDesc.Foreground(brightGreen).BorderLeftForeground(brightGreen),
			},
			Blurred: SectionsBaseStyle{
				Base: lipgloss.NewStyle().Width(sectionBarWidth).MarginTop(config.MarginTop),
				//TitleBar:           lipgloss.NewStyle().Background(black).Width(35-2).Margin(0, 1, 1, 4).Padding(0, 1).Foreground(gray),
				TitleBar:           lipgloss.NewStyle().Background(blue).Width(35-2).Margin(0, 1, 1, 2).Padding(0, 1).Foreground(white),
				SelectedSubtitle:   defaultItemStyle.SelectedDesc,
				UnselectedSubtitle: defaultItemStyle.DimmedDesc,
				SelectedTitle:      defaultItemStyle.SelectedTitle,
				UnselectedTitle:    defaultItemStyle.DimmedTitle,
				CopiedTitleBar:     lipgloss.NewStyle().Background(green).Width(35-2).Margin(0, 1, 1, 2).Padding(0, 1),
				CopiedTitle:        list.NewDefaultItemStyles().DimmedTitle,
				CopiedSubtitle:     list.NewDefaultItemStyles().DimmedDesc,
			},
		},
		Content: ContentStyle{
			Focused: ContentBaseStyle{
				Code:  lipgloss.NewStyle().Margin(1, 1),
				Title: lipgloss.NewStyle().Background(lipgloss.Color("62")).Width(contentBarWidth).Margin(config.MarginTop, 0, 0, 1).Padding(0, 1).Foreground(lipgloss.Color("230")),
				//Title:        lipgloss.NewStyle().Background(blue).Width(35-2).Foreground(white).Margin(0, 0, 0, 1).Padding(0, 1),
				//Separator:    lipgloss.NewStyle().Foreground(white).Margin(0, 0, 0, 1),
				LineNumber:   lipgloss.NewStyle().Foreground(brightBlack).MarginTop(1),
				EmptyHint:    lipgloss.NewStyle().Foreground(gray),
				EmptyHintKey: lipgloss.NewStyle().Foreground(brightBlue),
			},
			Blurred: ContentBaseStyle{
				Code:  lipgloss.NewStyle().Margin(1, 1),
				Title: lipgloss.NewStyle().Background(blue).Width(contentBarWidth).Margin(config.MarginTop, 0, 0, 1).Padding(0, 1).Foreground(white),
				//Title:        lipgloss.NewStyle().Background(black).Width(35-2).Foreground(gray).Margin(0, 0, 0, 1).Padding(0, 1),
				//Separator:    lipgloss.NewStyle().Foreground(gray).Margin(0, 0, 0, 1),
				LineNumber:   lipgloss.NewStyle().Foreground(brightBlack).MarginTop(1),
				EmptyHint:    lipgloss.NewStyle().Foreground(gray),
				EmptyHintKey: lipgloss.NewStyle().Foreground(brightBlue),
			},
		},
	}
}
