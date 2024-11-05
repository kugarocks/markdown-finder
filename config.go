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
	BaseMarginTop         = 1
	SnippetTitleBarWidth  = 33
	SectionTitleBarWidth  = 33
	ContentTitleBarWidth  = 86
	SnippetListMarginLeft = 1

	FocusedBarBgColor        = "62"
	FocusedBarFgColor        = "230"
	BlurredBarBgColor        = "#64708D"
	BlurredBarFgColor        = "#FFFFFF"
	SelectedItemFgColor      = "170"
	UnselectedItemFgColor    = "c7c7c7"
	CopiedBarBgColor         = "#527251"
	CopiedBarFgColor         = "#FFFFFF"
	CopiedItemFgColor        = "#BCE1AF"
	ContentLineNumberFgColor = "241"

	CodeBlockPrefix     = "------------- CodeBlock -------------"
	CodeBlockSuffix     = "---------------- End ----------------"
	CodeBlockCopyPrefix = "---------- Press %s to copy ----------"
	CodeBlockMarginZero = uint(0)

	TitlePadding      = []int{0, 1}
	SnippetBarMargin  = []int{0, 1, 1, 2}
	SectionBarMargin  = []int{0, 1, 1, 1}
	ContentCodeMargin = []int{1, 0}

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

	CodeBlockTheme      string `env:"MDF_THEME" yaml:"theme"`
	PrimaryColor        string `env:"MDF_PRIMARY_COLOR" yaml:"primary_color"`
	PrimaryColorSubdued string `env:"MDF_PRIMARY_COLOR_SUBDUED" yaml:"primary_color_subdued"`
	BrightGreenColor    string `env:"MDF_BRIGHT_GREEN" yaml:"bright_green"`
	GreenColor          string `env:"MDF_GREEN" yaml:"green"`
	BrightRedColor      string `env:"MDF_BRIGHT_RED" yaml:"bright_red"`
	RedColor            string `env:"MDF_RED" yaml:"red"`
	ForegroundColor     string `env:"MDF_FOREGROUND" yaml:"foreground"`
	BackgroundColor     string `env:"MDF_BACKGROUND" yaml:"background"`
	GrayColor           string `env:"MDF_GRAY" yaml:"gray"`
	BlackColor          string `env:"MDF_BLACK" yaml:"black"`
	WhiteColor          string `env:"MDF_WHITE" yaml:"white"`
	BaseMarginTop       int    `env:"MDF_MARGIN_TOP" yaml:"margin_top"`
}

func newConfig() Config {
	return Config{
		Home:                defaultHome(),
		SourceName:          defaultSourceName,
		SourceConfigFile:    "source-config.json",
		SnippetConfigFile:   "snippet-config.json",
		CodeBlockTheme:      "dracula",
		PrimaryColor:        "#AFBEE1",
		PrimaryColorSubdued: "#64708D",
		BrightGreenColor:    "#BCE1AF",
		GreenColor:          "#527251",
		BrightRedColor:      "#E49393",
		RedColor:            "#A46060",
		ForegroundColor:     "15",
		BackgroundColor:     "235",
		GrayColor:           "241",
		BlackColor:          "#373b41",
		WhiteColor:          "#FFFFFF",
		BaseMarginTop:       BaseMarginTop,
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
