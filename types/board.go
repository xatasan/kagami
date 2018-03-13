package types

import (
	"net/url"
	"sync"
)

// one board contains all it's threads
type Board interface {
	Name() string
	Threads(chan<- Thread) (*sync.WaitGroup, error)
	Thread(string) (Thread, error)
	GetFileUri(*File) *url.URL
	GetTmbUri(*File) *url.URL
	GetStaticUri(*File) *url.URL
}
