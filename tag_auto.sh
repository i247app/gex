#!/usr/bin/env bash

# Run tag.sh with an incrementing version number until it succeeds

CURRENT_VERSION=0.0.1

# Increment the version number
NEW_VERSION=$CURRENT_VERSION

# Run tag.sh with the new version number
while true; do
    bash tag.sh $NEW_VERSION

    if [ $? -eq 0 ]; then
        break
    fi

    echo "Tag $NEW_VERSION already exists, incrementing version number..."
    NEW_VERSION=$(echo $NEW_VERSION | awk -F. -v OFS=. '{$NF++; print}')
done