package sql

import (
	"examples/database"
	"time"
)

// SaveSession implements Storer, inserts a Session into the database, and also updates the ID field with the ID that is returned from insertion.
func (db *DB) SaveSession(in *database.Session) error {
	// Insert session into database, and update the session with returned ID
	return db.storage.QueryRow(`INSERT INTO sessions(encryptedcreds, expiration, endoflife) VALUES ($1, $2, $3) RETURNING id`,
		in.EncryptedCreds,
		in.Expires,
		in.EndOfLife,
	).Scan(&in.ID)
}

// LoadSession implements Storer, retrieves a Session from the database by ID.
func (db *DB) LoadSession(id int64) (database.Session, error) {
	// Prepare session
	var session database.Session
	// Load the first record that is found (Since we're querying by ID, this should only ever return 1 anyway, and an error if not found)
	err := db.storage.QueryRow(
		`SELECT * FROM sessions WHERE id = $1`,
		id,
	).Scan(
		&session.ID,
		&session.EncryptedCreds,
		&session.Expires,
		&session.EndOfLife,
	)
	if err != nil {
		// Most common error is simply no rows, return an empty session, and our not found error
		return database.Session{}, database.ErrNotFound
		// TODO (IME): There are other possible errors that can be returned, however for this demo, this is perfectly fine.
		// We may wish to handle other errors differently (For example if we can't reach our database, we may wish to return
		// a 500 or a 503 status to the frontend denote the user isn't doing anything wrong.)
	}
	// Return session
	return session, nil
}

// LogoutSession implements Storer, deletes a Session from the database by ID.
func (db *DB) LogoutSession(id int64) error {
	// Delete session record from database, here we intentionally discard the returned output, as we only care if there was an error.
	_, err := db.storage.Exec(`DELETE FROM sessions WHERE id = $1`, id)
	return err
}

// ExtendSession implements Storer, updates a Session record to have a new expiration. We intentially discard returned output as we are
// only concerned if there was an error.
func (db *DB) ExtendSession(id int64, lifespan time.Duration) error {
	// Refresh the expiration
	_, err := db.storage.Exec(
		`UPDATE sessions SET expiration = $1 WHERE id = $2`,
		time.Now().Add(lifespan),
		id,
	)
	return err
}

// ClearExpiredSessions implements Storer, deletes any Session records that are expired. Ideally this would be called from a background task
// at regular intervals to keep the database free of useless records.
func (db *DB) ClearExpiredSessions() (int, error) {
	// Delete expired session records from database
	result, err := db.storage.Exec(`DELETE FROM sessions WHERE expiration < current_timestamp OR endoflife < current_timestamp`)
	if err != nil {
		return 0, nil
	}
	ra, err := result.RowsAffected()
	return int(ra), err
}
