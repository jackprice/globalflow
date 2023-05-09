package db

import (
	"os"
	"testing"
	"time"
)

func TestDatabase_Get(t *testing.T) {
	p, err := os.CreateTemp(os.TempDir(), "bolt")
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Close(); err != nil {
		t.Fatal(err)
	}

	db, err := NewDatabase(p.Name())
	if err != nil {
		t.Fatal(err)
	}

	err = db.Set(time.Now(), "foo", "bar", 0)
	if err != nil {
		t.Fatal(err)
	}

	value, err := db.Get(1, "foo")
	if err != nil {
		t.Fatal(err)
	}

	if value != "bar" {
		t.Fatalf("expected %s, got %s", "bar", value)
	}
}
