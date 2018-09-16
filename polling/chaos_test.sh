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

function leader() {
	for ID in ${IDS}; do
		printf "${ID}: "
		curl --silent http://localhost:$((${BASE_PORT} + ${ID} - 1))/
		echo
	done
}

function show_help() {
	echo "$1 COMMAND"
	echo
	echo 'COMMAND:'
	echo '  start - run containers'
	echo '  stop - stop containers'
	echo '  gen - generate docker-compose.yml'
	echo '  leader - get leader information'
}

case $1 in
	start ) generate > docker-compose.yml && docker-compose up -d ;;
	stop ) generate > docker-compose.yml && docker-compose down ;;
	gen ) generate ;;
	leader ) leader ;;
	* ) show_help $0 >&2 ;;
esac
