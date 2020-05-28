# pacman

<p align="center">
  <img src="http://www.clipartbest.com/cliparts/7ia/Rga/7iaRgaaiA.gif" alt="pacman ghost" width="125">
</p>

**pacman spins up a simple server to help manage packages and their dependencies.**

:warning: _any references to real package managers are used fictitiously. Other names, types, consts, and vars are the product of the developer's imagination, and any resemblance to [actual package managers](https://www.archlinux.org/pacman/) or arcade games, living or dead, is purely coincidental._

## Getting Started

### Prerequisites

Before you can build and run the server, you'll need to have `go` installed.
[Click here](https://golang.org/dl/) to download go, and follow any relevant setup instructions.

### Build

To run this project, compile the pacman.go file with `go build pacman.go` and then run the compiled executable with `./pacman`, or do both at once with `go run pacman.go`.

If initialization is successful, you'll see `Listening for tcp connections at 0.0.0.0:8080`.

### Test

You can run the Go test suite with `go test`.

### Docker

A `Dockerfile` is included.

To run pacman inside of a Docker container and execute the provided test suite against it, run the following:

```bash
docker build -t pacman . && docker run -p 8080:8080 --init -d pacman && ./do-package-tree_<platform>
```

## Usage

Messages from clients should follow this pattern:

```
<command>|<package>|<dependencies>\n
```

Where:

- `<command>` is mandatory, and is either `INDEX`, `REMOVE`, or `QUERY`
- `<package>` is mandatory, the name of the package referred to by the command, e.g. `mysql`, `openssl`, `pkg-config`, `postgresql`, etc.
- `<dependencies>` is optional, and if present it will be a comma-delimited list of packages that need to be present before `<package>` is installed. e.g. `cmake,sphinx-doc,xz`
- The message always ends with the character `\n`

**sample messages:**

```
INDEX|blinky|pinky,inky,clyde\n
INDEX|miss-pacman|\n
REMOVE|blinky|\n
QUERY|blinky|\n
```

<p align="center">
  <img src="https://ya-webdesign.com/transparent250_/pacman-cherry-png-6.png" alt="pacman cherries"> 
</p>
