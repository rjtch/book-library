package schema

import "github.com/jmoiron/sqlx"

// Seed runs the set of seed-data queries against db. The queries are ran in a
// transaction and rolled back if any fail.
func Seed(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(seeds); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

// seeds is a string constant containing all of the queries needed to get the
// db seeded to a useful state for development.
//
// Note that database servers besides PostgreSQL may not support running
// multiple queries as part of the same execution so this single large constant
// may need to be broken up.
const seeds = `
INSERT INTO users (user_id, name, email, roles, password_hash, date_created, date_updated) VALUES
	('5cf37266-3473-4006-984f-9325122678b7', 'Admin Gopher', 'admin@example.com', '{ADMIN,USER}', '$2a$10$1ggfMVZV6Js0ybvJufLRUOWHS5f6KneuP0XwwHpJ8L8ipdry9f2/a', '2019-03-24 00:00:00', '2019-03-24 00:00:00'),
	('45b5fbd3-755f-4379-8f07-a58d4a30fa2f', 'User Gopher', 'users@example.com', '{USER}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', '2019-03-24 00:00:00', '2019-03-24 00:00:00')
	ON CONFLICT DO NOTHING;

INSERT INTO books (book_id, title, isbn, category, authors, description ,quantity ,date_created, date_updated) VALUES
	('a2b0639f-2cc6-44b8-b97b-15d69dbb511e', 'Comic Books', '5we0K', 'bwl' ,'John Lenon','learn the best way 1' ,'1' ,'2019-01-01 00:00:01.000001+00', 
	'2019-01-01 00:00:01.000001+00'),
	('a2b0639f-2cc6-44b8-b97b-15d69dbb511e', 'angular', 'fsn22', 'computer-science' ,'Bob Andre', 'learn the best way 2' 
	,'1' ,'2019-01-01 00:00:01.000001+00', '2019-01-01 00:00:01.000001+00'),
	('4ef52818-7f53-47ad-ae4a-b271b63f0a96', 'go programming language', 'bsn23', 'computer-science' ,'Google', 'learn the best way 3' ,'3' ,'2019-01-01 00:00:01.000001+00', '2019-01-01 00:00:01.000001+00')
	ON CONFLICT DO NOTHING;
	
INSERT INTO categories (category_id, name, books_in, books_out, date_created, date_updated) VALUES 
	('fe30348e-50db-11ea-8d77-2e728ce88125', 'computer-science', '2', '0', '2019-01-01 00:00:01.000001+00', '2019-01-01 00:00:01.000001+00') 
	ON CONFLICT DO NOTHING;

INSERT INTO loans (loan_id, title, isbn, quantity, book_id, loan_date, date_return, user_id) 
	VALUES ('10b57268-50dc-11ea-8d77-2e728ce88125', 'go programming language', 'bsn23', '1', '4ef52818-7f53-47ad-ae4a-b271b63f0a96', '2019-01-01 00:00:01.000001+00', '2019-01-01 00:00:01.000001+00',
	'5cf37266-3473-4006-984f-9325122678b7') ,

	('e85c41c6-a2ab-11ea-bb37-0242ac130002', 'angular', 'fsn22', '1', 'a2b0639f-2cc6-44b8-b97b-15d69dbb511e', '2019-01-01 00:00:01.000001+00', '2019-01-01 00:00:01.000001+00',
	'45b5fbd3-755f-4379-8f07-a58d4a30fa2f') 
	ON CONFLICT DO NOTHING;
`
