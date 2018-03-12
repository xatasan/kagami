package main

import (
	"flag"
	"net/url"
	"os"

	"github.com/xatasan/kagami"
)

func fetch(args []string) {
	switch len(args) {
	case 1:
		u, err := url.Parse(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: malformed URL (%s)", os.Args[0], err.String())
			os.Exit(1)
		}
		kagami.FetchUrl(engine, u, mirror)
	case 2:
		kagami.FetchBoard(args[0], args[1], mirror)
	case 3:
		kagami.FetchThread(args[0], args[1], args[2], mirror)
	default:
		help()
		os.Exit(1)
	}
}
