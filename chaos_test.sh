#!/bin/bash

IDS=`echo {1..50}`
PROF_PORT=3000
BASE_PORT=8000

function generate() {
	echo 'version: "3"'
	echo
	echo 'services:'

	for ID in ${IDS}; do
	  HOSTS=`for x in $(echo ${IDS} | sed -e "s/${ID}//"); do printf "http://cookfs_${x} "; done`

	  echo "  cookfs_${ID}:"
	  echo "    image: golang"
	  echo "    volumes: ['./bin/cookfs:/cookfs:ro']"
	  echo "    ports: ['$((${BASE_PORT} + ${ID})):80', '$((${PROF_PORT} + ${ID})):3000']"
	  echo "    command: /cookfs http://cookfs_${ID}:80 ${HOSTS}"
	  echo
	done
}

function start_all() {
	docker run --rm -v `pwd`:/go/src/github.com/macrat/cookfs -w /go/src/github.com/macrat/cookfs golang bash -c 'make bin/cookfs && chown 1000:1000 bin/cookfs'
	generate > docker-compose.yml
	docker-compose up -d
}

function stop_all() {
	generate > docker-compose.yml
	docker-compose down
	make clean
	rm docker-compose.yml
}

function info() {
	for ID in ${IDS}; do
		printf "${ID}: "
		curl --silent --max-time 0.1 http://localhost:$((${BASE_PORT} + ${ID}))/leader
		echo
	done
}

function leader() {
	make bin/cookctl >/dev/null 2>&1

	./bin/cookctl `for ID in ${IDS}; do printf -- "--server http://localhost:$((${BASE_PORT} + ${ID})) "; done` info --format json | jq -r .leader | sed -e 's/http:\/\///' -e 's/:80//'
}

function stop_random() {
	docker-compose stop $(echo `echo ${IDS} | grep -o '[0-9]\+' | sort -R | head -n ${1:-1} | xargs -I{} echo cookfs_{}`)
}

function stop_leader() {
	docker-compose stop $(leader)
}

function restart_random() {
	docker-compose restart $(echo `echo ${IDS} | grep -o '[0-9]\+' | sort -R | head -n ${1:-1} | xargs -I{} echo cookfs_{}`)
}

function restart_leader() {
	docker-compose restart $(leader)
}

function show_help() {
	echo "$1 COMMAND"
	echo
	echo 'COMMAND:'
	echo '  start - run containers'
	echo '  stop - stop and remove containers'
	echo '  stop_random [number] - stop random containers'
	echo '  stop_leader - stop leader container'
	echo '  restart_random [number] - restart random containers'
	echo '  restart_leader - restart leader container'
	echo '  gen - generate docker-compose.yml'
	echo '  info - get current term information'
}

case $1 in
	start ) start_all ;;
	stop ) stop_all ;;
	stop_random ) stop_random $2 ;;
	stop_leader ) stop_leader ;;
	restart_random ) restart_random $2 ;;
	restart_leader ) restart_leader ;;
	gen ) generate ;;
	info ) info ;;
	* ) show_help $0 >&2 ;;
esac
