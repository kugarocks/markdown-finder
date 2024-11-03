package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	defaultSourceName = "local/source"
)

type Source struct {
	Name string `json:"name"`
	Repo string `json:"repo"`
}

type SourceWrapper struct {
	SourceList []Source `json:"source_list"`
}

func readSources(config Config) ([]Source, error) {
	sourcesDir := config.getSourceBase()
	if err := os.MkdirAll(sourcesDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create sources directory %s: %w", sourcesDir, err)
	}

	configFile := filepath.Join(sourcesDir, config.SourceConfigFile)

	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			wrapper := SourceWrapper{
				SourceList: []Source{},
			}

			b, err := json.MarshalIndent(wrapper, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("unable to serialize default configuration: %w", err)
			}

			err = os.WriteFile(configFile, b, os.ModePerm)
			if err != nil {
				return nil, fmt.Errorf("unable to write configuration file %s: %w", configFile, err)
			}

			return wrapper.SourceList, nil
		}
		return nil, err
	}

	var wrapper SourceWrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("unable to parse configuration file %s: %w", configFile, err)
	}

	if wrapper.SourceList == nil {
		wrapper.SourceList = []Source{}
	}

	return wrapper.SourceList, nil
}

func writeSources(config Config, sources []Source) error {
	wrapper := SourceWrapper{
		SourceList: sources,
	}

	b, err := json.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to serialize configuration: %w", err)
	}
	b = append(b, '\n')

	configFile := filepath.Join(config.getSourceBase(), config.SourceConfigFile)

	err = os.WriteFile(configFile, b, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to write configuration file %s: %w", configFile, err)
	}

	return nil
}
