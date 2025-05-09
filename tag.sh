#!/usr/bin/env bash

TAG=$1

# Check if tag is provided
if [ -z "$TAG" ]; then
    echo "Usage: $0 <tag>"
    exit 1
fi

# Check if tag already exists
if git tag -l | grep -q "^v$TAG$"; then
    echo "Tag v$TAG already exists"
    exit 1
fi

# Prepend v to tag if it doesn't already have it
if [[ "$TAG" != v* ]]; then
    TAG="v$TAG"
fi

# Create tag
git tag $TAG
git push origin $TAG