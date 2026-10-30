package migrate

import (
	"testing"
	"testing/fstest"
)

func TestReadMigrationsSortsAndIgnoresUnrelatedFiles(t *testing.T) {
	files := fstest.MapFS{
		"002_add_users.sql":     {Data: []byte("SELECT 2;")},
		"001_create_roles.sql":  {Data: []byte("SELECT 1;")},
		"README.md":             {Data: []byte("ignored")},
		"nested/003_other.sql":  {Data: []byte("ignored")},
		"004_invalid-name!.sql": {Data: []byte("ignored")},
	}

	got, err := readMigrations(files)
	if err != nil {
		t.Fatalf("readMigrations() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d migrations, want 2", len(got))
	}
	if got[0].version != 1 || got[0].name != "001_create_roles.sql" {
		t.Errorf("first migration = %+v, want version 1", got[0])
	}
	if got[1].version != 2 || got[1].name != "002_add_users.sql" {
		t.Errorf("second migration = %+v, want version 2", got[1])
	}
}

func TestReadMigrationsRejectsDuplicateVersions(t *testing.T) {
	files := fstest.MapFS{
		"001_create_roles.sql": {Data: []byte("SELECT 1;")},
		"001_add_users.sql":    {Data: []byte("SELECT 2;")},
	}

	if _, err := readMigrations(files); err == nil {
		t.Fatal("readMigrations() succeeded, want duplicate-version error")
	}
}
