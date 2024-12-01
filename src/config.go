package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/caarlos0/env/v6"
	"github.com/charmbracelet/bubbles/key"
	"gopkg.in/yaml.v3"
)

var (
	TitlePadding        = []int{0, 1}
	SnippetBarMargin    = []int{0, 1, 1, 2}
	SectionBarMargin    = []int{0, 1, 1, 1}
	ContentCodeMargin   = []int{1, 0}
	CodeBlockMarginZero = uint(0)

	CodeBlockBorderLength = 39

	HelpText = `
mdf is a markdown finder in your terminal.
https://github.com/kugarocks/markdown-finder

Usage:
  mdf                   - for interactive mode (3 panes)
  mdf example           - fuzzy find snippet (2 panes)
  mdf get repo <name>   - get repo from github
  mdf set repo          - switch repo
  mdf set folder        - switch folder
  mdf list repo         - list all repos
  mdf list folder       - list all folders
  mdf list snippet      - list all snippets

`
	DefaultSnippetConfig = `{
	"snippet_list": []
}`

	DefaultSnippetContent = `## Quick Start

* n/N - next/prev pane
* j/k - cursor down/up
* c/d - copy code block
* i - edit snippet
* s - toggle snippet pane
* use "---" to separate sections
* each section needs a title

` + "```bash {copyable}" + `
echo "Charm.sh Rocks 🚀"
` + "```" + `

` + "```bash {title=\"Custom Title\"}" + `
echo "https://minions.wiki"
` + "```" + `

---

## GitHub Repository

Get repo from GitHub by SSH:

` + "```bash {copyable}" + `
mdf get repo kugarocks/rockman
` + "```" + `

HTTPS URL is also supported:

` + "```bash {copyable}" + `
mdf get repo https://github.com/kugarocks/rockman.git
` + "```" + `

Switch repo:

` + "```bash {copyable}" + `
mdf set repo
` + "```" + `

---

## More Commands

Switch folder:

` + "```bash {copyable}" + `
mdf set folder
` + "```" + `

Fuzzy find snippet:

` + "```bash {copyable}" + `
mdf examp
` + "```" + `

List folders:

` + "```bash {copyable}" + `
mdf list folder
` + "```" + `

---

## Configuration

Checkout:

` + "```bash {copyable}" + `
https://github.com/kugarocks/markdown-finder
` + "```" + `

---

## Raycast Script Command

` + "```bash {copyable}" + `
LANG=en_US.UTF-8 \
MDF_HOME=/Users/kuga/mdf \
/Applications/Alacritty.app/Contents/MacOS/alacritty \
    --config-file /Users/kuga/alacritty.toml \
    -e /usr/local/bin/mdf "$1" \
    > /dev/null 2>&1
` + "```" + `
`
)

