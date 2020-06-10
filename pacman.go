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

//	Empty struct takes no space in memory.
//	Use this as the value in the dependencies map since there's no information
//	to associate with each dependency, other than its presence in the map.
//	This is a fake Set in Go.
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
	address  = "0.0.0.0:8080"
	protocol = "tcp"

	indexCmd  command = "INDEX"
	queryCmd  command = "QUERY"
	removeCmd command = "REMOVE"

	okRes    result = "OK\n"
	failRes  result = "FAIL\n"
	errorRes result = "ERROR\n"
)

//  Stand up TCP server and start handling connections.
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

		//	a goroutine is a function that is capable of running concurrently with other functions.
		//	Normally when we invoke a function our program will execute all the statements in a function and then return to the next line following the invocation.
		//	With a goroutine we return immediately to the next line and don't wait for the function to complete.
		go handleConnection(conn, &db)
	}
}

//  Handle any number of requests on a connection and clean up after it.
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

//  Read a message off the connection, parse it, perform core operation.
func handleRequest(conn net.Conn, db *syncDatabase, reader *bufio.Reader) error {
	message, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Error while receiving bytes: ", err.Error())
		return err
	}

	command, pkg, deps, err := parse(message)

	var response result
	response = errorRes

	if err != nil {
		conn.Write([]byte(response))
	} else {
		switch command {
		case indexCmd:
			db.Lock()
			response = index(db.db, pkg, deps)
			db.Unlock()
		case removeCmd:
			db.Lock()
			response = remove(db.db, pkg)
			db.Unlock()
		case queryCmd:
			db.RLock()
			response = query(db.db, pkg)
			db.RUnlock()
		}
		conn.Write([]byte(response))
	}
	return nil
}

// Parse string as one of three possible commands.
//
// Errors:
//   - If command is not one of the three expected values.
func parseCommand(str string) (command, error) {
	switch str {
	case "INDEX":
		return indexCmd, nil
	case "REMOVE":
		return removeCmd, nil
	case "QUERY":
		return queryCmd, nil
	default:
		var cmd command
		return cmd, fmt.Errorf("Invalid command: %v", str)
	}
}

//	Take in a string and return a slice of pkgs.
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

	//  Construct a package for each dependency, appending the new package to the list of dependencies.
	for _, dep := range compacted {
		deps = append(deps, pkg(dep))
	}
	return deps
}

//	Divide message into its constituent data types.
func parse(message string) (command, pkg, []pkg, error) {
	var cmd command
	var candidate pkg
	var deps []pkg

	parsed := strings.Split(strings.TrimSpace(message), "|")

	if len(parsed) != 3 {
		return cmd, candidate, deps, errors.New("invalid message format")
	}

	cmd, err := parseCommand(parsed[0])
	if err != nil {
		return cmd, candidate, deps, err
	}

	if len(parsed[1]) > 0 {
		candidate = pkg(parsed[1])
	} else {
		return cmd, candidate, deps, errors.New("package name cannot be empty")
	}

	deps = parseDeps(parsed[2])

	return cmd, candidate, deps, nil
}

//	Attempt to index a pkg and its deps to the db, return a response code.
func index(db database, candidate pkg, deps []pkg) result {
	newDeps := make(dependencies)

	for _, dep := range deps {
		_, exists := db[dep]
		//	Return `FAIL\n` if the candidate cannot be indexed because some of its dependencies aren't indexed yet.
		if !exists {
			return failRes
		}
		//	Move deps that do already exist in db from slice of pkgs into newDeps map.
		newDeps[dep] = null
	}

	//	Index candidate and its newDeps into db.
	//	If a candidate already existed in db, update its list of dependencies to the one provided with the latest command.
	db[candidate] = newDeps
	//	Return `OK\n` if the candidate could be indexed.
	return okRes
}

//	Attempt to remove a pkg from the db, return a response code.
func remove(db database, candidate pkg) result {

	//	Return `OK\n` if the candidate wasn't indexed.
	_, exists := db[candidate]
	if !exists {
		return okRes
	}

	//	Return `FAIL\n` if the candidate could not be removed from the index because some other indexed package depends on it.
	for _, deps := range db {
		_, exists := deps[candidate]
		if exists {
			return failRes
		}
	}

	//	Remove candidate from db.
	delete(db, candidate)
	//	Return `OK\n` if the candidate could be removed.
	return okRes
}

//	Look up a pkg in the db, return a response code.
func query(db database, candidate pkg) result {
	_, exists := db[candidate]

	//	Return `FAIL\n` if the candidate isn't indexed.
	if !exists {
		return failRes
	}

	//	Return `OK\n` if the candidate exists in the db.
	return okRes
}
