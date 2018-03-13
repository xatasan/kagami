package infchan

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	k "github.com/xatasan/kagami/types"
	"github.com/xatasan/kagami/vichan"
)

const (
	host     = "http://8ch.net"
	asset    = "https://8ch.net/static/assets/%s/%s"
	file     = "https://media.8ch.net/file_store/"
	file_png = "https://8ch.net/static/file.png"
	thumb    = "https://media.8ch.net/file_store/thumb/%s"
	thread   = "https://8ch.net/%s/res/%s.json"
)

type engine interface{}

func (e engine) Name() string {
	return "8chan"
}

func (e engine) Host() string {
	return host
}

func (e engine) Board(board string) (*k.Board, error) {
	return board{board: board}, nil
}

type board struct{ name string }

func (b board) Name() string {
	return b.name
}

func (b board) Threads(ch chan<- *k.Thread) (*sync.WaitGroup, error) {
	threads, err := vichan.ThreadList(host, b.name)
	if err != nil {
		return nil, err
	}
	var wg *sync.WaitGroup
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

func (b board) Thread(name string) (*k.Thread, error) {
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

	// request /res/%d.json
	req := fmt.Sprintf(thread, board, no)
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

	qbi := make(map[int][]int) // quotes by ids
	var thread Thread
	first := true
	for _, post := range data["posts"] {
		var files []File
		if post.Filename != "" {
			files = append(files, File{
				Filename:        post.Tim + post.Ext,
				OrigFilename:    post.Filename + post.Ext,
				FileSize:        int(post.Fsize),
				FileMD5:         post.Md5,
				ImageWidth:      int(post.Width),
				ImageHeight:     int(post.Height),
				Thumbnail:       genThumbnail(post.Tim + post.Ext),
				ThumbnailWidth:  int(post.TmbWidth),
				ThumbnailHeight: int(post.TmbHeight),
				FileDeleted:     post.FileDeleted != 0 || post.Ext == "deleted",
				Spoiler:         post.Spoiler != 0,
			})
		}
		for _, file := range post.ExtraFiles {
			files = append(files, File{
				Filename:        file.Tim + file.Ext,
				OrigFilename:    file.Filename + file.Ext,
				FileSize:        int(file.Fsize),
				FileMD5:         file.Md5,
				ImageWidth:      int(file.Width),
				ImageHeight:     int(file.Height),
				Thumbnail:       genThumbnail(file.Tim + file.Ext),
				ThumbnailWidth:  int(file.TmbWidth),
				ThumbnailHeight: int(file.TmbHeight),
				FileDeleted:     file.FileDeleted != 0 || post.Ext == "deleted",
				Spoiler:         file.Spoiler != 0,
			})
		}

		id := int(post.No)
		quotes := vichan_link_re.FindAllStringSubmatch(post.Com, -1)
		for _, q := range quotes {
			qid, err := strconv.Atoi(q[1]) // get quoted id
			if err != nil {
				return nil, err
			}
			qbi[qid] = append(qbi[qid], id)
		}

		com := vichan_link_re.ReplaceAllString(post.Com,
			`<a class="r" href="./$3.html#$1">`)
		thread = append(thread, Post{
			PostNumber:  id,
			ReplyTo:     int(post.Resto),
			Sticky:      post.Sticky != 0,
			Closed:      post.Closed != 0,
			OP:          first,
			Time:        time.Unix(int64(post.Time), 0),
			Name:        post.Name,
			Tripcode:    post.Trip,
			Id:          post.Id,
			Capcode:     post.Capcode,
			Country:     post.Country,
			CountryName: post.CountryName,
			Subject:     post.Sub,
			Comment:     template.HTML(com),
			Files:       files,
			Images:      int(post.Images),
		})
		first = false
	}

	for i, t := range thread {
		thread[i].Quoted = qbi[t.PostNumber]
	}

	return thread, nil
}

func (_ board) GetFileUri(f File) *url.URL {
	path := file
	path += f.Filename
	u, _ := url.Parse(path)
	return u
}

func (_ board) genThumbnail(file string) string {
	ext := path.Ext(file)
	switch ext {
	case ".jpeg", ".jpg", ".png", ".gif":
		return file
	case ".mp4", ".webm":
		return strings.TrimSuffix(file, ext) + ".jpg"
	}
	return ""
}

func (_ board) GetTmbUri(f File) *url.URL {
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

func (_ board) GetStaticUri(f File) *url.URL {
	var path string
	if f.OrigFilename == "file.png" {
		path = file_png
	} else {
		path = fmt.Sprintf(asset, board, file)
	}
	u, _ := url.Parse(path)
	return u
}

func Engine(host string) Engine {
	return board{host: host}
}
