# Build
FROM golang:1.7 AS build
WORKDIR /go/src/github.com/target/pod-reaper
ENV CGO_ENABLED=0 GOOS=linux
COPY ./ ./
RUN go get github.com/Masterminds/glide
RUN glide install --strip-vendor
RUN go test . ./rules
RUN go build -a -installsuffix go

# Application
FROM scratch
COPY --from=build /go/src/github.com/target/pod-reaper/pod-reaper /
CMD ["/pod-reaper"]
