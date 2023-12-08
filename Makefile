SHELL := /bin/bash
TARGETS := uncork

.PHONY: all
all: $(TARGETS)

%: %.go
	go build -o $@ $<

.PHONY: clean
clean:
	rm -f $(TARGETS)

