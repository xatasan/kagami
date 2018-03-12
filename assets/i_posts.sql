-- -*- mode: sql; sql-product: sqlite -*-

INSERT INTO posts (
	   postno,
	   replyto,
	   sticky,
	   closed,
	   op,
	   time,
	   name,
	   tripcode,
	   id,
	   capcode,
	   country,
	   cname,
	   subject,
	   comment)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
