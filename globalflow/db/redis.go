package db

import (
	"fmt"
	bolt "go.etcd.io/bbolt"
	"time"
)

// Get gets a value from the database.
func (db *Database) Get(now Time, key string) (string, error) {
	data := Data{}

	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketData))
		if b == nil {
			panic(fmt.Errorf("bucket %s not found", BucketData))
		}

		v := b.Get([]byte(key))

		if v == nil {
			return &ErrorNotFound{Key: key}
		}

		err := data.Decode(v)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	if data.Type != DataTypeString {
		return "", fmt.Errorf("wrong type")
	}

	if data.ExpiresAt != 0 && data.ExpiresAt < now {
		// TODO: expire key
		return "", &ErrorNotFound{Key: key}
	}

	return data.StringValue, nil
}

// Set sets a value in the database.
func (db *Database) Set(now time.Time, key string, value string, expiresAt Time) error {
	data := Data{
		Type:        DataTypeString,
		StringValue: value,
		ExpiresAt:   expiresAt,
	}

	err := db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketData))
		if b == nil {
			panic(fmt.Errorf("bucket %s not found", BucketData))
		}

		encoded, err := data.Encode()
		if err != nil {
			return err
		}

		return b.Put([]byte(key), encoded)
	})
	if err != nil {
		return err
	}

	return nil
}

type ErrorNotFound struct {
	Key string
}

func (e ErrorNotFound) Error() string {
	return fmt.Sprintf("key %s not found", e.Key)
}

func IsErrorNotFound(err error) bool {
	_, ok := err.(*ErrorNotFound)
	return ok
}
