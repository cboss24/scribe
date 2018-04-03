# build stage
FROM golang:alpine AS build
RUN apk update && apk upgrade && apk add --no-cache alpine-sdk
RUN go get -u github.com/golang/dep/cmd/dep
RUN mkdir -p /go/src/github.com/cboss24/scribe
WORKDIR /go/src/github.com/cboss24/scribe
COPY vendor Gopkg.lock Gopkg.toml ./
RUN dep ensure -vendor-only
COPY src src
RUN go build -o scribe ./src

# final stage
FROM alpine
RUN apk update && apk upgrade && apk add --no-cache ca-certificates
COPY --from=build /go/src/github.com/cboss24/scribe/scribe scribe
ENTRYPOINT ["./scribe"]