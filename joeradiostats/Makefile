BINARY=joeradiostats
# VERSION_TAG=`git describe 2>/dev/null | cut -f 1 -d '-' 2>/dev/null`
$(eval VERSION_TAG = $(shell git describe 2>/dev/null | cut -f 1 -d '-' 2>/dev/null))

# If no git tag is set, fallback to 'DEVELOPMENT'
ifeq ($(strip ${VERSION_TAG}),)
  VERSION_TAG := "DEVELOPMENT"
endif

COMMIT_HASH=`git rev-parse --short=8 HEAD 2>/dev/null`
BUILD_TIME=`date +%FT%T%z`
LDFLAGS=-ldflags "-s -w \
	-X github.com/metskem/rommel/joeradiostats/conf.CommitHash=${COMMIT_HASH} \
	-X github.com/metskem/rommel/joeradiostats/conf.BuildTime=${BUILD_TIME} \
	-X github.com/metskem/rommel/joeradiostats/conf.VersionTag=${VERSION_TAG}"

all: build linux darwin arm64

clean:
	go clean
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

release: clean linux darwin

# Installs our project: copies binaries
install:
	go install ${LDFLAGS}

build:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
	go build -o ./target/linux_amd64/${BINARY} ${LDFLAGS} .

linux_amd64:
	GOOS=linux GOARCH=amd64 go build -o ./target/linux_amd64/${BINARY} ${LDFLAGS} .

linux_arm64:
	GOOS=linux GOARCH=arm64 GOARM=7 go build -o ./target/linux_arm64/${BINARY} ${LDFLAGS} .
#   could not get this working on my Mac. So no cross-compile, for now I just compile it on my Raspberry PI

darwin_arm64:
	GOOS=darwin GOARCH=arm64 go build -o ./target/darwin_arm64/${BINARY} ${LDFLAGS} .

darwin_amd64:
	GOOS=darwin GOARCH=amd64 go build -o ./target/darwin_amd64/${BINARY} ${LDFLAGS} .
