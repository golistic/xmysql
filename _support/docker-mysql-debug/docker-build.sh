#!/bin/sh

#
# Copyright (c) 2022, Geert JM Vanderkelen
#

FETCH="wget -q"

if ! command -v wget &>/dev/null;
then
    if command -v curl &>/dev/null;
    then
      FETCH="curl -s -O"
    else
      echo "neither wget or curl available"
      exit 1
    fi
fi

BASEURL=https://raw.githubusercontent.com/docker-library/mysql/master/8.0

${FETCH} ${BASEURL}/docker-entrypoint.sh
${FETCH} ${BASEURL}/Dockerfile.debian

CWD=$(pwd)
mkdir config config/conf.d &>/dev/null
cd config
${FETCH} ${BASEURL}/config/my.cnf
cd conf.d
${FETCH} ${BASEURL}/config/conf.d/docker.cnf


cd ${CWD}
patch -o Dockerfile Dockerfile.debian dockerfile.patch
rm Dockerfile.debian
docker build -t mysql-debug:8.0 .