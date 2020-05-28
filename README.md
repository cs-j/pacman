# pacman

![pacman ghost](http://www.clipartbest.com/cliparts/7ia/Rga/7iaRgaaiA.gif)

pacman spins up a simple server to help manage packages and their dependencies.

:warning: _Any references to real package managers are used fictitiously. Other names, types, consts, and vars are the product of the developer's imagination, and any resemblance to [actual package managers](https://www.archlinux.org/pacman/), living or dead, is purely coincidental._

## Getting Started

### prerequisites

Before you can build and run the server, you'll need to have `go` installed.
[Click here](https://golang.org/dl/) to download go, and follow any relevant setup instructions.

### build

To run this project, compile the pacman.go file with `go build pacman.go` and then run the compiled executable with `./pacman`, or do both at once with `go run pacman.go`.

If the connection is successful, you'll see `Listening for tcp connections at localhost:8080`.

### test

You can run the included test suite with `go test`.

## Usage

Messages from clients should follow this pattern:

```
<command>|<package>|<dependencies>
```

**where:**

- `<command>` (**mandatory**) - is either `INDEX`, `REMOVE`, or `QUERY`
- `<package>` (**mandatory**) - the name of the package referred to by the command
- `<dependencies>` (_optional_) - comma-delimited list of packages that need to
  be present before `<package>` is installed
- The message should end with a `\n` character

**sample messages:**

```
INDEX|blinky|pinky,inky,clyde\n
INDEX|miss-pacman|\n
REMOVE|blinky|\n
QUERY|blinky|\n
```

## Principles
