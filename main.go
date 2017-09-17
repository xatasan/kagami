package main

import (
	"io"
	"path"
	"strings"

	"github.com/kennygrant/sanitize"

	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"sync"
)

const (
	t_dir = "./res/"  // thread directory
	i_dir = "./file/" // image directory
	T_dir = "./tmb/"  // thumbnail directory
)

var (
	t *template.Template

	w_queues = 512 // writing queues
	d_queues = 64  // thread download queues
	f_queues = 12  // file download queues
	tpp      = 200 // threads per page

	verbose = false
	debug   = false
)

func debugL(format string, v ...interface{}) {
	if debug {
		log.Printf(format, v...)
	}
}

func verboseL(format string, v ...interface{}) {
	if verbose {
		log.Printf(format, v...)
	}
}

func init() {
	fm := template.FuncMap{
		"byteSize":    byteSize,
		"shortenFile": shortenFile,
		"genTitle": func(s template.HTML) string {
			txt := sanitize.HTML(string(s))
			txt = strings.Replace(txt, "&gt;", ">", -1)
			txt = strings.Replace(txt, "&lt;", "<", -1)
			if len(txt) < 50 {
				return txt
			}
			return txt[:47] + "..."
		},
		"shortenStr": func(s template.HTML) string {
			txt := sanitize.HTML(string(s))
			txt = strings.Replace(txt, "&gt;", ">", -1)
			txt = strings.Replace(txt, "&lt;", "<", -1)
			if len(txt) < 200 {
				return txt
			}
			return txt[:197] + "..."
		},
		"add": func(a, b int) int {
			return a + b
		},
	}
	t = template.Must(template.New("").
		Funcs(fm).
		ParseGlob("./tmpl/*.tmpl"))

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options] [siteurl] [board] [name]?\n", os.Args[0])
		flag.PrintDefaults()
	}

	var database, r_dir string
	flag.IntVar(&w_queues, "W", w_queues, "number of write queues")
	flag.IntVar(&d_queues, "D", d_queues, "number of thread download queues")
	flag.IntVar(&f_queues, "F", f_queues, "number of file download queues")
	flag.IntVar(&tpp, "t", tpp, "threads per catalog page")
	flag.StringVar(&database, "db", "kagami.db", "use file as sqlite database")
	flag.StringVar(&r_dir, "o", "out", "output directory")
	flag.BoolVar(&verbose, "v", false, "output verbosely")
	flag.BoolVar(&debug, "d", false, "output for debugging")
	flag.Parse()

	for _, d := range []string{
		r_dir,
		path.Join(r_dir, t_dir),
		path.Join(r_dir, i_dir),
		path.Join(r_dir, T_dir),
	} {
		err := os.MkdirAll(d, os.ModeDir|0755)
		if err != nil && !os.IsExist(err) {
			log.Fatal(err)
		}
	}
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	for _, fn := range []string{"style.css", "kagami.js"} {
		from := path.Join(cwd, fn)
		to := path.Join(r_dir, fn)
		if _, err := os.Stat(to); err != nil && !os.IsExist(err) {
			f, err := os.Open(from)
			if err != nil { // from
				log.Fatal(err)
			}
			t, err := os.Create(to)
			if err != nil { // to
				log.Fatal(err)
				f.Close()
			}
			if _, err := io.Copy(t, f); err != nil {
				log.Fatal(err)
			}
			t.Close()
			f.Close()
		}
	}

	if err := setupdb(database); err != nil {
		log.Fatal(err)
	}
	verboseL("Using %q as database", database)

	if err = os.Chdir(r_dir); err != nil {
		log.Fatal(err)
	}
}

func main() {
	if flag.NArg() < 2 {
		flag.Usage()
		os.Exit(1)
	}

	var (
		towrite  = make(chan Thread)
		tosave   = make(chan Thread)
		todl     = make(chan struct{ n, l float64 })
		todlfile = make(chan File)
		wg1, wg2 sync.WaitGroup

		engine = getEngine(flag.Arg(0))
		board  = flag.Arg(1)
		name   = flag.Arg(2)
	)
	verboseL("Using %q engine; board: /%s/\n", engine.name, board)

	for _, f := range []string{"spoiler.png", "deleted.png", "file.png"} {
		if err := dl(f, engine.getStatic(board, f)); err != nil {
			log.Fatal(err)
		}
	}

	for _, F := range []string{"sfiles", "sthreads", "about"} {
		f, err := os.Create("./" + F + ".html")
		if err != nil {
			log.Fatal(err)
		}

		pr, pw := io.Pipe()
		defer pr.Close()
		go func() {
			defer pw.Close()
			err = t.Lookup(F+".tmpl").Execute(pw, struct {
				N, B string
			}{name, board})
			if err != nil {
				log.Fatal(err)
			}
		}()
		m.Minify("text/html", f, pr)
		f.Close()
	}

	go func() {
		var wg sync.WaitGroup
		wg.Add(d_queues)
		for i := 0; i < d_queues; i++ {
			go func() {
				for thread := range todl {
					debugL("[tp/%05d] Processing thread %d\n", getGID(), int(thread.n))

					t, err := processThread(engine, board, thread)
					if err != nil {
						log.Printf("Error when processing thread %d: %s\n", int(thread.n), err.Error())
					} else if t != nil {
						towrite <- t
						tosave <- t
					}

					debugL("[tp/%05d] Finished thread %d\n", getGID(), int(thread.n))
				}
				wg.Done()
			}()
		}

		if err := engine.getThreads(board, todl); err != nil {
			log.Fatal(err)
		}
		wg.Wait()
		close(tosave)
		close(towrite)
		debugL("Closed \"tosave\" channel")
		debugL("Closed \"towrite\" channel")
	}()

	debugL("Starting %d file downloading threads\n", f_queues)
	wg1.Add(f_queues)
	for i := 0; i < f_queues; i++ {
		go FDLqueue(todlfile, engine, &wg1)
	}

	debugL("Starting %d file thread threads\n", w_queues)
	wg2.Add(w_queues)
	for i := 0; i < w_queues; i++ {
		go save2file(towrite, todlfile, board, name, &wg2)
	}

	defer db.Close()
	if err := write2db(tosave); err != nil {
		log.Fatal(err)
	}
	debugL("Closed \"todlfile\" channel")
	if err := mkCatalog(name, board); err != nil {
		log.Fatal(err)
	}

	wg2.Wait()
	close(todlfile)
	wg1.Wait()
	verboseL("Done")
}
