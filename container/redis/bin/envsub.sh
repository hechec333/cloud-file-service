#!/bin/bash

relative=../etc
envsubst < redis-conf.tmpl > ${relative}/${HostName}/redis.conf