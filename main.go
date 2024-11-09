package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

const (
	Version           = "v1.0.0"
	githubSSHPrefix   = "git@github.com:"
	githubHTTPSPrefix = "https://github.com/"
	githubSSHSuffix   = ".git"
)

func main() {
	runCLI(os.Args[1:])
}

func runCLI(args []string) {
	config := readConfig()

	err := initDefaultSource(config)
	if err != nil {
		fmt.Println("Init default source failed", err)
		return
	}

	validateSourceName(&config)
	snippets := readSnippets(config)
	snippets = scanSnippets(config, snippets)

	initFolderName(&config, snippets)

	var targetSnippet Snippet
	if len(args) > 1 {
		switch args[0] {
		case "list":
			if strings.Contains(args[1], "source") {
				if err = listSources(config); err != nil {
					fmt.Println(err)
				}
				return
			} else if strings.Contains(args[1], "folder") {
				if err = listFolders(config, snippets); err != nil {
					fmt.Println(err)
				}
				return
			} else if strings.Contains(args[1], "snippet") {
				listSnippets(snippets)
			}
		case "get":
			if len(args) < 3 || args[1] != "source" {
				fmt.Println("Usage: mdf get source <user/repo>")
				return
			}
			err = getSource(config, args[2])
			if err != nil {
				fmt.Printf("Failed to get source: %v\n", err)
			}
			return
		case "set":
			if strings.Contains(args[1], "source") {
				if err = setSource(&config); err != nil {
					fmt.Printf("set source failed: %v\n", err)
				}
				return
			} else if strings.Contains(args[1], "folder") {
				if err = setFolder(&config, snippets); err != nil {
					fmt.Printf("set folder failed: %v\n", err)
				}
				return
			}
		default:
			fmt.Println("Unknown command")
		}
		return
	} else if len(args) == 1 {
		switch args[0] {
		case "-h", "--help":
			fmt.Print(HelpText)
			return
		case "-v", "--version", "version":
			fmt.Println(Version)
			return
		default:
			targetSnippet = findSnippet(args[0], snippets)
		}
	}

	err = runInteractiveMode(config, snippets, targetSnippet)
	if err != nil {
		fmt.Println("Alas, there's been an error", err)
	}
}

// readSnippets reads the snippets file and returns the snippets
func readSnippets(config Config) []Snippet {
	var snippets []Snippet
	sourcePath := config.getSourcePath()
	file := filepath.Join(sourcePath, config.SnippetConfigFile)
	dir, err := os.ReadFile(file)
	if err != nil {
		// SnippetConfigFile does not exist, create one.
		err = os.MkdirAll(sourcePath, os.ModePerm)
		if err != nil {
			fmt.Printf("Unable to create directory %s, %+v", sourcePath, err)
		}
		f, err := os.Create(file)
		if err != nil {
			fmt.Printf("Unable to create file %s, %+v", file, err)
		}
		defer f.Close()
		dir = []byte(DefaultSnippetConfig)
		_, _ = f.Write(dir)
	}

	var wrapper SnippetsWrapper
	err = json.Unmarshal(dir, &wrapper)
	if err != nil {
		fmt.Printf("Unable to unmarshal %s file, %+v\n", file, err)
		return snippets
	}
	return wrapper.SnippetList
}

// scanSnippets scans for any new/removed snippets and adds them to snippet-config.json
func scanSnippets(config Config, snippets []Snippet) []Snippet {
	var modified bool
	snippetExists := func(path string) bool {
		for _, snippet := range snippets {
			if path == snippet.Path() {
				return true
			}
		}
		return false
	}

	sourcePath := config.getSourcePath()
	homeEntries, err := os.ReadDir(sourcePath)
	if err != nil {
		fmt.Printf("could not scan config home: %v\n", err)
		return snippets
	}

	for _, homeEntry := range homeEntries {
		if !homeEntry.IsDir() {
			continue
		}
		if strings.HasPrefix(homeEntry.Name(), ".") {
			continue
		}

		folderPath := filepath.Join(sourcePath, homeEntry.Name())
		folderEntries, err := os.ReadDir(folderPath)
		if err != nil {
			fmt.Printf("could not scan %q: %v\n", folderPath, err)
			continue
		}

		for _, folderEntry := range folderEntries {
			if folderEntry.IsDir() {
				continue
			}

			snippetPath := filepath.Join(homeEntry.Name(), folderEntry.Name())
			if !snippetExists(snippetPath) {
				name := folderEntry.Name()
				ext := filepath.Ext(name)
				snippets = append(snippets, Snippet{
					Folder:   homeEntry.Name(),
					Date:     time.Now(),
					Name:     strings.TrimSuffix(name, ext),
					File:     name,
					Language: strings.TrimPrefix(ext, "."),
				})
				modified = true
			}
		}
	}

	var idx int
	for _, snippet := range snippets {
		snippetPath := filepath.Join(sourcePath, snippet.Path())
		if _, err := os.Stat(snippetPath); !errors.Is(err, fs.ErrNotExist) {
			snippets[idx] = snippet
			idx++
			modified = true
		}
	}
	snippets = snippets[:idx]

	if modified {
		writeSnippets(config, snippets)
	}

	return snippets
}

