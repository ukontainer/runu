.PHONY: all

GO := go
SOURCES := $(shell find . 2>&1 | grep -E '.*\.(c|h|go)$$')
COMMIT_NO := $(shell git rev-parse HEAD 2> /dev/null || true)
COMMIT := $(if $(shell git status --porcelain --untracked-files=no),"${COMMIT_NO}-dirty","${COMMIT_NO}")
#VERSION := ${shell cat ./VERSION}
EXTRA_FLAGS := -gcflags "-N -l"

.DEFAULT: runu

runu: $(SOURCES) Makefile
	$(GO) build -buildmode=pie $(EXTRA_FLAGS) -ldflags \
	"-X main.gitCommit=${COMMIT} -X main.version=${VERSION} $(EXTRA_LDFLAGS)" \
	-tags "$(BUILDTAGS)" -o runu .

all: runu

install: all
	$(INSTALL) -d $(DESTDIR)$(PREFIX)/bin/
	$(INSTALL) -m 755 runu $(DESTDIR)$(PREFIX)/bin/
