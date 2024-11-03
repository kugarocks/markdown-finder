package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
	"github.com/sahilm/fuzzy"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

var (
	helpText = strings.TrimSpace(`
Nap is a code snippet manager for your terminal.
https://github.com/maaslalani/nap

Usage:
  nap           - for interactive mode
  nap list      - list all snippets
  nap <snippet> - print snippet to stdout

Create:
  nap < main.go                 - save snippet from stdin
  nap example/main.go < main.go - save snippet with name`)

	// 	defaultSourceConfigJson = `{
	// 	"source_list": [
	// 		{
	// 			"name": "local/default",
	// 			"repo": "https://github.com/local/default"
	// 		}
	// 	]
	// }`

	defaultSnippetConfigJson = `{
	"snippet_list": []
}`

	defaultSnippetContent = `## Quick Start

* e - edit snippet
* c/d - copy code block
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

	snippets := readSnippets(config)
	snippets = scanSnippets(config, snippets)

	stdin := readStdin()
	if stdin != "" {
		saveSnippet(stdin, args, config, snippets)
		return
	}

	if len(args) > 0 {
		switch args[0] {
		case "list":
			listSnippets(snippets)
		case "-h", "--help":
			fmt.Println(helpText)
		default:
			snippet := findSnippet(args[0], snippets)
			fmt.Print(snippet.Content(isatty.IsTerminal(os.Stdout.Fd())))
		}
		return
	}

	err = runInteractiveMode(config, snippets)
	if err != nil {
		fmt.Println("Alas, there's been an error", err)
	}
}

// parseName returns a folder, name, and language for the given name.
// this is useful for parsing file names when passed as command line arguments.
//
// Example:
//
//	Notes/Hello.go -> (Notes, Hello, go)
//	Hello.go       -> (Misc, Hello, go)
//	Notes/Hello    -> (Notes, Hello, go)
func parseName(s string) (string, string, string) {
	var (
		folder    = defaultSnippetFolder
		name      = defaultSnippetName
		language  = defaultLanguage
		remaining string
	)

	tokens := strings.Split(s, "/")
	if len(tokens) > 1 {
		folder = tokens[0]
		remaining = tokens[1]
	} else {
		remaining = tokens[0]
	}

	tokens = strings.Split(remaining, ".")
	if len(tokens) > 1 {
		name = tokens[0]
		language = tokens[1]
	} else {
		name = tokens[0]
	}

	return folder, name, language
}

// readStdin returns the stdin that is piped in to the command line interface.
func readStdin() string {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return ""
	}

	if stat.Mode()&os.ModeCharDevice != 0 {
		return ""
	}

	reader := bufio.NewReader(os.Stdin)
	var b strings.Builder

	for {
		r, _, err := reader.ReadRune()
		if err != nil && err == io.EOF {
			break
		}
		_, err = b.WriteRune(r)
		if err != nil {
			return ""
		}
	}

	return b.String()
}

// readSnippets reads the snippets file and returns the snippets
func readSnippets(config Config) []Snippet {
	var snippets []Snippet
	sourcePath := config.getSourcePath()
	file := filepath.Join(sourcePath, config.SnippetConfigFile)
	dir, err := os.ReadFile(file)
	if err != nil {
		// SnippetConfigFile does not exist, create one.
		err := os.MkdirAll(sourcePath, os.ModePerm)
		if err != nil {
			fmt.Printf("Unable to create directory %s, %+v", sourcePath, err)
		}
		f, err := os.Create(file)
		if err != nil {
			fmt.Printf("Unable to create file %s, %+v", file, err)
		}
		defer f.Close()
		dir = []byte(defaultSnippetConfigJson)
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

func saveSnippet(content string, args []string, config Config, snippets []Snippet) {
	// Save snippet to location
	name := defaultSnippetName
	if len(args) > 0 {
		name = strings.Join(args, " ")
	}

	folder, name, language := parseName(name)
	file := fmt.Sprintf("%s.%s", name, language)
	filePath := filepath.Join(config.Home, folder, file)
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		fmt.Println("unable to create folder")
		return
	}
	err := os.WriteFile(filePath, []byte(content), 0o644)
	if err != nil {
		fmt.Println("unable to create snippet")
		return
	}

	// Add snippet metadata
	snippet := Snippet{
		Folder:   folder,
		Date:     time.Now(),
		Name:     name,
		File:     file,
		Language: language,
	}

	snippets = append([]Snippet{snippet}, snippets...)
	writeSnippets(config, snippets)
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

func runInteractiveMode(config Config, snippets []Snippet) error {
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
	folderList.Title = "Folders"

	folderList.SetShowHelp(false)
	folderList.SetFilteringEnabled(false)
	folderList.SetShowStatusBar(false)
	folderList.DisableQuitKeybindings()
	folderList.Styles.NoItems = lipgloss.NewStyle().Margin(0, 2).Foreground(lipgloss.Color(config.GrayColor))
	folderList.SetStatusBarItemName("folder", "folders")

	snippetsMap := map[Folder]*list.Model{}
	for folder, items := range folders {
		snippetList := newList(items, 20, defaultStyles.Snippets.Focused)
		snippetsMap[folder] = snippetList
	}

	mdRender, _ := glamour.NewTermRenderer(
		glamour.WithStylesFromJSONFile("dark.json"),
	)

	// log file
	file, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("无法打开日志文件: %v", err)
	}
	defer file.Close()
	log.SetOutput(file)

	content := viewport.New(80, 0)
	m := &Model{
		SnippetsMap:  snippetsMap,
		Folders:      folderList,
		Code:         content,
		ContentStyle: defaultStyles.Content.Blurred,
		SnippetStyle: defaultStyles.Snippets.Focused,
		SectionStyle: defaultStyles.Sections.Blurred,
		keys:         DefaultKeyMap,
		help:         help.New(),
		config:       config,
		mdRender:     mdRender,
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
	//snippetList.SetShowStatusBar(false)
	snippetList.DisableQuitKeybindings()
	snippetList.Styles.Title = styles.Title
	snippetList.Styles.TitleBar = styles.TitleBar

	return &snippetList
}

func initDefaultSource(config Config) error {
	// 只有当配置的源名称为默认源时才进行初始化
	if config.SourceName != defaultSourceName {
		return nil
	}

	// 获取完整的源路径
	sourcePath := config.getSourcePath()

	// 构建默认片段文件夹的完整路径
	defaultFolderPath := filepath.Join(sourcePath, defaultSnippetFolder)

	// 检查并创建默认文件夹
	if _, err := os.Stat(defaultFolderPath); os.IsNotExist(err) {
		if err := os.MkdirAll(defaultFolderPath, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create default folder: %w", err)
		}
	}

	// 构建默认片段文件的完整路径
	defaultFilePath := filepath.Join(defaultFolderPath, defaultSnippetFileName)

	// 检查默认文件是否存在
	if _, err := os.Stat(defaultFilePath); os.IsNotExist(err) {
		// 创建并写入默认内容
		if err := os.WriteFile(defaultFilePath, []byte(defaultSnippetContent), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create default snippet file: %w", err)
		}
	}

	return nil
}
