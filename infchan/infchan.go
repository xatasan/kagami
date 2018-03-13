package infchan

import (
	"encoding/json"
	"fmt"
	"html/template"
	"image"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	k "github.com/xatasan/kagami/types"
	"github.com/xatasan/kagami/vichan"
)

const (
	host    = "8ch.net"
	asset   = "https://8ch.net/static/assets/%s/%s"
	file    = "https://media.8ch.net/file_store/"
	filePng = "https://8ch.net/static/file.png"
	thumb   = "https://media.8ch.net/file_store/thumb/%s"
	thread  = "https://8ch.net/%s/res/%s.json"
)

type engine struct{}

func (e engine) Name() string {
	return "8chan"
}

func (e engine) Host() string {
	return host
}

func (e engine) Board(brd string) (k.Board, error) {
	return board{brd}, nil
}

func (e engine) ReadUrl(u *url.URL) (k.Board, k.Thread, error) {
	if u.Host != "8ch.net" {
		return nil, nil,
			fmt.Errorf("URL contains wrong host")
	}

	match := vichan.ThreadReg.FindStringSubmatch(u.Path)
	if len(match) == 4 {
		brd, _ := e.Board(match[1])
		thr, err := brd.Thread(match[2])
		return brd, thr, err
	} else if len(match) >= 2 {
		brd, _ := e.Board(match[1])
		return brd, nil, nil
	} // else
	return nil, nil, fmt.Errorf("invalid URL path")
}

type board struct{ name string }

func (b board) Name() string {
	return b.name
}

func (b board) Threads(ch chan<- k.Thread) (*sync.WaitGroup, error) {
	threads, err := vichan.ThreadList(host, b.name)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	wg.Add(len(threads))
	for _, threadId := range threads {
		go func(n string) {
			t, err := b.Thread(n)
			if err != nil {
				log.Print("Error when fetching threads: ", err.Error())
			} else {
				ch <- t
			}
			wg.Done()
		}(threadId)
	}
	return &wg, nil
}

func (b board) Thread(name string) (k.Thread, error) {
	type efile struct {
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
	}

	var data map[string][]struct {
		efile
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
		ExtraFiles  []efile `json:"extra_files"`
	}

	req := fmt.Sprintf(thread, b.name, name)
	resp, err := http.Get(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Problematic status: %s", resp.Status)
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
				OrigFilename: post.Filename + post.Ext,
				FileSize:     int(post.Fsize),
				FileMD5:      post.Md5,
				Image: image.Point{
					int(post.Width),
					int(post.Height),
				},
				ThumbnailName: genThumbnail(post.Tim + post.Ext),
				Thumbnail: image.Point{
					int(post.TmbWidth),
					int(post.TmbHeight),
				},
				FileDeleted: post.FileDeleted != 0 || post.Ext == "deleted",
				Spoiler:     post.Spoiler != 0,
			})
		}
		for _, file := range post.ExtraFiles {
			files = append(files, &k.File{
				Filename:     file.Tim + file.Ext,
				OrigFilename: file.Filename + file.Ext,
				FileSize:     int(file.Fsize),
				FileMD5:      file.Md5,
				Image: image.Point{
					int(file.Width),
					int(file.Height),
				},
				ThumbnailName: genThumbnail(file.Tim + file.Ext),
				Thumbnail: image.Point{
					int(file.Width),
					int(file.Height),
				},
				FileDeleted: file.FileDeleted != 0 || post.Ext == "deleted",
				Spoiler:     file.Spoiler != 0,
			})
		}

		//quotes := vichan.LinkReg.FindAllStringSubmatch(post.Com, -1)
		com := vichan.LinkReg.ReplaceAllString(post.Com,
			`<a class="r" href="./$3.html#$1">`)

		p := &k.Post{
			PostNumber: int(post.No),
			Sticky:     post.Sticky != 0,
			Closed:     post.Sticky != 0,
			OP:         first,
			Time:       time.Unix(int64(post.Time), 0),
			Name:       post.Name,
			Tripcode:   post.Trip,
			Id:         post.Id,
			Capcode:    post.Capcode,
			Subject:    post.Sub,
			Comment:    template.HTML(com),
			Files:      files,
		}

		postMap[p.PostNumber] = p
		pmap, ok := postMap[int(post.Resto)]
		p.ReplyTo = pmap
		if ok {
			p.ReplyTo.QuotedBy = append(p.ReplyTo.QuotedBy, p)
		}

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

func (_ board) GetFileUri(f *k.File) *url.URL {
	path := file
	path += f.Filename
	u, _ := url.Parse(path)
	return u
}

func (b board) GetTmbUri(f *k.File) *url.URL {
	switch path.Ext(f.Filename) {
	case ".jpeg", ".jpg", ".png", ".gif", ".mp4", ".webm":
		tmb := genThumbnail(f.Filename)
		if tmb != "" {
			path := fmt.Sprintf(thumb, tmb)
			u, _ := url.Parse(path)
			return u
		}
	}
	return nil //"https://8ch.net/static/file.png"
}

func (b board) GetStaticUri(f *k.File) *url.URL {
	var path string
	if f.OrigFilename == "file.png" {
		path = filePng
	} else {
		path = fmt.Sprintf(asset, b.name, file)
	}
	u, _ := url.Parse(path)
	return u
}

func Engine() k.Engine {
	return engine{}
}

func genThumbnail(file string) string {
	ext := path.Ext(file)
	switch ext {
	case ".jpeg", ".jpg", ".png", ".gif":
		return file
	case ".mp4", ".webm":
		return strings.TrimSuffix(file, ext) + ".jpg"
	}
	return ""
}
