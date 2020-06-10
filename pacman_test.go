package main

import (
	"bufio"
	"io/ioutil"
	"net"
	"testing"
	"time"
)

func expect(t *testing.T, received interface{}, expected interface{}) {
	if received != expected {
		t.Fatalf("\nExpected: %v\nReceived: %v\n", expected, received)
	}
}

func expectError(t *testing.T, err error) {
	if err == nil {
		t.Fatal("\nExpected an error; received nil.\n")
	}
}

func testMessage(t *testing.T, db *syncDatabase, msg string, expected string) {
	server, client := net.Pipe()
	reader := bufio.NewReader(server)

	go handleRequest(server, db, reader)

	_, err := client.Write([]byte(msg))

	go func() {
		time.Sleep(1 * time.Millisecond)
		server.Close()
	}()

	buf, err := ioutil.ReadAll(client)
	if err != nil {
		t.Fatal(err)
	}
	response := string(buf[:])
	expect(t, response, expected)
}

func TestMessages(t *testing.T) {
	OK := "OK\n"
	FAIL := "FAIL\n"
	ERROR := "ERROR\n"
	db := syncDatabase{db: make(database)}

	testMessage(t, &db, "INDEX|foo|bar\n", FAIL)
	testMessage(t, &db, "INDEX|bar|\n", OK)
	testMessage(t, &db, "INDEX|foo|bar\n", OK)
	testMessage(t, &db, "REMOVE|bar|\n", FAIL)
	testMessage(t, &db, "REMOVE|foo|\n", OK)
	testMessage(t, &db, "REMOVE|baz|\n", OK)
	testMessage(t, &db, "QUERY|baz|\n", FAIL)
	testMessage(t, &db, "QUERY|bar|\n", OK)
	testMessage(t, &db, "TWIDDLE|bar|\n", ERROR)
}

func TestIndex(t *testing.T) {
	db := make(database)
	foo := pkg("foo")
	bar := pkg("bar")

	// Indexing a package with 1+ dependencies that haven't been indexed fails.
	res := index(db, foo, []pkg{bar})
	expect(t, res, failRes)
	expect(t, len(db), 0)

	// Indexing a package with no dependencies succeeds.
	res = index(db, bar, []pkg{})
	expect(t, res, okRes)
	deps, exists := db[bar]
	expect(t, exists, true)
	expect(t, len(deps), 0)

	// Indexing a package with no un-indexed dependencies succeeds.
	res = index(db, foo, []pkg{bar})
	expect(t, res, okRes)
	deps, exists = db[foo]
	expect(t, exists, true)
	expect(t, len(deps), 1)
	_, exists = deps[bar]
	expect(t, exists, true)

	// Reindexing a package replaces its dependencies.
	res = index(db, foo, []pkg{})
	expect(t, res, okRes)
	deps, exists = db[foo]
	expect(t, exists, true)
	expect(t, len(deps), 0)
}

func TestRemove(t *testing.T) {
	foo := pkg("foo")
	bar := pkg("bar")
	db := database{}

	// Removing a nonexistent package succeeds.
	res := remove(db, foo)
	expect(t, res, okRes)
	expect(t, len(db), 0)

	db = database{
		bar: dependencies{},
		foo: dependencies{bar: null},
	}

	// Removing a package that is a dependency of another package fails.
	res = remove(db, bar)
	expect(t, res, failRes)
	_, exists := db[bar]
	expect(t, exists, true)

	// Removing a package that is not a dependency of any other packages succeeds.
	res = remove(db, foo)
	expect(t, res, okRes)
	_, exists = db[foo]
	expect(t, exists, false)
}

func TestQuery(t *testing.T) {
	foo := pkg("foo")
	db := database{}

	// Querying for a nonexistent package fails.
	res := query(db, foo)
	expect(t, res, failRes)
	expect(t, len(db), 0)

	db = database{
		foo: dependencies{},
	}

	// Querying for an existing package succeeds.
	res = query(db, foo)
	expect(t, res, okRes)
	expect(t, len(db), 1)
}

func TestParse(t *testing.T) {
	// Invalid message: incorrect number of segments.
	_, _, _, err := parse("WHATEVER|\n")
	expectError(t, err)

	// Invalid message: unknown command.
	_, _, _, err = parse("TWIDDLE|foo|\n")
	expectError(t, err)

	foo := pkg("foo")
	bar := pkg("bar")
	baz := pkg("baz")

	// Valid REMOVE message.
	command, candidate, deps, err := parse("REMOVE|foo|\n")
	expect(t, err, nil)
	expect(t, command, removeCmd)
	expect(t, candidate, foo)
	expect(t, len(deps), 0)

	// Valid QUERY message.
	command, candidate, deps, err = parse("QUERY|foo|\n")
	expect(t, err, nil)
	expect(t, command, queryCmd)
	expect(t, candidate, foo)
	expect(t, len(deps), 0)

	// Valid INDEX message with no dependencies.
	command, candidate, deps, err = parse("INDEX|foo|\n")
	expect(t, err, nil)
	expect(t, command, indexCmd)
	expect(t, candidate, foo)
	expect(t, len(deps), 0)

	// Valid INDEX message with dependencies.
	command, candidate, deps, err = parse("INDEX|foo|bar,baz\n")
	expect(t, err, nil)
	expect(t, command, indexCmd)
	expect(t, candidate, foo)
	expect(t, len(deps), 2)
	expect(t, deps[0], bar)
	expect(t, deps[1], baz)

	// Valid INDEX message with extraneous commas.
	command, candidate, deps, err = parse("INDEX|foo|,,\n")
	expect(t, err, nil)
	expect(t, command, indexCmd)
	expect(t, candidate, foo)
	expect(t, len(deps), 0)

	// TODO: do we want to impose restrictions on package names?
	// Valid INDEX message with non-standard characters in package name.
	command, candidate, deps, err = parse("INDEX|\\|,,\n")
	expect(t, err, nil)
	expect(t, command, indexCmd)
	expect(t, candidate, pkg("\\"))
	expect(t, len(deps), 0)

	// Invalid INDEX message: missing package name.
	command, candidate, deps, err = parse("INDEX||bar,baz\n")
	expectError(t, err)
}
