-- -*- mode: sql; sql-product: sqlite -*-

CREATE TABLE IF NOT EXISTS links (
            post INTEGER,
			file INTEGER,
			ord INTEGER,
            PRIMARY KEY (post, file),
            FOREIGN KEY (post)
					REFERENCES posts(postno)
					ON DELETE CASCADE,
            FOREIGN KEY (file)
					REFERENCES files(id)
					ON DELETE CASCADE)
