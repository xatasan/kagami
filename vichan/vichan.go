package vichan

import (
	"encoding/json"
	"fmt"
	"html/template"
	img "image"
	"log"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"time"

	k "github.com/xatasan/kagami/types"
)

const (
	threads = "https://%s/%s/threads.json"
	thread  = "https://%s/%s/res/%s.json"
	image   = "https://%s/src/%s"
	file    = "https://%s/static/file.jpg"
	static  = "https://%s/static/%s"
)

var linkRegexp *regexp.Regexp

func init() {
	linkRegexp = regexp.MustCompile(`<a onclick="highlightReply\('(\d+)'(?:, event)?\);"\s+href="/(\w+)/res/(\d+).html#\d+">`)
}

type engine struct{ host string }

func (e engine) Name() string {
	return "vichan"
}

func (e engine) Host() string {
	return e.host
}

func (e engine) Board(b string) (k.Board, error) {
	return board{e.host, b}, nil
}

type board struct{ host, name string }

func (b board) Name() string {
	return b.name
}

func (b board) Threads(ch chan<- k.Thread) error {
	threads, err := ThreadList(b.host, b.name)
	for _, threadId := range threads {
		go func(n string) {
			t, err := b.Thread(n)
			if err != nil {
				log.Print("Error when fetching threads: ", err.Error())
			} else {
				ch <- t

			}
		}(threadId)
	}
	return err
}

func (b board) Thread(name string) (k.Thread, error) {
	var data map[string][]struct {
		No          float64 `json:"no"`
		Resto       float64 `json:"resto"`
		Sticky      float64 `json:"sticky"`
		Closed      float64 `json:"closed"`
		Time        float64 `json:"time"`
		Name        string  `json:"name"`
		Trip        string  `json:"trip"`
		Id          string  `json:"id"`
		Capcode     string  `json:"capcode"`
		Country     string  `json:"country"`
		CountryName string  `json:"country_name"`
		Sub         string  `json:"sub"`
		Com         string  `json:"com"`
		Tim         string  `json:"tim"`
		Filename    string  `json:"filename"`
		Ext         string  `json:"ext"`
		Fsize       float64 `json:"fsize"`
		Md5         string  `json:"md5"`
		Width       float64 `json:"w"`
		Height      float64 `json:"h"`
		TmbWidth    float64 `json:"tn_w"`
		TmbHeight   float64 `json:"tn_h"`
		FileDeleted float64 `json:"filedeleted"`
		Spoiler     float64 `json:"spoiler"`
		Images      float64 `json:"images"`
	}

	// request /res/%d.json
	req := fmt.Sprintf(thread, b.name, b.name, name)
	resp, err := http.Get(req)
	if err != nil {
		return nil, err
	}

	// decode JSON structure from /res/%d.json
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&data)
	if err != nil {
		return nil, err
	}

	var (
		T       k.Thread
		postMap = make(map[int]*k.Post)
		flagMap = make(map[string]*k.Flag)
		first   = true
	)

	for _, post := range data["posts"] {
		var files []*k.File
		if post.Filename != "" {
			files = append(files, &k.File{
				Filename:     post.Tim + post.Ext,
				OrigFilename: post.Filename,
				FileSize:     int(post.Fsize),
				FileMD5:      post.Md5,
				Image: img.Point{
					int(post.Width),
					int(post.Height),
				},
				ThumbnailName: post.Tim + "s.jpg",
				Thumbnail: img.Point{
					int(post.TmbWidth),
					int(post.TmbHeight),
				},
				FileDeleted: post.FileDeleted != 0,
				Spoiler:     post.Spoiler != 0,
			})
		}

		com := linkRegexp.ReplaceAllString(post.Com,
			`<a class="r" href="./res/$3.html#$1">`)

		var p *k.Post

		p.PostNumber = int(post.No)
		p.Sticky = post.Sticky != 0
		p.Closed = post.Sticky != 0
		p.OP = first
		p.Time = time.Unix(int64(post.Time), 0)
		p.Name = post.Name
		p.Tripcode = post.Trip
		p.Id = post.Id
		p.Capcode = post.Capcode
		p.Subject = post.Sub
		p.Comment = template.HTML(com)
		p.Files = files

		postMap[p.PostNumber] = p
		p.ReplyTo = postMap[int(post.Resto)]
		p.ReplyTo.QuotedBy = append(p.ReplyTo.QuotedBy, p)

		if flag, ok := flagMap[post.CountryName]; ok {
			p.Flag = flag
		} else {
			p.Flag = &k.Flag{post.Country, post.CountryName}
			flagMap[post.CountryName] = p.Flag
		}

		T = append(T, p)
		first = false
	}

	return T, nil
}

func (b board) GetFileUri(f *k.File) *url.URL {
	path := fmt.Sprintf(image, b.host, f.Filename)
	u, _ := url.Parse(path)
	return u
}

func (b board) GetTmbUri(f *k.File) *url.URL {
	var uri string
	switch path.Ext(f.Filename) {
	case ".jpeg", ".jpg", ".png":
		uri = fmt.Sprintf(image, b.host, f.ThumbnailName)
	default:
		uri = fmt.Sprintf(file, b.host)
	}
	u, _ := url.Parse(uri)
	return u
}

func (b board) GetStaticUri(f *k.File) *url.URL {
	u, _ := url.Parse(fmt.Sprintf(static, b.host, f))
	return u
}

func Engine(host string) k.Engine {
	return engine{host}
}

func ThreadList(host, board string) ([]string, error) {
	req := fmt.Sprintf(threads, host, board)
	resp, err := http.Get(req)
	if err != nil {
		return nil, err
	} else if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Failed to get threads: %d", resp.StatusCode)
	}

	var data []struct {
		Threads []map[string]float64 `json:"threads"`
		Page    float64              `json:"page"`
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&data)
	if err != nil {
		return nil, err
	}

	var T []string
	for _, page := range data {
		for _, threads := range page.Threads {
			T = append(T, fmt.Sprintf("%.f",
				threads["no"]))
		}
	}
	return T, nil
}
