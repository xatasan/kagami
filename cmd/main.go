package main

import (
	"flag"
	"os"

	"github.com/xatasan/kagami"
)

var (
	database, engine string
	verbose, mirror  bool
)

func help() {
	args0 := os.Args[0]
	fmt.Fprintf(os.Stderr, "usage:\t%s [options] <siteurl> <board>\n", args0)
	fmt.Fprintf(os.Stderr, "\t\t%s [options] <siteurl> <board> <thread>?\n", args0)
	fmt.Fprintf(os.Stderr, "\t\t%s fetch [options] <siteurl> <board>\n", args0)
	fmt.Fprintf(os.Stderr, "\t\t%s fetch [options] <siteurl> <board> <thread>?\n", args0)
	fmt.Fprintf(os.Stderr, "\t\t%s update\n", args0)
	fmt.Fprintf(os.Stderr, "\t\t%s search <interface>\n", args0)
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
	kagami.SetupDatabase(database)
	kagami.Verbose(verbose)

	if len(flag.Args) < 2 {
		fetch(os.Args)
	} else {
		switch flag.Args[0] {
		case "help":
			help()
		case "search":
			http.HandleFunc("/", kagami.Search)
			if len(flag.Args) < 2 {
				help()
				os.Exit(1)
			}
			log.Fatal(http.ListenAndServe(flag.Args[1], nil))
		case "update":
			kagami.Update()
		case "fetch":
			fetch(os.Args[1:])
		default:
			fetch(os.Args)
		}
	}
}
