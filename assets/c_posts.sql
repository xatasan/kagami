-- -*- mode: sql; sql-product: sqlite -*-

CREATE TABLE IF NOT EXISTS posts (
	   postno INTEGER PRIMARY KEY ON CONFLICT IGNORE,
       replyto INTEGER ,
       sticky INTEGER,
	   closed INTEGER,
	   op INTEGER,
       time DATETIME,
	   name TEXT,
	   tripcode TEXT,
       id TEXT,
	   capcode TEXT,
	   country TEXT,
	   cname TEXT,
       subject TEXT,
	   comment TEXT,
       FOREIGN KEY (replyto)
	   		   REFERENCES posts(postno)
			   ON DELETE CASCADE);
