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
	configFile := filepath.Join(config.Home, config.SourceConfigFile)

	if err := os.MkdirAll(config.Home, os.ModePerm); err != nil {
		return nil, fmt.Errorf("无法创建目录 %s: %w", config.Home, err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			wrapper := SourceWrapper{
				SourceList: []Source{},
			}

			b, err := json.MarshalIndent(wrapper, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("无法序列化默认配置: %w", err)
			}

			err = os.WriteFile(configFile, b, os.ModePerm)
			if err != nil {
				return nil, fmt.Errorf("无法写入配置文件 %s: %w", configFile, err)
			}

			return wrapper.SourceList, nil
		}
		return nil, err
	}

	var wrapper SourceWrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("无法解析配置文件 %s: %w", configFile, err)
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
		return fmt.Errorf("无法序列化配置: %w", err)
	}

	configFile := filepath.Join(config.Home, config.SourceConfigFile)

	err = os.WriteFile(configFile, b, os.ModePerm)
	if err != nil {
		return fmt.Errorf("无法写入配置文件 %s: %w", configFile, err)
	}

	return nil
}
