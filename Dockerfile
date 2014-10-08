# Start with a golang image. It's that easy yo!
FROM golang

# Get revel tool and yolocation.
RUN go get github.com/revel/cmd/revel
RUN go get -u -d github.com/mtsgrd/yolocation/...

EXPOSE 8080
CMD revel run github.com/mtsgrd/yolocation prod
