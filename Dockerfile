# Build
FROM golang:1.17-alpine AS build
WORKDIR /src
ENV CGO_ENABLED=0 GOOS=linux

# leverage Docker layer caching for dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /pod-reaper -a -installsuffix go ./reaper

# Application
FROM scratch
COPY --from=build /pod-reaper /pod-reaper
CMD ["/pod-reaper"]
