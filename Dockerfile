FROM golang:1.14
ADD . /go/src/pacman
WORKDIR /go/src/pacman
RUN go install
EXPOSE 8080
ENTRYPOINT ["/go/bin/pacman"]