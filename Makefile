GO=go

BUILDDIR=build

all: tgrep

tgrep:
	$(GO) build -o $(BUILDDIR)/$@ cmd/$@/*.go

clean:
	rm -rf $(BUILDDIR)

.PHONY: clean all
