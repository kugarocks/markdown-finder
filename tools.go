package main

import "strings"

func getMarkdownFirstTitle(s string) string {
	title := ""
	rows := strings.Split(s, "\n")
	for _, row := range rows {
		row = strings.TrimSpace(row)
		if strings.HasPrefix(row, "#") {
			title = row
			break
		}
	}
	return strings.TrimSpace(strings.Trim(title, "#"))
}
