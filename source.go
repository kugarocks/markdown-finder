package main

const (
	defaultSourceName = "local/source"
)

type Source struct {
	Name string `json:"name"`
	Repo string `json:"repo"`
}
