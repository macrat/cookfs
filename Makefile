.PHONY: all
all: get cookfs_server

.PHONY: get
get:
	go get -d

.PHONY: cookfs_server
cookfs_server:
	go build -o $@

.PHONY: clean
clean:
	rm cookfs_server
