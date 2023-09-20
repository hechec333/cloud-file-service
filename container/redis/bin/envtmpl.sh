#!/bin/bash

envsubst < redis-cluster.sh.tmpl > redis-cluster.sh
chmod +x redis-cluster.sh 
echo "monitering redis-cluster...."
./redis-cluster.sh 