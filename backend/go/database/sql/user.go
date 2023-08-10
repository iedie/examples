package sql

import "examples/database"

// CreateUser implements Storer, inserts a new User record into the database, the ID field will be generated as part of this process
func (db *DB) CreateUser(in *database.User) error {
	// Insert User into database, and update the User with returned ID
	return db.storage.QueryRow(`INSERT INTO users(first, last, email) VALUES ($1, $2, $3) RETURNING id`,
		in.First,
		in.Last,
		in.Email,
	).Scan(&in.ID)
}

// GetUserByID implements Storer, retrieves a User record by the ID field
func (db *DB) GetUserByID(id int64) (database.User, error) {
	// Prepare a User type
	var user database.User
	// Load the first record that is found (Since we're querying by ID, this should only ever return 1 anyway, and an error if not found)
	err := db.storage.QueryRow(
		`SELECT * FROM users WHERE id = $1`,
		id,
	).Scan(
		&user.ID,
		&user.First,
		&user.Last,
		&user.Email,
	)
	if err != nil {
		// Most common error is simply no rows, return an empty user, and our not found error
		return database.User{}, database.ErrNotFound
		// TODO (IME): There are other possible errors that can be returned, however for this demo, this is perfectly fine.
		// We may wish to handle other errors differently (For example if we can't reach our database, we may wish to return
		// a 500 or a 503 status to the frontend denote the user isn't doing anything wrong.)
	}
	// Return user
	return user, nil
}

// GetUserByEmail implements Storer, retrieves a User record by the Email field
func (db *DB) GetUserByEmail(email string) (database.User, error) {
	// Prepare a User type
	var user database.User
	// Load the first record that is found
	err := db.storage.QueryRow(
		`SELECT * FROM users WHERE email = $1`,
		email,
	).Scan(
		&user.ID,
		&user.First,
		&user.Last,
		&user.Email,
	)
	if err != nil {
		// Most common error is simply no rows, return an empty user, and our not found error
		return database.User{}, database.ErrNotFound
	}
	// Return user
	return user, nil
}

// DeleteUser implements Storer, deletes a User record from the database
func (db *DB) DeleteUser(id int64) error {
	// Delete User record from database, here we intentionally discard the returned output, as we only care if there was an error.
	_, err := db.storage.Exec(`DELETE FROM users WHERE id = $1`, id)
	return err
}
