package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

var files = []string{
	"kagami.db",
	"kagami.js",
	"search.js",
	"style.css",
	"index.html",
	"tmb/",
	"res/",
	"file/",
}

func clean() {
	fmt.Fprintf(os.Stderr, "Are you sure you want to delete everything? [y/N] ")

	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	resp := make([]byte, 1)
	os.Stdin.Read(resp)
	if resp[0] == 'y' || resp[0] == 'Y' {
		for _, f := range files {
			log.Println("Deleting " + f)
			os.RemoveAll(f)
		}
	} else {
		log.Println("Not deleting anything")
	}
}
