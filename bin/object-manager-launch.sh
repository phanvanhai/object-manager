#!/bin/bash
#
# Copyright (c) 2018
# Tencent
#
# SPDX-License-Identifier: Apache-2.0
#

DIR=$PWD
CMD=../cmd

# Kill all object-manager* stuff
function cleanup {
	pkill object-manager
}

cd $CMD
exec -a object-manager ./object-manager  &
cd $DIR

trap cleanup EXIT

while : ; do sleep 1 ; done
