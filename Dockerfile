FROM alpine
RUN apk update && apk upgrade && apk add --no-cache ca-certificates
COPY scribe ./
ENTRYPOINT ["./scribe"]