FROM golang:1.14
ADD . /go/src/github.com/anon-user/pacman
WORKDIR /go/src/github.com/anon-user/pacman
RUN go install
EXPOSE 8080
ENTRYPOINT ["/go/bin/pacman"]