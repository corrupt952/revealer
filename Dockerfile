FROM golang:1.11.3 AS builder
WORKDIR /go/src/app
COPY . .
RUN GOARCH=amd64 CGO_ENABLED=0 go build -o env-revealer ./...

FROM scratch
COPY --from=builder /go/src/app/env-revealer /usr/local/bin/env-revealer
CMD ["/usr/local/bin/env-revealer"]