// Config holds the configuration options for the application.
//
// At the moment, it is quite limited, only supporting the home folder and the
// file name of the metadata.
type Config struct {
	Home              string `yaml:"-"`
	RepoName          string `env:"MDF_REPO_NAME" yaml:"repo_name"`
	FolderName        string `env:"MDF_FOLDER_NAME" yaml:"folder_name"`
	RepoConfigFile    string `env:"MDF_REPO_CONFIG_FILE" yaml:"repo_config_file"`
	SnippetConfigFile string `env:"MDF_SNIPPET_CONFIG_FILE" yaml:"snippet_config_file"`

	// Pane
	DefaultPane           string `env:"MDF_DEFAULT_PANE" yaml:"default_pane"`
	AlwaysShowSnippetPane bool   `env:"MDF_ALWAYS_SHOW_SNIPPET_PANE" yaml:"always_show_snippet_pane"`
	ExitAfterCopy         bool   `env:"MDF_EXIT_AFTER_COPY" yaml:"exit_after_copy"`

	// Layout
	BaseMarginTop         int `env:"MDF_BASE_MARGIN_TOP" yaml:"base_margin_top"`
	SnippetTitleBarWidth  int `env:"MDF_SNIPPET_TITLE_BAR_WIDTH" yaml:"snippet_title_bar_width"`
	SectionTitleBarWidth  int `env:"MDF_SECTION_TITLE_BAR_WIDTH" yaml:"section_title_bar_width"`
	ContentTitleBarWidth  int `env:"MDF_CONTENT_TITLE_BAR_WIDTH" yaml:"content_title_bar_width"`
	SnippetListMarginLeft int `env:"MDF_SNIPPET_LIST_MARGIN_LEFT" yaml:"snippet_list_margin_left"`

	// Colors
	FocusedBarBgColor        string `env:"MDF_FOCUSED_BAR_BG_COLOR" yaml:"focused_bar_bg_color"`
	FocusedBarFgColor        string `env:"MDF_FOCUSED_BAR_FG_COLOR" yaml:"focused_bar_fg_color"`
	BlurredBarBgColor        string `env:"MDF_BLURRED_BAR_BG_COLOR" yaml:"blurred_bar_bg_color"`
	BlurredBarFgColor        string `env:"MDF_BLURRED_BAR_FG_COLOR" yaml:"blurred_bar_fg_color"`
	SelectedItemFgColor      string `env:"MDF_SELECTED_ITEM_FG_COLOR" yaml:"selected_item_fg_color"`
	UnselectedItemFgColor    string `env:"MDF_UNSELECTED_ITEM_FG_COLOR" yaml:"unselected_item_fg_color"`
	CopiedBarBgColor         string `env:"MDF_COPIED_BAR_BG_COLOR" yaml:"copied_bar_bg_color"`
	CopiedBarFgColor         string `env:"MDF_COPIED_BAR_FG_COLOR" yaml:"copied_bar_fg_color"`
	CopiedItemFgColor        string `env:"MDF_COPIED_ITEM_FG_COLOR" yaml:"copied_item_fg_color"`
	ContentLineNumberFgColor string `env:"MDF_CONTENT_LINE_NUMBER_FG_COLOR" yaml:"content_line_number_fg_color"`

	// Code Block
	CodeBlockTheme         string `env:"MDF_THEME" yaml:"theme"`
	CodeBlockBorderPadding string `env:"MDF_CODE_BLOCK_BORDER_PADDING" yaml:"code_block_border_padding"`
	CodeBlockBorderLength  int    `env:"MDF_CODE_BLOCK_BORDER_LENGTH" yaml:"code_block_border_length"`
	CodeBlockTitleCopy     string `env:"MDF_CODE_BLOCK_TITLE_COPY" yaml:"code_block_title_copy"`
	CodeBlockPrefixTemp    string `yaml:"-"`
	CodeBlockSuffixTemp    string `yaml:"-"`
	CodeBlockBorderDefault string `yaml:"-"`

	// keys
	CopyContentKeys        []string `env:"MDF_COPY_CONTENT_KEYS" envSeparator:"," yaml:"copy_content_keys"`
	CopyContentKeysCapital []string `yaml:"-"`
	EditSnippetKeys        []string `env:"MDF_EDIT_SNIPPET_KEYS" envSeparator:"," yaml:"edit_snippet_keys"`
	NextPaneKeys           []string `env:"MDF_NEXT_PANE_KEYS" envSeparator:"," yaml:"next_pane_keys"`
	PrevPaneKeys           []string `env:"MDF_PREV_PANE_KEYS" envSeparator:"," yaml:"prev_pane_keys"`
	ToggleSnippetPaneKeys  []string `env:"MDF_TOGGLE_SNIPPET_PANE_KEYS" envSeparator:"," yaml:"toggle_snippet_pane_keys"`
}

func newConfig() Config {
	return Config{
		Home:              defaultHome(),
		RepoName:          defaultRepoName,
		RepoConfigFile:    "repo-config.json",
		SnippetConfigFile: "snippet-config.json",

		// Pane
		DefaultPane:           "section",
		AlwaysShowSnippetPane: false,
		ExitAfterCopy:         false,

		// Layout
		BaseMarginTop:         1,
		SnippetTitleBarWidth:  33,
		SectionTitleBarWidth:  33,
		ContentTitleBarWidth:  86,
		SnippetListMarginLeft: 1,

		// Colors
		// https://commons.wikimedia.org/wiki/File:Xterm_256color_chart.svg
		// Support RGB color as well: #5f5fd7
		FocusedBarBgColor:        "62",
		FocusedBarFgColor:        "255",
		BlurredBarBgColor:        "103",
		BlurredBarFgColor:        "255",
		SelectedItemFgColor:      "170",
		UnselectedItemFgColor:    "252",
		CopiedBarBgColor:         "42",
		CopiedBarFgColor:         "238",
		CopiedItemFgColor:        "42",
		ContentLineNumberFgColor: "241",

		// Code Block
		CodeBlockTheme:         "dracula",
		CodeBlockBorderPadding: "-",
		CodeBlockBorderLength:  CodeBlockBorderLength,
		CodeBlockTitleCopy:     "Press {key} to copy",
		CodeBlockPrefixTemp:    "------------------BEG------------------",
		CodeBlockSuffixTemp:    "------------------END------------------",

		// keys
		CopyContentKeys:       []string{"c", "d", "e", "f", "g"},
		EditSnippetKeys:       []string{"i"},
		NextPaneKeys:          []string{"n", "tab", "right"},
		PrevPaneKeys:          []string{"N", "shift+tab", "left"},
		ToggleSnippetPaneKeys: []string{"s", "p"},
	}
}

// defaultHome returns the default home directory for the application.
func defaultHome() string {
	// check environment variable first
	if envHome := strings.TrimSpace(os.Getenv("MDF_HOME")); envHome != "" {
		// if the environment variable starts with ~
		if strings.HasPrefix(envHome, "~") {
			if home, err := os.UserHomeDir(); err == nil {
				return filepath.Join(home, envHome[1:])
			}
			// fallback to xdg
			return filepath.Join(xdg.DataHome, "mdf")
		}
		return envHome
	}

	// try user home directory
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".mdf")
	}

	// fallback to xdg
	return filepath.Join(xdg.DataHome, "mdf")
}

