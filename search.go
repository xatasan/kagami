package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

var post_search *sql.Stmt

func init() {
	var err error
	post_search, err = db.Prepare(`
	SELECT posts.postno,
               posts.replyto,
               posts.time,
               posts.name,
               posts.tripcode,
               posts.id,
	       posts.capcode,
               posts.country,
               posts.cname,
               posts.subject,
               highlight(search, 0, "<strong class=search>", "</strong>")
	FROM search
        LEFT JOIN posts ON posts.postno = search.rowid
        WHERE search MATCH ?
        ORDER BY search.rank
	LIMIT ? OFFSET ?`)
	if err != nil {
		log.Fatal(err)
	}
}

func search(rw http.ResponseWriter, req *http.Request) {
	var ( // get data from query
		squery   = req.URL.Query().Get("q") // the query
		spage_r  = req.URL.Query().Get("p") // page, raw
		slimit_r = req.URL.Query().Get("l") // limit
	)

	type Resp struct {
		Msg   string
		Items []Post
	}

	if squery == "" {
		rw.WriteHeader(http.StatusBadRequest)
		resp := Resp{Msg: "no query"}
		if err := json.NewEncoder(rw).Encode(resp); err != nil {
			log.Fatal(err)
		}
		return
	}

	if spage_r == "" {
		spage_r = "1"
	}

	spage, err := strconv.Atoi(spage_r)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		resp := struct{ msg string }{msg: err.Error()}
		if err := json.NewEncoder(rw).Encode(resp); err != nil {
			log.Fatal(err)
		}
		return
	}

	var slimit int // don't read in any number
	switch slimit_r {
	case "100":
		slimit = 100
	case "50":
		slimit = 50
	case "25":
		fallthrough
	default:
		slimit = 25
	}

	rows, err := post_search.Query(squery, slimit, (spage-1)*slimit)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		resp := struct{ msg string }{msg: err.Error()}
		if err := json.NewEncoder(rw).Encode(resp); err != nil {
			log.Fatal(err)
		}
		return
	}
	defer rows.Close()

	data := make([]map[string]interface{}, 0, slimit)
	for rows.Next() {
		var (
			postno, respto          int64
			ptime                   time.Time
			name, trip, id          string
			capcode, country, cname string
			subject, comment        string
		)
		rows.Scan(&postno, &respto, &ptime, &name,
			&trip, &id, &capcode, &country,
			&cname, &subject, &comment)
		data = append(data, map[string]interface{}{
			"postno":  postno,
			"respto":  respto,
			"time":    ptime.Format("2006/01/_2 (Mon) 15:04:05"),
			"name":    name,
			"trip":    trip,
			"id":      id,
			"capcode": capcode,
			"country": country,
			"cname":   cname,
			"subject": subject,
			"comment": comment,
		})
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	enc := json.NewEncoder(rw)
	enc.SetEscapeHTML(true)
	enc.Encode(data)
}
