package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"time"
)

func processThread(board string, t struct{ n, l float64 }) (Thread, error) {
	no := fmt.Sprintf("%d", int(t.n))
	last_modified := time.Unix(int64(t.l), 0)

	file, err := os.Open(t_dir + no + ".html")
	defer file.Close()
	if os.IsNotExist(err) {
		return engine.genThread(board, no)
	} else if err != nil {
		log.Fatal(err)
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		log.Fatal(err)
		return nil, err

	} else if stat.ModTime().Before(last_modified) {
		return engine.genThread(board, no)
	}

	debugL("[pt/%05d] Thread %s already exits\n", getGID(), no)
	return nil, nil // don't do anything
}
