FROM golang:alpine3.11 as builder

RUN apk add build-base

WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download 2> /dev/null

COPY . /app/
RUN make dist

FROM alpine:3.11
RUN apk add bash
COPY --from=builder /app/dist/* /usr/local/bin/
ENTRYPOINT [ "bash", "/usr/local/bin/entrypoint.sh" ]
