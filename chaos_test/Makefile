.PHONY: all
all: docker-compose.yml ../bin/cookfs ../bin/cookctl

../bin/cookfs: $(shell ls ../*.go ../cooklib/*.go ../plugins/*.go) ../Makefile
	cd .. && make bin/cookfs

../bin/cookctl: $(shell ls ../cookctl/*.go ../cooklib/*.go ../plugins/*.go) ../Makefile
	cd .. && make bin/cookctl

docker-compose.yml: generate-compose env
	./$< > $@

.PHONY: clean
clean: docker-compose.yml
	docker-compose down
	rm docker-compose.yml
