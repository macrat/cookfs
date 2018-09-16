#!/bin/bash

IDS=`echo {1..20}`
BASE_PORT=8000

function generate() {
	echo 'version: "3"'
	echo
	echo 'services:'

	for ID in ${IDS}; do
	  HOSTS=`for x in $(echo ${IDS} | sed -e "s/${ID}//"); do printf "http://cookfs_${x} "; done`

	  echo "  cookfs_${ID}:"
	  echo "    image: golang:alpine"
	  echo "    volumes: ['./:/cookfs']"
	  echo "    ports: ['$((${BASE_PORT} + ${ID} - 1)):80']"
	  echo "    command: ash -c 'go run /cookfs/*.go http://cookfs_${ID}:80 ${HOSTS}'"
	  echo
	done
}

function info() {
	for ID in ${IDS}; do
		printf "${ID}: "
		curl --silent http://localhost:$((${BASE_PORT} + ${ID} - 1))/
		echo
	done
}

function stop_random() {
	docker-compose stop $(for i in `seq ${1:-1}`; do echo cookfs_`echo ${IDS} | grep -o '[0-9]\+' | sort -R | head -n 1`; done)
}

function stop_leader() {
	docker-compose stop cookfs_`curl -s http://localhost:${BASE_PORT}/ | jq -r .leader | sed -e 's/http:\/\/cookfs_//' -e 's/:80//'`
}

function show_help() {
	echo "$1 COMMAND"
	echo
	echo 'COMMAND:'
	echo '  start - run containers'
	echo '  stop - stop and remove containers'
	echo '  stop_random [number] - stop random containers'
	echo '  stop_leader - stop leader container'
	echo '  gen - generate docker-compose.yml'
	echo '  info - get current term information'
}

case $1 in
	start ) generate > docker-compose.yml && docker-compose up -d ;;
	stop ) generate > docker-compose.yml && docker-compose down ;;
	stop_random ) stop_random $2 ;;
	stop_leader ) stop_leader ;;
	gen ) generate ;;
	info ) info ;;
	* ) show_help $0 >&2 ;;
esac
