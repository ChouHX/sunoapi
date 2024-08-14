FROM golang:latest AS builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -ldflags "-s -w -extldflags '-static'" -o sunoapi

FROM debian:latest

RUN apt-get update && apt-get install -y ca-certificates

WORKDIR /root/

COPY --from=builder /build/sunoapi /root/
COPY --from=builder /build/template /root/template/

RUN chmod +x /root/sunoapi

CMD ["/root/sunoapi"]

