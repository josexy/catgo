NAME=catgo
BINDIR=bin
MODULE=github.com/josexy/catgo
PACKAGE=main.go
VERSION=$(shell git describe --abbrev=0 --tags HEAD 2> /dev/null || echo "unknown")
GIT_COMMIT=$(shell git rev-parse --short HEAD)
GO_VERSION=$(shell go version 2> /dev/null)
BUILD_TIME=$(shell date 2> /dev/null)
LDFLAGS+=-w -s
LDFLAGS+=-X "$(MODULE)/version.Version=$(VERSION)"
LDFLAGS+=-X "$(MODULE)/version.GitCommit=$(GIT_COMMIT)"
LDFLAGS+=-X "$(MODULE)/version.BuildTime=$(BUILD_TIME)"
LDFLAGS+=-X "$(MODULE)/version.GoVersion=$(GO_VERSION)"

GOBUILD=CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)'

UNIX_ARCH_LIST = \
	darwin-amd64 \
	darwin-arm64 \
	linux-amd64 \
	linux-arm64 \
	linux-armv5 \
	linux-armv6 \
	linux-armv7

WINDOWS_ARCH_LIST = \
	windows-amd64 \
	windows-arm64

all: linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64 windows-arm64

build: 
	$(GOBUILD) -o $(BINDIR)/$(NAME) $(PACKAGE)

linux-amd64:
	GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(PACKAGE)

linux-arm64:
	GOARCH=arm64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(PACKAGE)

linux-armv5:
	GOARCH=arm GOARM=5 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(PACKAGE)

linux-armv6:
	GOARCH=arm GOARM=6 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(PACKAGE)

linux-armv7:
	GOARCH=arm GOARM=7 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(PACKAGE)

darwin-amd64:
	GOARCH=amd64 GOOS=darwin $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(PACKAGE)

darwin-arm64:
	GOARCH=arm64 GOOS=darwin $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(PACKAGE)

windows-amd64:
	GOARCH=amd64 GOOS=windows $(GOBUILD) -o $(BINDIR)/$(NAME)-$@.exe $(PACKAGE)

windows-arm64:
	GOARCH=arm64 GOOS=windows $(GOBUILD) -o $(BINDIR)/$(NAME)-$@.exe $(PACKAGE)

unix_releases := $(addsuffix .zip, $(UNIX_ARCH_LIST))
windows_releases := $(addsuffix .zip, $(WINDOWS_ARCH_LIST))

$(unix_releases): %.zip: %
	@zip -qmj $(BINDIR)/$(NAME)-$(basename $@).zip $(BINDIR)/$(NAME)-$(basename $@)

$(windows_releases): %.zip: %
	@zip -qmj $(BINDIR)/$(NAME)-$(basename $@).zip $(BINDIR)/$(NAME)-$(basename $@).exe

releases: $(unix_releases) $(windows_releases)

clean:
	rm $(BINDIR)/$(NAME)-*