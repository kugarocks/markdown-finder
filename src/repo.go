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
	defaultRepoName = "local/repo"
)

var (
	repoTitleStyle        = lipgloss.NewStyle().MarginLeft(2)
	repoItemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	repoSelectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	repoPaginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	repoHelpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	repoQuitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type Repo struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type RepoWrapper struct {
	RepoList []Repo `json:"repo_list"`
}

func readRepos(config Config) ([]Repo, error) {
	repoDir := config.getRepoBase()
	if err := os.MkdirAll(repoDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create repo directory %s: %w", repoDir, err)
	}

	configFile := filepath.Join(repoDir, config.RepoConfigFile)

	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			wrapper := RepoWrapper{
				RepoList: []Repo{},
			}

			b, err := json.MarshalIndent(wrapper, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("unable to serialize default configuration: %w", err)
			}

			err = os.WriteFile(configFile, b, os.ModePerm)
			if err != nil {
				return nil, fmt.Errorf("unable to write configuration file %s: %w", configFile, err)
			}

			return wrapper.RepoList, nil
		}
		return nil, err
	}

	var wrapper RepoWrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("unable to parse configuration file %s: %w", configFile, err)
	}

	if wrapper.RepoList == nil {
		wrapper.RepoList = []Repo{}
	}

	return wrapper.RepoList, nil
}

func writeRepos(config Config, repos []Repo) error {
	wrapper := RepoWrapper{
		RepoList: repos,
	}

	b, err := json.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to serialize configuration: %w", err)
	}
	b = append(b, '\n')

	configFile := filepath.Join(config.getRepoBase(), config.RepoConfigFile)

	err = os.WriteFile(configFile, b, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to write configuration file %s: %w", configFile, err)
	}

	return nil
}

func setRepo(config *Config) error {
	repos, err := readRepos(*config)
	if err != nil {
		return fmt.Errorf("读取源配置失败: %w", err)
	}

	// 转换源列表为 list.Item，同时找到当前源的索引
	var items []list.Item
	currentIndex := 0
	for i, repo := range repos {
		items = append(items, repoItem(repo))
		if repo.Name == config.RepoName {
			currentIndex = i
		}
	}

	// 创建列表，设置默认选中项
	l := list.New(items, repoDelegate{}, 30, 14)
	l.Title = "Choose a repo"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = repoTitleStyle
	l.Styles.PaginationStyle = repoPaginationStyle
	l.Styles.HelpStyle = repoHelpStyle

	// 设置当前选中项
	l.Select(currentIndex)

	// 运行交互程序
	p := tea.NewProgram(repoModel{
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
type repoItem Repo

func (i repoItem) FilterValue() string { return i.Name }

// 源列表代理
type repoDelegate struct{}

func (d repoDelegate) Height() int                             { return 1 }
func (d repoDelegate) Spacing() int                            { return 0 }
func (d repoDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d repoDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(repoItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.Name)

	fn := repoItemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return repoSelectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	_, _ = fmt.Fprint(w, fn(str))
}

type repoModel struct {
	list    list.Model
	choice  string
	config  *Config
	quiting bool
}

func (m repoModel) Init() tea.Cmd {
	return nil
}

func (m repoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quiting = true
			return m, tea.Quit
		case "enter":
			if i, ok := m.list.SelectedItem().(repoItem); ok {
				m.choice = i.Name
				m.config.RepoName = i.Name
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

func (m repoModel) View() string {
	if m.choice != "" {
		return repoQuitTextStyle.Render(fmt.Sprintf("Switched to repo: %s", m.choice))
	}
	if m.quiting {
		return ""
	}
	return "\n" + m.list.View()
}

func listRepos(config Config) error {
	repos, err := readRepos(config)
	if err != nil {
		return fmt.Errorf("failed to read repo configuration: %w", err)
	}

	for _, repo := range repos {
		if repo.Name == config.RepoName {
			fmt.Printf("%s\n", repoSelectedItemStyle.Render("> "+repo.Name))
		} else {
			fmt.Printf("%s\n", repoItemStyle.Render(repo.Name))
		}
	}
	return nil
}
