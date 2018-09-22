.PHONY: all
all: get bin/cookfs bin/cookctl

.PHONY: get
get:
	go get -d

bin/cookfs: $(shell ls *.go cookfs/*.go)
	go build -o $@

bin/cookctl: $(shell ls cookctl/*.go cookfs/*.go)
	cd cookctl && go build -o ../$@

.PHONY: clean
clean:
	rm -r bin
