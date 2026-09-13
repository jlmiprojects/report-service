# This how we want to name the binary output
BINARY=report-service
GOOS=linux
GOARCH=amd64
DOCKER_HUB=ghcr.io/blueasset

# These are the values we want to pass for VERSION and BUILD
VERSION=`git describe --tags --always --dirty`
BUILD=`git rev-parse HEAD`

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-X main.ServiceVersion=${VERSION} -X main.Build=${BUILD}"

# Compiles assets/input.css -> static/output.css and inlines it into
# reports/styles.partial.html (the partial report templates pull in with
# {{ template "reportstyles" . }}). Run after changing any report template's
# classes or the brand palette. Needs `npm install` once. Kept separate from
# `build` so a Go-only build box doesn't need node — commit the generated
# reports/styles.partial.html.
css:
	npm run build

# Builds the project
build: clean prepare
	CGO_ENABLED=0 GOPRIVATE=github.com/blueasset/* GOOS=${GOOS} GOARCH=${GOARCH} go build ${LDFLAGS} -o builds/${VERSION}/${BINARY} cmd/main.go

dist: build
	rm -rf dist/${BINARY}-${VERSION}
	mkdir -p dist/${BINARY}-${VERSION}
	mkdir -p dist/${BINARY}-${VERSION}/bin
	mkdir -p dist/${BINARY}-${VERSION}/conf
	cp builds/${VERSION}/${BINARY} dist/${BINARY}-${VERSION}/bin
	tar -czvf dist/${BINARY}-${VERSION}.tar.gz dist/${BINARY}-${VERSION}

# Installs our project: copies binaries
install:
	go install ${LDFLAGS}

# Cleans our project: deletes binaries
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

prepare:
	mkdir -p builds/${VERSION}

docker: build
	docker build --build-arg version=${VERSION} --build-arg ms=${BINARY} -t ${DOCKER_HUB}/${BINARY}:${VERSION} .
	docker push ${DOCKER_HUB}/${BINARY}:${VERSION}
run: 
	go run cmd/main
