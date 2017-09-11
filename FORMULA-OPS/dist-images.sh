#!/usr/bin/env bash

source ./util.sh

pushd ${PWD}
cd ../archive
rsync -rlmvt -e ssh ./* us-west-01:/cdn-content/v014/
rsync -rlmvt -e ssh ./* us-south-01:/cdn-content/v014/
rsync -rlmvt -e ssh ./* us-south-02:/cdn-content/v014/
rsync -rlmvt -e ssh ./* us-east-01:/cdn-content/v014/
popd
