# Build
FROM golang:1.14 AS build
WORKDIR /go/src/github.com/target/pod-reaper
ENV CGO_ENABLED=0 GOOS=linux
COPY ./ ./
RUN go get github.com/Masterminds/glide \
 && glide install --strip-vendor \
 && go test $(glide nv) \
 && go build -o pod-reaper -a -installsuffix go ./reaper

# Application
FROM scratch
COPY --from=build /go/src/github.com/target/pod-reaper/pod-reaper /
CMD ["/pod-reaper"]
