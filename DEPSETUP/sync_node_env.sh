#!/usr/bin/env bash

rsync -rlmvt -e ssh ${HOME}/Workspace/GOPLACE/src/github.com/stkim1/pc-node-agent/* pc-node1:/opt/gopkg/src/github.com/stkim1/pc-node-agent
rsync -rlmvt -e ssh ${HOME}/Workspace/GOPLACE/src/github.com/stkim1/pc-node-agent/* pc-node2:/opt/gopkg/src/github.com/stkim1/pc-node-agent
rsync -rlmvt -e ssh ${HOME}/Workspace/GOPLACE/src/github.com/stkim1/pc-node-agent/* pc-node3:/opt/gopkg/src/github.com/stkim1/pc-node-agent
rsync -rlmvt -e ssh ${HOME}/Workspace/GOPLACE/src/github.com/stkim1/pc-node-agent/* pc-node4:/opt/gopkg/src/github.com/stkim1/pc-node-agent
rsync -rlmvt -e ssh ${HOME}/Workspace/GOPLACE/src/github.com/stkim1/pc-node-agent/* pc-node5:/opt/gopkg/src/github.com/stkim1/pc-node-agent
rsync -rlmvt -e ssh ${HOME}/Workspace/GOPLACE/src/github.com/stkim1/pc-node-agent/* pc-node6:/opt/gopkg/src/github.com/stkim1/pc-node-agent