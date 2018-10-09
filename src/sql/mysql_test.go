package sql_test

import (
	goSQL "database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/cndjp/qicoo-api/src/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/lestrrat-go/test-mysqld"
)

var testMysqld *mysqltest.TestMysqld
var mySQLDataSrc string

// 元のhandlerから参照すると循環参照になってしまうと言う悲劇から適当に作った。
// time.Time型の扱いが面倒なので、時間もstringにしてしまった。
// よほど暇なら直すかもしれない。
type mock struct {
	ID        string `json:"id" db:"id"`
	Object    string `json:"object" db:"object"`
	Username  string `json:"username" db:"username"`
	EventID   string `json:"event_id" db:"event_id"`
	ProgramID string `json:"program_id" db:"program_id"`
	Comment   string `json:"comment" db:"comment"`
	CreatedAt string `json:"created_at" db:"created_at"`
	UpdatedAt string `json:"updated_at" db:"updated_at"`
	Like      int    `json:"like" db:"like_count"`
}

func isTravisEnv() bool {
	if os.Getenv("TRAVIS") == "true" {
		fmt.Println("travis env is true")
		return true
	}
	fmt.Println("travis env is false")
	return false
}

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {

	if isTravisEnv() {
		mySQLDataSrc = "root@tcp(localhost:3306)/test"
	} else {
		mysqld, err := mysqltest.NewMysqld(nil)
		if err != nil {
			log.Fatal("runTests: failed launch mysql server:", err)
		}
		defer mysqld.Stop()

		testMysqld = mysqld
		fmt.Println(testMysqld.Datasource("test", "", "", 0))
		mySQLDataSrc = testMysqld.Datasource("test", "", "", 0)
	}

	return m.Run()
}

func truncateTables() {
	fmt.Println("debug AAAA")
	db, err := goSQL.Open("mysql", mySQLDataSrc)
	if err != nil {
		log.Fatal("db connection error:", err)
	}
	defer db.Close()

	fmt.Println("debug BBBB")
	_, err = db.Exec("TRUNCATE test.mock")
	if err != nil {
		log.Fatal("truncate table error:", err)
	}
}

func TestMappingDBandTable(t *testing.T) {
	fmt.Println("debug 1111")
	defer truncateTables()

	fmt.Println("debug 2222")
	db, err := goSQL.Open("mysql", mySQLDataSrc)
	if err != nil {
		t.Fatal("db connection error:", err)
	}
	defer db.Close()
	// if isTravisEnv() {
	// 	databaseRow, err := db.Query(`CREATE DATABASE test`)
	// 	if err != nil {
	// 		t.Fatal("create databases error:", err)
	// 	}
	// 	defer databaseRow.Close()
	// }

	fmt.Println("debug 3333")
	tableRow, err := db.Query(`
		CREATE TABLE test.mock (
		id         INT          NOT NULL,
		object     VARCHAR(255) NOT NULL,
		username   VARCHAR(255) NOT NULL,
		event_id   VARCHAR(255) NOT NULL,
		program_id VARCHAR(255) NOT NULL,
		comment    TEXT         NOT NULL,
		created_at VARCHAR(255) NOT NULL,
		updated_at VARCHAR(255) NOT NULL,
		like_count INT          NOT NULL)`)
	if err != nil {
		t.Fatal("create tables error:", err)
	}
	defer tableRow.Close()

	fmt.Println("debug 4444")
	insertRow, err := db.Query("INSERT INTO test.mock VALUES (1, 'question', 'anonymous', '1', '1', 'I am mock', 'now', 'mock', 100000)")
	if err != nil {
		t.Fatal("insert row error:", err)
	}
	defer insertRow.Close()

	fmt.Println("debug 5555")
	m := sql.MappingDBandTable(db)
	m.AddTableWithName(mock{}, "mock")

	fmt.Println("debug 6666")
	var mks []mock
	_, err = m.Select(&mks, "SELECT * FROM test.mock WHERE id = 1")
	if err != nil {
		t.Fatal("select error:", err)
	}

	fmt.Println("debug 7777")
	var mockComment string
	var expectedComment = "I am mock"
	for _, mk := range mks {
		js, err := json.Marshal(mk)
		if err != nil {
			t.Fatal("json marshal error:", err)
		}
		t.Log("mock rows:", string(js))
		mockComment = mk.Comment
	}

	fmt.Println("debug 8888")
	if !reflect.DeepEqual(expectedComment, mockComment) {
		t.Errorf("expected %q to eq %q", expectedComment, mockComment)
	}
}
