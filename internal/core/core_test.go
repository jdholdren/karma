package core

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"gotest.tools/assert"

	coredb "github.com/jdholdren/karma/internal/core/db"
	"github.com/jdholdren/karma/internal/core/models"
)

var (
	sqlxDB *sqlx.DB
	coreDB coredb.DB
	cr     Core
)

func removeDB() {
	os.Remove("../../test.sqlite")
	os.Remove("../../test.sqlite-shm")
	os.Remove("../../test.sqlite-wal")
}

func TestMain(t *testing.M) {
	u, err := url.Parse("../../test.sqlite")
	if err != nil {
		fmt.Println("error parsing url: ", err)
		os.Exit(1)
	}

	q := u.Query()
	q.Add("_journal", "WAL")
	u.RawQuery = q.Encode()

	sqlxDB, err = sqlx.Open("sqlite3", u.String())
	if err != nil {
		fmt.Println("error opening test db: ", err)
		removeDB()
		os.Exit(1)
	}

	// Perform migrations
	ups, err := ioutil.ReadDir("../../migrate")
	if err != nil {
		fmt.Println("error reading migrate dir: ", err)
		removeDB()
		os.Exit(1)
	}

	for _, up := range ups {
		if up.IsDir() {
			continue
		}

		if !strings.HasSuffix(up.Name(), "sql") {
			continue
		}

		upBytes, err := ioutil.ReadFile(filepath.Join("../../migrate", up.Name()))
		if err != nil {
			fmt.Println("error reading migration file: ", err)
			removeDB()
			os.Exit(1)
		}

		_, err = sqlxDB.Exec(string(upBytes))
		if err != nil {
			fmt.Println("error executing migration: ", err)
			removeDB()
			os.Exit(1)
		}
	}

	coreDB = coredb.New(sqlxDB)
	cr = New(coreDB)

	code := t.Run()

	removeDB()
	os.Exit(code)
}

func TestIncrementKarma(t *testing.T) {
	ctx := context.Background()

	_, err := cr.AddKarma(ctx, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	_, err = cr.AddKarma(ctx, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	got, err := cr.GetKarma(ctx, "user-1")
	if err != nil {
		t.Fatalf("unexpected error getting karma: %s", err)
	}

	want := models.KarmaCount{
		UserID: "user-1",
		Count:  2,
	}
	assert.Equal(t, want, got)
}
