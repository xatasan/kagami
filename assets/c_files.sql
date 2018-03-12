-- -*- mode: sql; sql-product: sqlite -*-

CREATE TABLE IF NOT EXISTS files (
	   id INTEGER PRIMARY KEY ON CONFLICT IGNORE AUTOINCREMENT,
	   filename TEXT,
	   ofilename TEXT,
	   filesize INTEGER,
	   md5 TEXT,
	   width INTEGER,
	   height INTEGER,
	   tfilename TEXT,
	   twidth INTEGER,
	   theight INTEGER,
	   deleted INTEGER,
	   spoiler INTEGER);
