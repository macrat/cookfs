.PHONY: all
all: get bin/cookfs bin/cookctl

.PHONY: get
get:
	go get -d

bin/cookfs: $(shell ls *.go cooklib/*.go)
	go build -o $@

bin/cookctl: $(shell ls cookctl/*.go cooklib/*.go)
	cd cookctl && go build -o ../$@

.PHONY: clean
clean:
	rm -r bin