// writeSnippets writes the snippets to the snippets file
func writeSnippets(config Config, snippets []Snippet) {
	wrapper := SnippetsWrapper{
		SnippetList: snippets,
	}

	b, err := json.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		fmt.Println("Could not marshal latest snippet data.", err)
		return
	}
	b = append(b, '\n')

	sourcePath := config.getSourcePath()
	err = os.WriteFile(filepath.Join(sourcePath, config.SnippetConfigFile), b, os.ModePerm)
	if err != nil {
		fmt.Println("Could not save snippets file.", err)
	}
}

func listSnippets(snippets []Snippet) {
	for _, snippet := range snippets {
		fmt.Println(snippet)
	}
}

func findSnippet(search string, snippets []Snippet) Snippet {
	matches := fuzzy.FindFrom(search, Snippets{snippets})
	if len(matches) > 0 {
		return snippets[matches[0].Index]
	}
	return Snippet{}
}

func runInteractiveMode(config Config, snippets []Snippet, targetSnippet Snippet) error {
	if len(snippets) == 0 {
		// welcome to nap!
		snippets = append(snippets, defaultSnippet)
	}

	folders := make(map[Folder][]list.Item)
	for _, snippet := range snippets {
		folders[Folder(snippet.Folder)] = append(folders[Folder(snippet.Folder)], list.Item(snippet))
	}

	defaultStyles := DefaultStyles(config)

	var folderItems []list.Item
	foldersSlice := maps.Keys(folders)
	slices.Sort(foldersSlice)
	for _, folder := range foldersSlice {
		folderItems = append(folderItems, list.Item(folder))
	}
	if len(folderItems) <= 0 {
		folderItems = append(folderItems, list.Item(Folder(defaultSnippetFolder)))
	}
	folderList := list.New(folderItems, folderDelegate{defaultStyles.Folders.Blurred}, 0, 0)

	for idx, folder := range foldersSlice {
		if string(folder) == targetSnippet.Folder {
			folderList.Select(idx)
			break
		}
		if string(folder) == config.FolderName {
			folderList.Select(idx)
			break
		}
	}

	hideSnippetPane := false
	selectedFolder := folderList.SelectedItem().(Folder)
	snippetsMap := map[Folder]*list.Model{}
	for folder, items := range folders {
		snippetList := newList(items, 20, defaultStyles.Snippets.Focused)
		snippetsMap[folder] = snippetList
		if folder == selectedFolder {
			for idx, item := range snippetList.Items() {
				if s, ok := item.(Snippet); ok && s.File == targetSnippet.File {
					snippetList.Select(idx)
					hideSnippetPane = true
					break
				}
			}
		}
	}

	if config.AlwaysShowSnippetPane {
		hideSnippetPane = false
	}

	mdRender, _ := glamour.NewTermRenderer(
		glamour.WithStyles(defaultStyles.Glamour["dark"]),
	)

	content := viewport.New(80, 0)
	m := &Model{
		SnippetsMap:     snippetsMap,
		Folders:         folderList,
		Code:            content,
		help:            help.New(),
		config:          config,
		mdRender:        mdRender,
		hideSnippetPane: hideSnippetPane,
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	model, err := p.Run()
	if err != nil {
		return err
	}
	fm, ok := model.(*Model)
	if !ok {
		return err
	}
	var allSnippets []Snippet
	for _, snippetList := range fm.SnippetsMap {
		for _, item := range snippetList.Items() {
			allSnippets = append(allSnippets, item.(Snippet))
		}
	}
	writeSnippets(config, allSnippets)
	return nil
}

func newList(items []list.Item, height int, styles SnippetsBaseStyle) *list.Model {
	snippetList := list.New(items, snippetDelegate{snippetPane, styles, navigatingState}, 25, height)
	snippetList.SetShowHelp(false)
	snippetList.SetShowFilter(false)
	snippetList.SetShowTitle(false)
	snippetList.Styles.StatusBar = lipgloss.NewStyle().Margin(1, 3).Foreground(lipgloss.Color("240")).MaxWidth(35 - 2)
	snippetList.Styles.NoItems = lipgloss.NewStyle().Margin(0, 3).Foreground(lipgloss.Color("8")).MaxWidth(35 - 2)
	snippetList.FilterInput.Prompt = "Find: "
	snippetList.FilterInput.PromptStyle = styles.Title
	snippetList.SetStatusBarItemName("Snippet", "Snippets")
	snippetList.DisableQuitKeybindings()
	snippetList.Styles.Title = styles.Title
	snippetList.Styles.TitleBar = styles.TitleBar

	return &snippetList
}

func initDefaultSource(config Config) error {
	// Read existing sources
	sources, err := readSources(config)
	if err != nil {
		return fmt.Errorf("failed to read sources: %w", err)
	}

	// Check if default source already exists
	for _, source := range sources {
		if source.Name == defaultSourceName {
			return nil // Already exists, nothing to do
		}
	}

	// Add default source
	defaultSource := Source{
		Name: defaultSourceName,
		Repo: "", // Local source has no repo URL
	}
	sources = append(sources, defaultSource)

	// Save updated sources
	if err := writeSources(config, sources); err != nil {
		return fmt.Errorf("failed to save sources: %w", err)
	}

	// Create default folder and file structure
	sourcePath := config.getDefaultSourcePath()
	defaultFolderPath := filepath.Join(sourcePath, defaultSnippetFolder)

	if err := os.MkdirAll(defaultFolderPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create default folder: %w", err)
	}

	defaultFilePath := filepath.Join(defaultFolderPath, defaultSnippetFileName)
	if _, err := os.Stat(defaultFilePath); os.IsNotExist(err) {
		if err := os.WriteFile(defaultFilePath, []byte(DefaultSnippetContent), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create default snippet file: %w", err)
		}
	}

	return nil
}

func parseGitHubURL(repoURL string) (user, repo string, err error) {
	// Supported formats:
	// https://github.com/user/repo.git
	// https://github.com/user/repo
	// git@github.com:user/repo.git
	// user/repo

	// Remove .git suffix
	repoURL = strings.TrimSuffix(repoURL, githubSSHSuffix)

	// Handle SSH format
	if strings.HasPrefix(repoURL, githubSSHPrefix) {
		parts := strings.Split(strings.TrimPrefix(repoURL, githubSSHPrefix), "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid GitHub repository URL: %s", repoURL)
		}
		return parts[0], parts[1], nil
	}

	// Handle HTTPS format
	if strings.HasPrefix(repoURL, githubHTTPSPrefix) {
		parts := strings.Split(strings.TrimPrefix(repoURL, githubHTTPSPrefix), "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid GitHub repository URL: %s", repoURL)
		}
		return parts[0], parts[1], nil
	}

	// Handle short format user/repo
	parts := strings.Split(repoURL, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid GitHub repository URL: %s", repoURL)
	}
	return parts[0], parts[1], nil
}

