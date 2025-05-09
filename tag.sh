#!/usr/bin/env bash

TAG=$1

git tag v$TAG
git push origin v$TAG