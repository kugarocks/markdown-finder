package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	defaultSourceName = "local/source"
)

var (
	sourceTitleStyle        = lipgloss.NewStyle().MarginLeft(2)
	sourceItemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	sourceSelectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	sourcePaginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	sourceHelpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	sourceQuitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
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

func setSource(config *Config) error {
	sources, err := readSources(*config)
	if err != nil {
		return fmt.Errorf("读取源配置失败: %w", err)
	}

	// 转换源列表为 list.Item，同时找到当前源的索引
	var items []list.Item
	currentIndex := 0
	for i, source := range sources {
		items = append(items, sourceItem(source))
		if source.Name == config.SourceName {
			currentIndex = i
		}
	}

	// 创建列表，设置默认选中项
	l := list.New(items, sourceDelegate{}, 30, 14)
	l.Title = "Choose a source"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = sourceTitleStyle
	l.Styles.PaginationStyle = sourcePaginationStyle
	l.Styles.HelpStyle = sourceHelpStyle

	// 设置当前选中项
	l.Select(currentIndex)

	// 运行交互程序
	p := tea.NewProgram(sourceModel{
		list:    l,
		config:  config,
		quiting: false,
	})

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("运行交互界面失败: %w", err)
	}

	return nil
}

// 源列表项
type sourceItem Source

func (i sourceItem) FilterValue() string { return i.Name }

// 源列表代理
type sourceDelegate struct{}

func (d sourceDelegate) Height() int                             { return 1 }
func (d sourceDelegate) Spacing() int                            { return 0 }
func (d sourceDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d sourceDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(sourceItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.Name)

	fn := sourceItemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return sourceSelectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	_, _ = fmt.Fprint(w, fn(str))
}

type sourceModel struct {
	list    list.Model
	choice  string
	config  *Config
	quiting bool
}

func (m sourceModel) Init() tea.Cmd {
	return nil
}

func (m sourceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quiting = true
			return m, tea.Quit
		case "enter":
			if i, ok := m.list.SelectedItem().(sourceItem); ok {
				m.choice = i.Name
				m.config.SourceName = i.Name
				err := m.config.writeConfig()
				if err != nil {
					fmt.Printf("Failed to save config: %v\n", err)
				}
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m sourceModel) View() string {
	if m.choice != "" {
		return sourceQuitTextStyle.Render(fmt.Sprintf("Switched to source: %s", m.choice))
	}
	if m.quiting {
		return ""
	}
	return "\n" + m.list.View()
}

func listSources(config Config) error {
	sources, err := readSources(config)
	if err != nil {
		return fmt.Errorf("failed to read source configuration: %w", err)
	}

	for _, source := range sources {
		if source.Name == config.SourceName {
			fmt.Printf("%s\n", sourceSelectedItemStyle.Render("> "+source.Name))
		} else {
			fmt.Printf("%s\n", sourceItemStyle.Render(source.Name))
		}
	}
	return nil
}
