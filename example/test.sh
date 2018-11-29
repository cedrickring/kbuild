#!/usr/bin/env bash

if [[ $# == 0 ]]; then
    echo "Please provide docker repository as first parameter"
    exit 0
fi

kbuild -t $1/test:latest