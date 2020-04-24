ARG GO_VERSION=1.12
FROM golang:${GO_VERSION}-alpine AS builder
RUN apk update && apk add gcc libc-dev make git
WORKDIR /lil/
COPY ./ ./
ENV CGO_ENABLED=0 GOOS=linux
RUN make build

FROM alpine:latest AS deploy
RUN apk --no-cache add bash ca-certificates
WORKDIR /lil/
COPY --from=builder /lil/templates ./templates
COPY --from=builder /lil/lil.bin /lil/config.toml.sample ./
RUN mkdir -p /etc/lil && cp config.toml.sample /etc/lil/config.toml
# Define data volumes
VOLUME ["/etc/lil"]
# Mount the config file and instruments when running container. Image doesn't need
CMD ["./lil.bin", "--config", "/etc/lil/config.toml"]
