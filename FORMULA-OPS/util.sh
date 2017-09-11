#!/usr/bin/env bash

pushd () {
    command pushd "$@" > /dev/null
}

popd () {
    command popd "$@" > /dev/null
}

if [[ ${PWD} != *"pocket-formula/devops" ]]; then
  echo "should be executed in pocket-formula/devops directory"
  exit
fi
