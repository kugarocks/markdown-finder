package main

const (
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
