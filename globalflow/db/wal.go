package db

type WALEntry struct {
	Index int64
}

//
//func (db *Database) WriteWAL(entry WALEntry) error {
//	return db.db.Update(func(tx *bolt.Tx) error {
//		b := tx.Bucket([]byte(BucketWAL))
//		if b == nil {
//			panic(fmt.Errorf("bucket %s not found", BucketWAL))
//		}
//
//		encoded, err := entry.Encode()
//		if err != nil {
//			return err
//		}
//
//		err = b.Put([]byte(entry.ID()), encoded)
//		if err != nil {
//			return err
//		}
//
//		return nil
//	})
