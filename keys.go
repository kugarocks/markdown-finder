package main

import "github.com/charmbracelet/bubbles/key"

// KeyMap is the mappings of actions to key bindings.
type KeyMap struct {
	Quit            key.Binding
	Search          key.Binding
	ToggleHelp      key.Binding
	MoveSnippetUp   key.Binding
	MoveSnippetDown key.Binding
	CopyContent     key.Binding
	EditSnippet     key.Binding
	NextPane        key.Binding
	PrevPane        key.Binding
}

// DefaultKeyMap is the default key map for the application.
var DefaultKeyMap = KeyMap{
	Quit:            key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "exit")),
	Search:          key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
	ToggleHelp:      key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	MoveSnippetDown: key.NewBinding(key.WithKeys("J"), key.WithHelp("J", "move snippet down")),
	MoveSnippetUp:   key.NewBinding(key.WithKeys("K"), key.WithHelp("K", "move snippet up")),
	CopyContent:     key.NewBinding(key.WithKeys("c", "d", "f"), key.WithHelp("c", "copy")),
	EditSnippet:     key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
	NextPane:        key.NewBinding(key.WithKeys("tab", "right"), key.WithHelp("tab", "navigate")),
	PrevPane:        key.NewBinding(key.WithKeys("shift+tab", "left"), key.WithHelp("shift+tab", "navigate")),
}

// ShortHelp returns a quick help menu.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.NextPane,
		k.Search,
		k.EditSnippet,
		k.ToggleHelp,
	}
}

// FullHelp returns all help options in a more detailed view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.CopyContent, k.EditSnippet},
		{k.MoveSnippetDown, k.MoveSnippetUp},
		{k.NextPane, k.PrevPane},
		{k.Search, k.ToggleHelp},
		{k.Quit},
	}
}
