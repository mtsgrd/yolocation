# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Build the outyet command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go get github.com/revel/cmd/revel
RUN go get -u -d github.com/mtsgrd/yolocation/...

# Document that the service listens on port 8080.
EXPOSE 8080

# Run the outyet command by default when the container starts.
CMD revel run github.com/mtsgrd/yolocation prod
