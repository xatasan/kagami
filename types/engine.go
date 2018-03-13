package types

import "net/url"

// an engine abstracts the communication between
// kagami and a specific image board
type Engine interface {
	Name() string
	Host() string
	Board(string) (Board, error)
	ReadUrl(*url.URL) (Board, Thread, error)
}
