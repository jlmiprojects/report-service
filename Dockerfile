# ghcr/production build: see the Makefile's `docker`/`docker-prod` targets.
# `Dockerfile.local` is the near-identical copy
# broker-portal/deploy/docker-compose.yml builds from.
FROM golang:1.25-alpine AS build
ARG VERSION=0.0.0
ARG BUILD_SHA=unknown
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.ServiceVersion=${VERSION} -X main.Build=${BUILD_SHA}" -o /out/report-service ./cmd/main.go

FROM alpine:3
RUN apk add --no-cache tzdata
ENV TZ=Africa/Johannesburg
WORKDIR /app
COPY --from=build /out/report-service /app/report-service
COPY static /app/static
VOLUME ["/reports","/scripts"]
EXPOSE 3006
CMD ["/app/report-service"]
