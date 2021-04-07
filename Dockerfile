# Build
FROM golang:1.15 AS build
WORKDIR /go/src/github.com/target/pod-reaper
ENV CGO_ENABLED=0 GOOS=linux
COPY ./ ./
RUN go test ./cmd/... && \
    go test ./internal/... && \
    go build -o pod-reaper -a -installsuffix go ./cmd/pod-reaper

# Application
FROM scratch
COPY --from=build /go/src/github.com/target/pod-reaper/pod-reaper /
CMD ["/pod-reaper"]
