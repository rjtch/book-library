package schema

import (
	"github.com/dimiro1/darwin"
	"github.com/jmoiron/sqlx"
)

// Migrate attempts to bring the schema for db up to date with the migrations
// defined in this package.
func Migrate(db *sqlx.DB) error {
	driver := darwin.NewGenericDriver(db.DB, darwin.PostgresDialect{})

	d := darwin.New(driver, migrations, nil)

	return d.Migrate()
}

// migrations contains the queries needed to construct the database schema.
// Entries should never be removed from this slice once they have been ran in
// production.
//
// Using constants in a .go file is an easy way to ensure the queries are part
// of the compiled executable and avoids pathing issues with the working
// directory. It has the downside that it lacks syntax highlighting and may be
// harder to read for some cases compared to using .sql files. You may also
// consider a combined approach using a tool like packr or go-bindata.
var migrations = []darwin.Migration{
	{
		Version:     1,
		Description: "Add books",
		Script: `
CREATE TABLE books (
	book_id      UUID,
	title        TEXT,
	isbn         TEXT,
    category     TEXT,
	authors      TEXT,
	description  TEXT,
	quantity	 INT,
	date_created TIMESTAMP,
	date_updated TIMESTAMP,

	PRIMARY KEY (book_id)
);`,
	}, {
		Version:     2,
		Description: "Add users",
		Script: `
CREATE TABLE users (
	user_id       UUID,
	name          TEXT,
	email         TEXT UNIQUE,
	roles         TEXT[],
	password_hash TEXT,
	
	date_created TIMESTAMP,
	date_updated TIMESTAMP,

	PRIMARY KEY (user_id)
);`,
	}, {
		Version:     3,
		Description: "Add category",
		Script: `
CREATE TABLE categories (
	category_id   UUID,
	name          TEXT,
	books_in      INT,
	books_out     INT,

	date_created TIMESTAMP,
	date_updated TIMESTAMP,

	PRIMARY KEY (category_id)
);`,
	}, {
		Version:     4,
		Description: "Add loan",
		Script: `
CREATE TABLE loans (
	loan_id     UUID,
	title       TEXT,
	isbn        TEXT,
    quantity    INT,
	book_id		UUID,
	loan_date TIMESTAMP,
	date_return TIMESTAMP,
	user_id     UUID,

	PRIMARY KEY (loan_id),

	FOREIGN KEY (book_id) REFERENCES books(book_id) ON DELETE CASCADE
);

CREATE TABLE sessions (
	token TEXT PRIMARY KEY,
	data BYTEA NOT NULL,
	expiry TIMESTAMPTZ NOT NULL
);

    CREATE INDEX sessions_expiry_idx ON sessions (expiry);
`,
	},
}
