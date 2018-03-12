package types

import "net/url"

// one board contains all it's threads
type Board interface {
	Name() string
	Threads(chan<- Thread) error
	Thread(string) (Thread, error)
	GetFileUri(*File) *url.URL
	GetTmbUri(*File) *url.URL
	GetStaticUri(*File) *url.URL
}
