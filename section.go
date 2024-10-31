package main

import (
	"fmt"
	"path/filepath"
)

// default values for empty state.
const (
	defaultSectionFolder  = "default"
	defaultSectionFile    = "example.md"
	defaultSectionTitle   = "No Title"
	defaultSectionContent = ""
)

// Section represents a partial content of section in markdown file.
type Section struct {
	Folder     string   `json:"folder"`
	File       string   `json:"file"`
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	CodeBlocks []string `json:"code_blocks"`
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

// String returns the string of the section at the specified position i
func (s Sections) String(i int) string {
	return s.sections[i].String()
}

// Len returns the length of the sections array.
func (s Sections) Len() int {
	return len(s.sections)
}