func getSource(config Config, repoURL string) error {
	// Parse GitHub repository URL
	user, repo, err := parseGitHubURL(repoURL)
	if err != nil {
		return err
	}

	// Determine the clone URL based on the input format
	var cloneURL string
	switch {
	case strings.HasPrefix(repoURL, githubSSHPrefix):
		cloneURL = repoURL
	case strings.HasPrefix(repoURL, githubHTTPSPrefix):
		cloneURL = repoURL
	default:
		cloneURL = fmt.Sprintf("git@github.com:%s/%s.git", user, repo)
	}

	// Read existing sources
	sources, err := readSources(config)
	if err != nil {
		return fmt.Errorf("failed to read source configuration: %w", err)
	}

	// Check if the source already exists
	sourceName := fmt.Sprintf("%s/%s", user, repo)
	for _, source := range sources {
		if source.Name == sourceName {
			return fmt.Errorf("source %s already exists", sourceName)
		}
	}

	// Create source directory
	sourcePath := filepath.Join(config.getSourceBase(), user, repo)
	err = os.MkdirAll(sourcePath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create source directory: %w", err)
	}

	cmd := exec.Command("git", "clone", cloneURL, sourcePath)

	// Set up pipes for real-time output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Add new source to the list
	newSource := Source{
		Name: sourceName,
		Repo: cloneURL,
	}
	sources = append(sources, newSource)

	// Save configuration
	err = writeSources(config, sources)
	if err != nil {
		return fmt.Errorf("failed to save source configuration: %w", err)
	}

	fmt.Printf("Successfully added source: %s\n", sourceName)
	return nil
}

// validateSourceName if the source directory does not exist, switch to the default source
func validateSourceName(config *Config) {
	sourcePath := config.getSourcePath()
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		config.SourceName = defaultSourceName
	}
}

func initFolderName(config *Config, snippets []Snippet) {
	// create folder name list and sort
	folderSet := make(map[string]struct{})
	for _, snippet := range snippets {
		folderSet[snippet.Folder] = struct{}{}
	}

	folderNameList := make([]string, 0, len(folderSet))
	for folder := range folderSet {
		folderNameList = append(folderNameList, folder)
	}
	slices.Sort(folderNameList)

	if len(folderNameList) == 0 {
		return
	}

	// if FolderName is empty, use the first folder
	if strings.TrimSpace(config.FolderName) == "" {
		config.FolderName = folderNameList[0]
		config.writeConfig()
		return
	}

	// validate folder name
	found := false
	for _, folder := range folderNameList {
		if folder == config.FolderName {
			found = true
			break
		}
	}

	// if not found, use the first folder
	if !found {
		config.FolderName = folderNameList[0]
		config.writeConfig()
	}
}
