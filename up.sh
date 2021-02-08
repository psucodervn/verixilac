#!/usr/bin/env bash

REMOTE=ec2-user@ec2-3-0-57-42.ap-southeast-1.compute.amazonaws.com
REMOTE_DIR=/home/ec2-user/verixilac
TAG=${1}

sed -i.bak "s/verixilac:.*/verixilac:${TAG}/g" deploy/docker-compose.yaml

rsync -a ./deploy/ ${REMOTE}:${REMOTE_DIR}

# shellcheck disable=SC2029
ssh ${REMOTE} "cd $REMOTE_DIR && ./up.sh"
