package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/xatasan/kagami"
)

var (
	database, engine string
	verbose, mirror  bool
)

func help() {
	args0 := path.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, "usage:\t%s [options] <siteurl> <board>\n", args0)
	fmt.Fprintf(os.Stderr, "\t%s [options] <siteurl> <board> [thread]\n", args0)
	fmt.Fprintf(os.Stderr, "\t%s fetch [options] <siteurl> <board>\n", args0)
	fmt.Fprintf(os.Stderr, "\t%s fetch [options] <siteurl> <board> [thread]\n", args0)
	fmt.Fprintf(os.Stderr, "\t%s update\n", args0)
	fmt.Fprintf(os.Stderr, "\t%s search <interface>\n", args0)
	fmt.Fprintf(os.Stderr, "\t%s clean\n", args0)
	fmt.Fprintln(os.Stderr, "flags:")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = help
	flag.StringVar(&database, "d", "kagami.db", "file to use as metadata database")
	flag.StringVar(&engine, "e", "", "force a certain engine")
	flag.BoolVar(&verbose, "v", false, "verbose output")
	flag.BoolVar(&mirror, "m", false, "mirror static content")
	flag.Parse()
	if err := kagami.SetupDatabase(database); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	kagami.Verbose(verbose)

	args := flag.Args()
	if len(args) == 0 {
		help()
		os.Exit(1)
	}
	switch args[0] {
	case "search":
		http.HandleFunc("/", kagami.Search)
		if len(args) < 2 {
			help()
			os.Exit(1)
		}
		log.Fatal(http.ListenAndServe(args[1], nil))
	case "update":
		kagami.Update(mirror)
	case "fetch":
		fetch(args[1:])
	case "clean":
		clean()
	case "help":
		help()
	default:
		fetch(args)
		kagami.Update(mirror)
	}
}
