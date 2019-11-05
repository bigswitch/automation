#!/bin/bash -eux
# Copyright 2019, Big Switch Networks, Inc.

ROOTDIR=$(dirname $(readlink -f $0))/..
cd "$ROOTDIR"

make docker-terraform

# Add write permissions, so Jenkins can cleanup artifacts later
sudo chmod -R a+rwX ${ROOTDIR}
