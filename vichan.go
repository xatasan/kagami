package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"regexp"
	"time"
)

var vichan_link_re *regexp.Regexp

func init() {
	vichan_link_re = regexp.MustCompile(`<a onclick="highlightReply\('(\d+)'(?:, event)?\);"\s+href="/(\w+)/res/(\d+).html#\d+">`)
}

type e_vichan struct {
	Engine
	host string
}

func (e e_vichan) getName() string {
	return e.getHost()
}

func (e e_vichan) getHost() string {
	return e.host
}

func (e e_vichan) getFile(f string) string {
	return "http://" + e.getHost() + "/src/" + f
}

func (e e_vichan) getTmb(f string) string {
	var tmb string
	switch path.Ext(f) {
	case ".jpeg", ".jpg", ".png":
		tmb = f
	default:
		return "https://" + e.getHost() + "/static/file.jpg"
	}

	return "http://" + e.getHost() + "/src/" + tmb + "s.jpg"
}

func (e e_vichan) getStatic(_, f string) string {
	return fmt.Sprintf("https://%s/static/%s", e.getHost(), f)
}

func (e e_vichan) genThread(board, no string) (Thread, error) {
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
	req := fmt.Sprintf("https://%s/%s/res/%s.json", e.getHost(), board, no)
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

	var thread Thread
	first := true
	for _, post := range data["posts"] {
		var files []File
		if post.Filename != "" {
			files = append(files, File{
				Filename:        post.Tim + post.Ext,
				OrigFilename:    post.Filename,
				FileSize:        int(post.Fsize),
				FileMD5:         post.Md5,
				ImageWidth:      int(post.Width),
				ImageHeight:     int(post.Height),
				Thumbnail:       post.Tim + "s.jpg",
				ThumbnailWidth:  int(post.TmbWidth),
				ThumbnailHeight: int(post.TmbHeight),
				FileDeleted:     post.FileDeleted != 0,
				Spoiler:         post.Spoiler != 0,
			})
		}

		com := vichan_link_re.ReplaceAllString(post.Com,
			`<a class="r" href="./res/$3.html#$1">`)
		thread = append(thread, Post{
			PostNumber:  int(post.No),
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

	return thread, nil
}

func (e e_vichan) getThreads(b string, tc chan<- struct{ n, l float64 }) error {
	req := fmt.Sprintf("https://%s/%s/threads.json", e.getHost(), b)
	resp, err := http.Get(req)
	if err != nil {
		return err
	} else if resp.StatusCode >= 400 {
		return fmt.Errorf("Failed to get threads: %d", resp.StatusCode)
	}

	var data []struct {
		Threads []map[string]float64 `json:"threads"`
		Page    float64              `json:"page"`
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&data)
	if err != nil {
		return err
	}

	for _, page := range data {
		for _, thread := range page.Threads {
			tc <- struct{ n, l float64 }{thread["no"], thread["last_modified"]}
		}
	}
	close(tc)
	return nil
}
