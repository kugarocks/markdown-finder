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

// Config holds the configuration options for the application.
//
// At the moment, it is quite limited, only supporting the home folder and the
// file name of the metadata.
type Config struct {
	Home              string `env:"MDF_HOME" yaml:"home"`
	SourceConfigFile  string `env:"MDF_SOURCE_CONFIG_FILE" yaml:"source_config_file"`
	SnippetConfigFile string `env:"MDF_SNIPPET_CONFIG_FILE" yaml:"snippet_config_file"`

	Theme               string `env:"MDF_THEME" yaml:"theme"`
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
	MarginTop           int    `env:"MDF_MARGIN_TOP" yaml:"margin_top"`
}

func newConfig() Config {
	return Config{
		Home:                defaultHome(),
		SourceConfigFile:    "source-config.json",
		SnippetConfigFile:   "snippet-config.json",
		Theme:               "dracula",
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
		MarginTop:           1,
	}
}

// default helpers for the configuration.
// We use $XDG_DATA_HOME to avoid cluttering the user's home directory.
// For macOS: ~/Library/Application Support/mdf
func defaultHome() string { return filepath.Join(xdg.DataHome, "mdf") }

// defaultConfig returns the default config path
func defaultConfig() string {
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
	fi, err := os.Open(defaultConfig())
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return newConfig()
	}
	if fi != nil {
		defer fi.Close()
		if err := yaml.NewDecoder(fi).Decode(&config); err != nil {
			return newConfig()
		}
	}

	if err := env.Parse(&config); err != nil {
		return newConfig()
	}

	if strings.HasPrefix(config.Home, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			config.Home = filepath.Join(home, config.Home[1:])
		}
	}

	return config
}

// writeConfig returns a configuration read from the environment.
func (config Config) writeConfig() error {
	fi, err := os.Create(defaultConfig())
	if err != nil {
		return err
	}
	if fi != nil {
		defer fi.Close()
		if err := yaml.NewEncoder(fi).Encode(&config); err != nil {
			return err
		}
	}

	return nil
}
