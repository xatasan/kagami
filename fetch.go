package kagami

import (
	"fmt"
	"net/url"

	k "github.com/xatasan/kagami/types"

	"github.com/xatasan/kagami/infchan"
	"github.com/xatasan/kagami/vichan"
)

// returns an engine, with a specific code
// name for a specific host
func getEngine(engine, host string) k.Engine {
	switch engine {
	case "8ch", "8chan", "8ch.net":
		return infchan.Engine()
	case "vichan":
		return vichan.Engine(host)
	default:
		return nil
	}
}

func FetchUrl(engine string, u *url.URL) error {
	var e k.Engine
	if engine != "" {
		e = getEngine(engine, u.Host)
	} else {
		e = getEngine(u.Host, u.Host)
	}
	if e == nil {
		return fmt.Errorf("Host not supported")
	}

	b, t, err := e.ReadUrl(u)
	if err != nil {
		return err
	} else if t != nil && b != nil {
		return saveThread(t)
	} else if b != nil {
		return saveBoard(b)
	} // else
	panic("ReadUrl gave invalid response")
}

func FetchBoard(host, board string) error {
	e := getEngine(host, host)
	if e == nil {
		return fmt.Errorf("Host not supported")
	}

	b, err := e.Board(board)
	if err != nil {
		return err
	}
	return saveBoard(b)
}

func FetchThread(host, board, post string) error {
	e := getEngine(host, host)
	if e == nil {
		return fmt.Errorf("Host not supported")
	}

	b, err := e.Board(board)
	if err != nil {
		return err
	}

	t, err := b.Thread(post)
	if err != nil {
		return err
	}
	return saveThread(t)
}
