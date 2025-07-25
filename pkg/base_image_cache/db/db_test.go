package db_test

// Tests are currently very simple and straighforward...

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/coffeemakingtoaster/whale-watcher/pkg/base_image_cache/db"
)

func GetTestDb() *sql.DB {
	// open sqlite in memory
	conn, err := db.LoadOrInitDB(":memory:")
	if err != nil {
		panic(err)
	}
	return conn
}
func GetRowCount(conn *sql.DB, table string) (int, error) {
	row := conn.QueryRow(fmt.Sprintf("SELECT Count(*) FROM %s", table))
	var count int
	row.Scan(&count)
	return count, row.Err()
}

func TestInitOfEmptyDb(t *testing.T) {
	conn := GetTestDb()
	defer conn.Close()

	// Check for tables named 'A' and 'B'
	rows, err := conn.Query(`
        SELECT name FROM sqlite_master 
        WHERE type='table' AND name IN ('image_package_lookup', 'image_digest_lookup')
    `)
	if err != nil {
		t.Fatalf("failed to query sqlite_master: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("failed to scan table name: %v", err)
		}
		tables = append(tables, name)
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("row iteration error: %v", err)
	}

	if len(tables) != 2 {
		t.Errorf("Encountered tables mismatch: Expected 'image_{package,digest}_lookup Got %v", tables)
	}
}

func TestAddImageDigest(t *testing.T) {
	conn := GetTestDb()
	defer conn.Close()

	count, err := GetRowCount(conn, "image_digest_lookup")
	if err != nil {
		t.Fatalf("Could not fetch row count for table! %s", err.Error())
	}
	if count != 0 {
		t.Fatalf("Initial count mismatch: Expected 0 Got %d", count)
	}
	err = db.AddImageDigest(conn, "abc", "mydigest1")
	if err != nil {
		t.Errorf("Could not insert image due to an error: %s", err.Error())
	}
	err = db.AddImageDigest(conn, "def", "mydigest2")
	if err != nil {
		t.Errorf("Could not insert image due to an error: %s", err.Error())
	}
	count, err = GetRowCount(conn, "image_digest_lookup")
	if err != nil {
		t.Errorf("Could not fetch row count for table! %s", err.Error())
	}
	if count != 2 {
		t.Errorf("Row count mismatch: Expected %d Got %d", 2, count)
	}
}

func TestAddImagePackage(t *testing.T) {
	conn := GetTestDb()
	defer conn.Close()

	count, err := GetRowCount(conn, "image_package_lookup")
	if err != nil {
		t.Fatalf("Could not fetch row count for table! %s", err.Error())
	}
	if count != 0 {
		t.Fatalf("Initial count mismatch: Expected 0 Got %d", count)
	}
	err = db.AddImagePackage(conn, "abc", "mydigest1", "v")
	if err != nil {
		t.Errorf("Could not insert image due to an error: %s", err.Error())
	}
	err = db.AddImagePackage(conn, "def", "mydigest2", "v")
	if err != nil {
		t.Errorf("Could not insert image due to an error: %s", err.Error())
	}
	count, err = GetRowCount(conn, "image_package_lookup")
	if err != nil {
		t.Errorf("Could not fetch row count for table! %s", err.Error())
	}
	if count != 2 {
		t.Errorf("Row count mismatch: Expected %d Got %d", 2, count)
	}

}

// TODO: extend in train
