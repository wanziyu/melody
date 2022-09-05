# Build the manager binary
FROM golang:alpine AS build-env

# Copy in the go src
ADD . /go/src/melody/cmd/controllers
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct

WORKDIR /go/src/melody/cmd/controllers
# Build
RUN if [ "$(uname -m)" = "ppc64le" ]; then \
        CGO_ENABLED=0 GOOS=linux GOARCH=ppc64le go build -a -o melody-controller .; \
    elif [ "$(uname -m)" = "aarch64" ]; then \
        CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -o melody-controller .; \
    else \
        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o melody-controller .; \
    fi

# Copy the controller-manager into a thin image
FROM alpine:3.7
WORKDIR /app
RUN apk update && apk add ca-certificates
COPY --from=build-env /go/src/melody/cmd/controllers/melody-controller .
USER 1000
ENTRYPOINT ["./melody-controller"]