// getConfigFilePath returns the config file path
func getConfigFilePath() string {
	return filepath.Join(defaultHome(), "config.yaml")
}

// readConfig returns a configuration read from the environment.
func readConfig() Config {
	config := newConfig()
	configFilePath := getConfigFilePath()

	fi, err := os.Open(configFilePath)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		_ = config.writeConfig()
	}
	if fi != nil {
		defer fi.Close()
		err = yaml.NewDecoder(fi).Decode(&config)
		if err != nil {
			config = newConfig()
		}
	}

	if err = env.Parse(&config); err != nil {
		config = newConfig()
	}

	// set code block default config
	if config.CodeBlockBorderLength <= 0 {
		config.CodeBlockBorderLength = CodeBlockBorderLength
	}
	config.CodeBlockBorderPadding = config.CodeBlockBorderPadding[:1]
	config.CodeBlockBorderDefault = strings.Repeat(config.CodeBlockBorderPadding, config.CodeBlockBorderLength)

	// Capital CopyContentKeys
	config.CopyContentKeysCapital = make([]string, len(config.CopyContentKeys))
	for i, key := range config.CopyContentKeys {
		config.CopyContentKeysCapital[i] = strings.ToUpper(key)
	}

	return config
}

// writeConfig writes the configuration to a YAML file.
func (config Config) writeConfig() error {
	// Open file for writing
	fi, err := os.Create(getConfigFilePath())
	if err != nil {
		return err
	}
	defer fi.Close()

	// Create encoder with indentation
	enc := yaml.NewEncoder(fi)
	enc.SetIndent(2)
	defer enc.Close()

	// Encode config to YAML node
	var node yaml.Node
	if err := node.Encode(config); err != nil {
		return err
	}

	// Set flow style for array fields
	setFlowStyle(&node, map[string]struct{}{
		"copy_content_keys":        {},
		"edit_snippet_keys":        {},
		"next_pane_keys":           {},
		"prev_pane_keys":           {},
		"toggle_snippet_pane_keys": {},
	})

	return enc.Encode(&node)
}

// setFlowStyle recursively traverses the YAML node tree and sets flow style
// for specified fields that are sequences (arrays).
func setFlowStyle(node *yaml.Node, fields map[string]struct{}) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.DocumentNode:
		// Process document content
		for _, n := range node.Content {
			setFlowStyle(n, fields)
		}
	case yaml.MappingNode:
		// Process key-value pairs
		for i := 0; i < len(node.Content); i += 2 {
			k, value := node.Content[i], node.Content[i+1]
			// Set flow style if field matches and is a sequence
			if _, ok := fields[k.Value]; ok && value.Kind == yaml.SequenceNode {
				value.Style = yaml.FlowStyle
			}
			setFlowStyle(value, fields)
		}
	}
}

// getRepoBase returns the base path for the configured repo name
func (config Config) getRepoBase() string {
	return filepath.Join(config.Home, "repos")
}

// getRepoPath returns the full path for the configured repo name
func (config Config) getRepoPath() string {
	parts := strings.Split(config.RepoName, "/")
	return filepath.Join(append([]string{config.getRepoBase()}, parts...)...)
}

// getDefaultRepoPath returns the full path for the default repo name
func (config Config) getDefaultRepoPath() string {
	parts := strings.Split(defaultRepoName, "/")
	return filepath.Join(append([]string{config.getRepoBase()}, parts...)...)
}

func (config Config) newKeyMap() KeyMap {
	var km = KeyMap{
		Quit:            key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "exit")),
		Search:          key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		ToggleHelp:      key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		MoveSnippetDown: key.NewBinding(key.WithKeys("J"), key.WithHelp("J", "move snippet down")),
		MoveSnippetUp:   key.NewBinding(key.WithKeys("K"), key.WithHelp("K", "move snippet up")),
	}

	setKeyBinding := func(binding *key.Binding, keys []string, helpText string) {
		if len(keys) > 0 {
			binding.SetKeys(keys...)
			if len(keys) > 2 {
				keys = keys[:2]
			}
			binding.SetHelp(strings.Join(keys, "/"), helpText)
		}
	}

	setKeyBinding(&km.CopyContent, config.CopyContentKeys, "copy")
	setKeyBinding(&km.CopyContentExit, config.CopyContentKeysCapital, "copy & exit")
	setKeyBinding(&km.EditSnippet, config.EditSnippetKeys, "edit")
	setKeyBinding(&km.NextPane, config.NextPaneKeys, "next")
	setKeyBinding(&km.PrevPane, config.PrevPaneKeys, "prev")
	setKeyBinding(&km.ToggleSnippetPane, config.ToggleSnippetPaneKeys, "toggle snippet")

	return km
}
