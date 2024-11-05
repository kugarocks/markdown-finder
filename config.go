package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v3"
)

var (
	TitlePadding        = []int{0, 1}
	SnippetBarMargin    = []int{0, 1, 1, 2}
	SectionBarMargin    = []int{0, 1, 1, 1}
	ContentCodeMargin   = []int{1, 0}
	CodeBlockMarginZero = uint(0)

	HelpText = `
Nap is a code snippet manager for your terminal.
https://github.com/maaslalani/nap

Usage:
  nap           - for interactive mode
  nap list      - list all snippets
  nap <snippet> - print snippet to stdout

Create:
  nap < main.go                 - save snippet from stdin
  nap example/main.go < main.go - save snippet with name

`
	DefaultSnippetConfig = `{
	"snippet_list": []
}`

	DefaultSnippetContent = `## Quick Start

* tab - switch pane
* j/k - cursor down/up
* c/d - copy code block
* e - edit snippet
* use "---" to separate sections
* each section needs a title

` + "```bash" + `
echo "hello world"
` + "```" + `

` + "```bash" + `
echo "Bananaaaaa 🍌"
` + "```" + `

---

## Charm.sh

We make the command line glamorous.

` + "```bash" + `
echo "Charm Rocks 🚀"
` + "```" + `
`
)

// Config holds the configuration options for the application.
//
// At the moment, it is quite limited, only supporting the home folder and the
// file name of the metadata.
type Config struct {
	Home              string `env:"MDF_HOME" yaml:"home"`
	SourceName        string `env:"MDF_SOURCE_NAME" yaml:"source_name"`
	FolderName        string `env:"MDF_FOLDER_NAME" yaml:"folder_name"`
	SourceConfigFile  string `env:"MDF_SOURCE_CONFIG_FILE" yaml:"source_config_file"`
	SnippetConfigFile string `env:"MDF_SNIPPET_CONFIG_FILE" yaml:"snippet_config_file"`

	// Layout
	BaseMarginTop         int  `env:"MDF_BASE_MARGIN_TOP" yaml:"base_margin_top"`
	SnippetTitleBarWidth  int  `env:"MDF_SNIPPET_TITLE_BAR_WIDTH" yaml:"snippet_title_bar_width"`
	SectionTitleBarWidth  int  `env:"MDF_SECTION_TITLE_BAR_WIDTH" yaml:"section_title_bar_width"`
	ContentTitleBarWidth  int  `env:"MDF_CONTENT_TITLE_BAR_WIDTH" yaml:"content_title_bar_width"`
	SnippetListMarginLeft int  `env:"MDF_SNIPPET_LIST_MARGIN_LEFT" yaml:"snippet_list_margin_left"`
	AlwaysShowSnippetPane bool `env:"MDF_ALWAYS_SHOW_SNIPPET_PANE" yaml:"always_show_snippet_pane"`

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
	CodeBlockTheme     string `env:"MDF_THEME" yaml:"theme"`
	CodeBlockPrefix    string `env:"MDF_CODE_BLOCK_PREFIX" yaml:"code_block_prefix"`
	CodeBlockSuffix    string `env:"MDF_CODE_BLOCK_SUFFIX" yaml:"code_block_suffix"`
	CodeBlockCopedHint string `env:"MDF_CODE_BLOCK_COPIED_HINT" yaml:"code_block_copied_hint"`
}

func newConfig() Config {
	return Config{
		Home:              defaultHome(),
		SourceName:        defaultSourceName,
		SourceConfigFile:  "source-config.json",
		SnippetConfigFile: "snippet-config.json",

		// Layout
		BaseMarginTop:         1,
		SnippetTitleBarWidth:  33,
		SectionTitleBarWidth:  33,
		ContentTitleBarWidth:  86,
		SnippetListMarginLeft: 1,
		AlwaysShowSnippetPane: false,

		// Colors
		FocusedBarBgColor:        "62",
		FocusedBarFgColor:        "230",
		BlurredBarBgColor:        "#64708D",
		BlurredBarFgColor:        "#FFFFFF",
		SelectedItemFgColor:      "170",
		UnselectedItemFgColor:    "c7c7c7",
		CopiedBarBgColor:         "#527251",
		CopiedBarFgColor:         "#FFFFFF",
		CopiedItemFgColor:        "#BCE1AF",
		ContentLineNumberFgColor: "241",

		// Code Block
		CodeBlockTheme:     "dracula",
		CodeBlockPrefix:    "------------- CodeBlock -------------",
		CodeBlockSuffix:    "---------------- End ----------------",
		CodeBlockCopedHint: "---------- Press %s to copy ----------",
	}
}

// default helpers for the configuration.
// We use $XDG_DATA_HOME to avoid cluttering the user's home directory.
// For macOS: ~/Library/Application Support/mdf
func defaultHome() string { return filepath.Join(xdg.DataHome, "mdf") }

// getConfigFilePath returns the config file path
func getConfigFilePath() string {
	if c := os.Getenv("MDF_CONFIG"); c != "" {
		return c
	}
	cfgPath, err := xdg.ConfigFile("mdf/config.yaml")
	if err != nil {
		return "config.yaml"
	}
	return cfgPath
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
		if err = yaml.NewDecoder(fi).Decode(&config); err != nil {
			return newConfig()
		}
	}

	if err = env.Parse(&config); err != nil {
		return newConfig()
	}

	if strings.HasPrefix(config.Home, "~") {
		var home string
		home, err = os.UserHomeDir()
		if err == nil {
			config.Home = filepath.Join(home, config.Home[1:])
		}
	}

	return config
}

// writeConfig returns a configuration read from the environment.
func (config Config) writeConfig() error {
	fi, err := os.Create(getConfigFilePath())
	if err != nil {
		return err
	}
	if fi != nil {
		defer fi.Close()
		if err = yaml.NewEncoder(fi).Encode(&config); err != nil {
			return err
		}
	}

	return nil
}

// getSourceBase returns the base path for the configured source name
func (config Config) getSourceBase() string {
	return filepath.Join(config.Home, "sources")
}

// getSourcePath returns the full path for the configured source name
func (config Config) getSourcePath() string {
	parts := strings.Split(config.SourceName, "/")
	return filepath.Join(append([]string{config.getSourceBase()}, parts...)...)
}

// getDefaultSourcePath returns the full path for the default source name
func (config Config) getDefaultSourcePath() string {
	parts := strings.Split(defaultSourceName, "/")
	return filepath.Join(append([]string{config.getSourceBase()}, parts...)...)
}
