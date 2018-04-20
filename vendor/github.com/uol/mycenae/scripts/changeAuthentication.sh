#!/bin/bash

sed -ri 's/^([# ]+)?('"authenticator"':).*/\2 '"PasswordAuthenticator"'/' /etc/cassandra/cassandra.yaml
sed -ri 's/authorizer:.*/authorizer: "CassandraAuthorizer"/' /etc/cassandra/cassandra.yaml

sed -ri 's/#MAX_HEAP_SIZE="4G"/MAX_HEAP_SIZE="512M"/' /etc/cassandra/cassandra-env.sh
sed -ri 's/#HEAP_NEWSIZE="800M"/HEAP_NEWSIZE="512M"/' /etc/cassandra/cassandra-env.sh

exec /docker-entrypoint.sh "cassandra" "-f" "$@"