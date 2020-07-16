FROM golang:alpine3.11 as builder

RUN apk add make

WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download 2> /dev/null

COPY . /app/
RUN make dist

FROM alpine:3.11
COPY --from=builder /app/dist/* /usr/local/bin/
