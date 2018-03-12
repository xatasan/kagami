-- -*- mode: sql; sql-product: sqlite -*-

CREATE VIRTUAL TABLE IF NOT EXISTS search
USING fts5(
	  content=posts,
      content_rowid=postno,
      comment,
      subject);
	  
CREATE TRIGGER IF NOT EXISTS post_ai
AFTER INSERT ON posts BEGIN
	  INSERT INTO search(
	  		 rowid,
			 comment,
			 subject)
	  VALUES (new.postno,
			  new.comment,
			  new.subject);
END;

CREATE TRIGGER IF NOT EXISTS tbl_ad
AFTER DELETE ON posts BEGIN
	  INSERT INTO search(
	  		 search,
			 rowid,
			 comment,
			 subject)
	  VALUES('delete',
			 old.postno,
			 old.comment,
			 old.subject);
END;

CREATE TRIGGER IF NOT EXISTS tbl_au
AFTER UPDATE ON posts BEGIN
	  INSERT INTO search(
	  		 search,
			 rowid,
			 comment,
			 subject)
	  VALUES('delete',
			 old.postno,
			 old.comment,
			 old.subject);

	  INSERT INTO search(
	  		 rowid,
			 comment,
			 subject)
      VALUES (new.postno,
	  		  new.comment,
			  new.subject);
END;
