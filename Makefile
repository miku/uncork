SHELL := /bin/bash
TARGETS := uncork

.PHONY: all
all: $(TARGETS)

%: %.go
	go build -o $@ $<

.PHONY: release
	# need to: export GITHUB_TOKEN="ghp_12345"
	goreleaser release --rm-dist

.PHONY: clean
clean:
	rm -f $(TARGETS)

