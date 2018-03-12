-- -*- mode: sql; sql-product: sqlite -*-

SELECT
	posts.postno,
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
LIMIT ? OFFSET ?
