FROM golang:1.18-alpine as builder
WORKDIR /usr/src/app
# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" .

FROM scratch
COPY --from=builder /usr/src/app/psql-demo /usr/local/bin/psql-demo
CMD ["psql-demo"]
