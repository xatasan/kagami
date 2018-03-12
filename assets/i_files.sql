-- -*- mode: sql; sql-product: sqlite -*-

INSERT INTO files (
	   filename,
	   ofilename,
	   filesize,
	   md5,
	   width,
	   height,
	   tfilename,
	   twidth,
	   theight,
	   deleted,
	   spoiler)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
