#!/bin/bash

# zookeeper
zkPodName='zookeeper'
docker rm -f "${zkPodName}"
docker run -d --name "${zkPodName}" -v $(pwd)/zoo.cfg:/conf/zoo.cfg zookeeper:3.4.11
sleep 5
zkIP=$(docker inspect --format "{{ .NetworkSettings.IPAddress }}" $zkPodName)
echo "$zkPodName listening on ip $zkIP"
echo "Zookeeper OK"
# /zookeeper

# build the node image
./build.sh 

function quit {
    docker rm -f et1 et2 et3 zookeeper
    exit
}

function createSlave {
    docker run -it -d --name ${1} --add-host="zookeeper.intranet:${zkIP}" electiontest
    sleep 5
    grepResult=$(docker logs ${1} | grep 'slave node created')
    if [ -z "$grepResult" ]; then
        echo "FAIL: expecting ${1} be a slave"
        quit
    else 
        echo "OK: ${1} is slave"
    fi
}

docker run -it -d --name et1 --add-host="zookeeper.intranet:${zkIP}" electiontest
sleep 5
grepResult=$(docker logs et1 | grep master)
if [ -z "$grepResult" ]; then
    echo "FAIL: expecting et1 be the master"
    quit
else 
    echo "OK: et1 is the master"
fi

createSlave 'et2'
createSlave 'et3'

echo "killing master node..."
docker rm -f et1
sleep 10

grepResultET2=$(docker logs et2 | grep 'trying to be the new master')
grepResultET3=$(docker logs et3 | grep 'trying to be the new master')
if [ -z "$grepResultET2" ] && [ -z "$grepResultET3" ]; then
    echo "FAIL: expecting et2 and et3 to try to be the new master"
    quit
else
    echo "OK: et2 and et3 tried to be the new master"
fi

grepResultET2=$(docker logs et2 | grep 'master node created')
grepResultET3=$(docker logs et3 | grep 'master node created')
if [ -z "$grepResultET2" ] && [ -z "$grepResultET3" ]; then
    echo "FAIL: expecting et2 or et3 be the master"
    quit
else
    echo "OK: et2 or et3 is the master: ${grepResultET2}${grepResultET3}"
fi

createSlave 'et1'

echo "OK: test is done!"

quit