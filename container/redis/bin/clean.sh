#!/bin/bash


docker-compose down
rm -f redis-cluster.sh
rm -rf ../etc/*
rm -rf ../data/*
rm -rf ../../var/.redis