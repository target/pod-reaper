# Build
FROM golang:1.13 AS build
WORKDIR /go/src/github.com/target/pod-reaper
ENV CGO_ENABLED=0 GOOS=linux
RUN go get github.com/Masterminds/glide
COPY glide.* ./
RUN glide install --strip-vendor
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN go test $(glide nv | grep -v /test/)
RUN go build -o pod-reaper -a -installsuffix go ./cmd/pod-reaper

# Application
FROM scratch
COPY --from=build /go/src/github.com/target/pod-reaper/pod-reaper /
CMD ["/pod-reaper"]
