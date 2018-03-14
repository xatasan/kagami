package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/xatasan/kagami"
)

func fetch(args []string) {
	var err error
	switch len(args) {
	case 1:
		var u *url.URL
		u, err = url.Parse(args[0])
		if err != nil {
			err = fmt.Errorf("%s: malformed URL (%v)", os.Args[0], err)
		} else {
			err = kagami.FetchUrl(engine, u)
		}
	case 2:
		err = kagami.FetchBoard(args[0], args[1])
	case 3:
		err = kagami.FetchThread(args[0], args[1], args[2])
	default:
		help()
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
