#!/bin/bash

POD_NAME='consulServer'
docker rm -f "${POD_NAME}"

arguments=(
	'-d'
	'--name' "${POD_NAME}"
)

docker run "${arguments[@]}" 'jenkins.macs.intranet:5000/mycenae/consul-server:v1'

consulServerIp=$(docker inspect --format "{{ .NetworkSettings.IPAddress }}" "${POD_NAME}")

echo "Consul Server OK (${consulServerIp})"