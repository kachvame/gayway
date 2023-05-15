FROM golang as development

WORKDIR /app

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

FROM development as build

RUN go build -o app ./cmd/gayway

FROM alpine:latest as certs

RUN apk --update add ca-certificates

FROM scratch as app

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build app /

ENTRYPOINT ["/app"]

