# vim: set ft=dockerfile :
# syntax=docker/dockerfile:1

ARG BASE_IMAGE=alpine:3.17
ARG GO_IMAGE=golang:1.20.6-alpine3.17

FROM ${GO_IMAGE} as go-builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /go/src/alertmanager-mqtt-bridge
COPY go.* ./

# Install library dependencies
RUN go mod download

# Copy the entire project and build it
COPY . ./
RUN go build -o /bin/alertmanager-mqtt-bridge

# Final stage
FROM ${BASE_IMAGE}

COPY --from=go-builder /bin/alertmanager-mqtt-bridge /bin/alertmanager-mqtt-bridge

EXPOSE 8031

ENTRYPOINT ["/bin/alertmanager-mqtt-bridge"]
