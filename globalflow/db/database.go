package db

import bolt "go.etcd.io/bbolt"

// Database is a wrapper around BoltDB that provides a higher-level API.
// It implements both a Redis-style KV store and a WAL.
type Database struct {
	// db is the underlying BoltDB database.
	db *bolt.DB
}

const BucketData = "DATA"
const BucketWAL = "WAL"

// NewDatabase creates or opens a database file at the given path.
func NewDatabase(path string) (*Database, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(BucketData))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte(BucketWAL))
		if err != nil {
			return err
		}

		return nil
	})

	return &Database{
		db: db,
	}, nil
}

func (db *Database) Close() error {
	return db.db.Close()
}
