# Local docker-compose build: builds reports-service from source. Not used by
# the ghcr/production pipeline (see Makefile's `docker` target for that) —
# see local-stack/ in the sibling directory for the compose file that
# references this.
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/report-service ./cmd/main.go

FROM alpine:3.15.4
RUN apk add --no-cache tzdata
ENV TZ=Africa/Johannesburg
WORKDIR /app
COPY --from=build /out/report-service /app/report-service
COPY static /app/static
VOLUME ["/reports","/scripts"]
EXPOSE 3006
CMD ["/app/report-service"]
