.PHONY: all
all: bin/cookfs bin/cookctl

.PHONY: get
get:
	go get -d

bin/cookfs: get $(shell ls *.go cookfs/*.go)
	go build -o $@

bin/cookctl: get $(shell ls cookctl/*.go cookfs/*.go)
	cd cookctl && go build -o ../$@

.PHONY: clean
clean:
	rm -r bin
