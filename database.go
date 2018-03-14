package kagami

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	k "github.com/xatasan/kagami/types"
)

var (
	db *sql.DB

	insertThr  *sql.Stmt
	insertFile *sql.Stmt
	insertLink *sql.Stmt
	postSearch *sql.Stmt

	toSave = make(chan<- k.Thread)
)

func SetupDatabase(dbfile string) (err error) {
	conn := "file:" + dbfile + "?cache=shared&mode=rwc"
	db, err = sql.Open("sqlite3", conn)
	if err != nil {
		return
	}

	for _, exec := range []string{
		"x_setup.sql",
		"c_files.sql",
		"c_posts.sql",
		"c_links.sql",
		// "c_search.sql",
	} {
		_, err = db.Exec(string(MustAsset(exec)))
		if err != nil {
			return
		}
	}

	for stmt, vp := range map[string]**sql.Stmt{
		"i_files.sql": &insertFile,
		"i_link.sql":  &insertLink,
		"i_posts.sql": &insertThr,
		// "s_search.sql": &postSearch,
	} {
		*vp, err = db.Prepare(string(MustAsset(stmt)))
		if err != nil {
			return
		}
	}

	return nil
}

func saveBoard(brd k.Board) error {
	threads := make(chan k.Thread)
	wg, err := brd.Threads(threads)
	if err != nil {
		return err
	}
	go func() {
		wg.Wait()
		close(threads)
	}()
	return save(threads)
}

func saveThread(thr k.Thread) error {
	threads := make(chan k.Thread)
	threads <- thr
	close(threads)
	return save(threads)
}

func save(save <-chan k.Thread) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return
	}
	for thread := range save {
		for _, post := range thread {
			err = preparePost(post, tx)
			if err != nil {
				return
			}
		}
	}
	return tx.Commit()
}

func preparePost(p *k.Post, tx *sql.Tx) error {
	var resp int
	if p.ReplyTo != nil {
		resp = p.ReplyTo.PostNumber
	}
	_, err := tx.Stmt(insertThr).Exec(
		p.PostNumber, resp, p.Sticky, p.Closed,
		p.OP, p.Time, p.Name, p.Tripcode, p.Id,
		p.Capcode, p.Flag.Icon, p.Flag.Name,
		p.Subject, string(p.Comment))
	if err != nil {
		return err
	}

	for order, f := range p.Files {
		res, err := tx.Stmt(insertFile).Exec(
			f.Filename, f.OrigFilename, f.FileSize,
			f.FileMD5, f.Image.X, f.Image.Y,
			f.ThumbnailName, f.Thumbnail.X,
			f.Thumbnail.Y, f.FileDeleted, f.Spoiler)
		if err != nil {
			return err
		}

		id, err := res.LastInsertId()
		if err != nil {
			return err
		}

		_, err = tx.Stmt(insertLink).Exec(
			p.PostNumber, id, order)
		if err != nil {
			return err
		}
	}
	return nil
}
