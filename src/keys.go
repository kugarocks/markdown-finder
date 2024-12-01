package main

import "github.com/charmbracelet/bubbles/key"

// KeyMap is the mappings of actions to key bindings.
type KeyMap struct {
	Quit              key.Binding
	Search            key.Binding
	ToggleHelp        key.Binding
	MoveSnippetUp     key.Binding
	MoveSnippetDown   key.Binding
	CopyContent       key.Binding
	CopyContentExit   key.Binding
	EditSnippet       key.Binding
	NextPane          key.Binding
	PrevPane          key.Binding
	ToggleSnippetPane key.Binding
}

// ShortHelp returns a quick help menu.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.NextPane,
		k.Search,
		k.EditSnippet,
		k.ToggleSnippetPane,
		k.ToggleHelp,
	}
}

// FullHelp returns all help options in a more detailed view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.CopyContent, k.EditSnippet},
		{k.MoveSnippetDown, k.MoveSnippetUp},
		{k.NextPane, k.PrevPane},
		{k.Search, k.ToggleSnippetPane},
		{k.ToggleHelp, k.Quit},
	}
}
