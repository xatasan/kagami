package main

import (
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/html"

	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

var m *minify.M

func init() {
	m = minify.New()
	m.AddFunc("text/html", html.Minify)
}

func getFile(e Engine, file File) error {
	local := fmt.Sprintf("%s/%s", i_dir, file.Filename)
	remote := e.getFile(file.Filename)
	return dl(local, remote)
}

func getThumbnail(e Engine, file File) error {
	local := fmt.Sprintf("%s/%s", T_dir, file.Thumbnail)
	remote := e.getTmb(file.Thumbnail)
	if remote == "" {
		return nil
	}
	return dl(local, remote)
}

func FDLqueue(dl <-chan File, e Engine, wg *sync.WaitGroup) {
	defer wg.Done()
	for file := range dl {
		debugL("[fq/%05d] Getting file(s) %s\n", getGID(), file.Filename)
		err := getFile(e, file)
		if err != nil {
			log.Fatal(err)
		}
		err = getThumbnail(e, file)
		if err != nil {
			log.Fatal(err)
		}
		debugL("[fq/%05d] Saved file(s) %s\n", getGID(), file.Filename)
	}
}

func save2file(write <-chan Thread, dl chan<- File, board, name string, wg *sync.WaitGroup) {
	defer wg.Done()

	for T := range write {
		t_id := T[0].PostNumber
		debugL("[sf/%05d] Saving thread %d\n", getGID(), t_id)
		f, err := os.Create(fmt.Sprintf("%s%d.html", t_dir, t_id))
		if err != nil {
			log.Fatal(err)
		}

		pr, pw := io.Pipe()
		defer pr.Close()
		go func() {
			err = t.Lookup("thread.tmpl").Execute(pw, struct {
				F Post          // head of the thread
				T Thread        // thread without head
				B string        // board name
				N template.HTML // mirror name
				U time.Time     // time generated
			}{T[0], T[1:], board, template.HTML(name), time.Now()})
			if err != nil {
				log.Fatal(err)
			}
			debugL("[sf/%05d] Generated thread %d\n", getGID(), t_id)
			pw.Close()
		}()

		m.Minify("text/html", f, pr)
		f.Close()

		for _, post := range T {
			for _, file := range post.Files {
				dl <- file
			}
		}

		debugL("[sf/%05d] Minified thread %d\n", getGID(), t_id)
	}
}
