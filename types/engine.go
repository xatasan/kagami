package types

// an engine abstracts the communication between
// kagami and a specific image board
type Engine interface {
	Name() string
	Host() string
	Board(string) (Board, error)
}
