# pacman

![pacman ghost](https://png2.cleanpng.com/sh/af1843210f0c268f59adbeddefc63756/L0KzQYm3UsE1N6VuiZH0aYP2gLBuTgBia15yedC2Z3HwdcS0hBhwe6V4RdR1dXWwd7n2kCQua51uiNN7dIOwRbKBVMgyaZRpUdRtY0SxRoK5WMgyP2I2TaMDNkO2Q4mBWMkyQV91htk=/kisspng-pac-man-games-ghosts-blue-ghost-cliparts-5a8481acd9bdc4.6128817115186333888919.png)

pacman spins up a simple server to help manage packages and their dependencies.

Any references to real package managers are used fictitiously. Other names, types, consts, and vars are the product of the developer's imagination, and any resemblance to [actual package managers](https://www.archlinux.org/pacman/), living or dead, is purely coincidental.

## getting started

### prerequisites

Before you can build and run the server, you'll need to have `go` installed.
[Click here](https://golang.org/dl/) to download go, and follow any relevant setup instructions.

### building

To run this project, compile the pacman.go file with go build pacman.go and then run the compiled executable with ./pacman, or do both at once with go run pacman.go.

If the connection is successful, you'll see `Listening for tcp connections at localhost:8080`.

### testing

You can run the included test suite with `go test`.

## usage

Messages from clients should follow this pattern:

```
<command>|<package>|<dependencies>
```

**Where:**

- `<command>` (**mandatory**) - is either `INDEX`, `REMOVE`, or `QUERY`
- `<package>` (**mandatory**) - the name of the package referred to by the command
- `<dependencies>` (_optional_) - comma-delimited list of packages that need to
  be present before `<package>` is installed
- The message should end with a `\n` character

### sample messages

```
INDEX|blinky|pinky,inky,clyde\n
INDEX|miss-pacman|\n
REMOVE|blinky|\n
QUERY|blinky|\n
```

## principles
