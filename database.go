package main

import (
	"time"

	_ "github.com/mattn/go-sqlite3"

	"database/sql"
)

var (
	db         *sql.DB
	insertThr  *sql.Stmt
	insertFile *sql.Stmt
	insertLink *sql.Stmt
)

func setupdb(dbfile string) (err error) {
	db, err = sql.Open("sqlite3", "file:"+dbfile+"?cache=shared&mode=rwc")
	if err != nil {
		return
	}

	// files
	if _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS files (
            id INTEGER PRIMARY KEY ON CONFLICT IGNORE AUTOINCREMENT,
            filename TEXT,
            ofilename TEXT,
            filesize INTEGER, md5 TEXT,
            width INTEGER, height INTEGER,
            tfilename TEXT,
            twidth INTEGER, theight INTEGER,
            deleted INTEGER, spoiler INTEGER
        )`); err != nil {
		return
	}

	if insertFile, err = db.Prepare(`
        INSERT INTO files (filename, ofilename, filesize, md5, width, height, tfilename, twidth, theight, deleted, spoiler)
                   VALUES (?,         ?,         ?,        ?,   ?,     ?,      ?,         ?,      ?,       ?,       ?)
        `); err != nil {
		return
	}

	// posts
	if _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS posts (
            postno INTEGER PRIMARY KEY ON CONFLICT IGNORE,
            replyto INTEGER ,
            sticky INTEGER, closed INTEGER, op INTEGER,
            time DATETIME, name TEXT, tripcode TEXT,
            id TEXT, capcode TEXT, country TEXT, cname TEXT,
            subject TEXT, comment TEXT,
            FOREIGN KEY (replyto) REFERENCES posts(postno) ON DELETE CASCADE
        )`); err != nil {
		return
	}

	insertThr, err = db.Prepare(`
        INSERT INTO posts (postno, replyto, sticky, closed, op, time, name, tripcode, id, capcode, country, cname, subject, comment)
                   VALUES (?,      ?,       ?,      ?,      ?,  ?,    ?,    ?,        ?,  ?,       ?,       ?,     ?,       ?)
        `)

	// links (between files and posts) - WARNING: hacked together
	// `ord` = order
	if _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS links (
            post INTEGER, file INTEGER, ord INTEGER,
            PRIMARY KEY (post, file),
            FOREIGN KEY (post) REFERENCES posts(postno) ON DELETE CASCADE,
            FOREIGN KEY (file) REFERENCES files(id) ON DELETE CASCADE
        )`); err != nil {
		return
	}

	insertLink, err = db.Prepare(`
        INSERT INTO links (post, file, ord) VALUES (?, ?, ?)
        `)

	return
}

func write2db(save <-chan Thread) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	var t, f int
	start := time.Now()
	for T := range save {
		for _, P := range T {
			if _, err := tx.Stmt(insertThr).Exec(P.PostNumber,
				P.ReplyTo, P.Sticky, P.Closed, P.OP, P.Time,
				P.Name, P.Tripcode, P.Id, P.Capcode, P.Country,
				P.CountryName, P.Subject, string(P.Comment)); err != nil {
				return err
			}
			for i, F := range P.Files {
				res, err := tx.Stmt(insertFile).Exec(F.Filename,
					F.OrigFilename, F.FileSize, F.FileMD5,
					F.ImageWidth, F.ImageHeight, F.Thumbnail,
					F.ThumbnailWidth, F.ThumbnailHeight,
					F.FileDeleted, F.Spoiler)
				if err != nil {
					return err
				}

				id, err := res.LastInsertId()
				if err != nil {
					return err
				}

				if _, err := tx.Stmt(insertLink).Exec(P.PostNumber, id, i); err != nil {
					return err
				}
				f++
			}
			t++
		}

	}
	tx.Commit()
	verboseL("Inserted %d posts and %d files into the database in %v", t, f, time.Since(start))
	return nil
}
