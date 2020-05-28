package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

var null struct{}

type command string
type result string

type pkg string

type dependencies map[pkg]struct{}

//  index of packages
//  Each key is a registered package, and each value is that package's list of dependencies.
type database map[pkg]dependencies

type syncDatabase struct {
	sync.RWMutex
	db database
}

const (
	address  = "localhost:8080"
	protocol = "tcp"

	index_cmd  command = "INDEX"
	query_cmd  command = "QUERY"
	remove_cmd command = "REMOVE"

	ok_res    result = "OK\n"
	fail_res  result = "FAIL\n"
	error_res result = "ERROR\n"
)

func main() {
	listener, err := net.Listen(protocol, address)
	if err != nil {
		log.Fatalln("Failed to listen over "+protocol+" at "+address+" -> ", err.Error())
	}
	defer listener.Close()

	log.Println("Listening for " + protocol + " connections at " + address)

	db := syncDatabase{db: make(database)}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln("Error while accepting connection: ", err.Error())
		}

		go handleConnection(conn, &db)
	}
}

func handleConnection(conn net.Conn, db *syncDatabase) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		err := handleRequest(conn, db, reader)
		if err != nil {
			return
		}
	}
}

func handleRequest(conn net.Conn, db *syncDatabase, reader *bufio.Reader) error {
	message, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Error while receiving bytes: ", err.Error())
		return err
	}

	err, command, pkg, deps := parse(message)

	var response result
	response = error_res

	if err != nil {
		conn.Write([]byte(response))
	} else {
		switch command {
		case index_cmd:
			db.Lock()
			response = index(db.db, pkg, deps)
			db.Unlock()
		case remove_cmd:
			db.Lock()
			response = remove(db.db, pkg)
			db.Unlock()
		case query_cmd:
			db.RLock()
			response = query(db.db, pkg)
			db.RUnlock()
		}
		conn.Write([]byte(response))
	}
	return nil
}

func parseCommand(str string) (error, command) {
	switch str {
	case "INDEX":
		return nil, index_cmd
	case "REMOVE":
		return nil, remove_cmd
	case "QUERY":
		return nil, query_cmd
	default:
		var cmd command
		return errors.New(fmt.Sprintf("Invalid command: %v", str)), cmd
	}
}

func parseDeps(str string) []pkg {
	deps := []pkg{}

	//  Split 'str' on comma characters and filter empty strings out of resulting array.
	parsed := strings.Split(str, ",")
	var compacted []string
	for _, element := range parsed {
		if element != "" {
			compacted = append(compacted, element)
		}
	}

	//  construct a package for each dependency, appending the new package to the list of dependencies
	for _, dep := range compacted {
		deps = append(deps, pkg(dep))
	}
	return deps
}

func parse(message string) (error, command, pkg, []pkg) {
	var cmd command
	var candidate pkg
	var deps []pkg

	parsed := strings.Split(strings.TrimSpace(message), "|")

	if len(parsed) != 3 {
		return errors.New("Invalid message format."), cmd, candidate, deps
	}

	err, cmd := parseCommand(parsed[0])
	if err != nil {
		return err, cmd, candidate, deps
	}

	if len(parsed[1]) > 0 {
		candidate = pkg(parsed[1])
	} else {
		return errors.New("Package name cannot be empty."), cmd, candidate, deps
	}

	deps = parseDeps(parsed[2])

	return nil, cmd, candidate, deps
}

func index(db database, candidate pkg, deps []pkg) result {
	new_deps := make(dependencies)

	for _, dep := range deps {
		_, exists := db[dep]
		if !exists {
			return fail_res
		}
		new_deps[dep] = null
	}

	// if all of candidate's dependencies are indexed, index candidate
	db[candidate] = new_deps
	return ok_res
}

func remove(db database, candidate pkg) result {
	_, exists := db[candidate]
	if !exists {
		return ok_res
	}

	for _, deps := range db {
		_, exists := deps[candidate]
		if exists {
			return fail_res
		}
	}

	delete(db, candidate)
	return ok_res
}

func query(db database, candidate pkg) result {
	_, exists := db[candidate]
	if !exists {
		return fail_res
	}
	return ok_res
}
