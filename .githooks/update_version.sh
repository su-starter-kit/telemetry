#!/bin/bash

## Defining binaries and tools.
GET_VER=./getver


## Defining useful functions:
### Defining log functions:
Info () {
    echo "[INFO] " $1
}
Error () {
    echo "[ERROR] " $1
}
Hint () {
    echo "[HINT] " $1
}

## Check if the current branch is main
CURRENT_BRANCH=$(git symbolic-ref HEAD | sed -e 's,.*/\(.*\),\1,')
if [[ $CURRENT_BRANCH != "main" ]]; then
    Error "Can only version in main branch!"
    exit 1
fi

## Get current dir
CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

## Getting the right version of scripts
SCRIPTS_BIN=$CURRENT_DIR/bin_mac # defaults to mac

if [[ $(uname -s) == "Linux" ]]; then
  Info "Linux platform, setting the configs to linux..."
  SCRIPTS_BIN=$CURRENT_DIR/bin_lnx # sets to linux
fi

# Clean all local tags
Info "cleaning all local tags..."
git tag -d $(git tag -l)

# Get latest tag
Info "fetching repo for latest tags..."
git fetch --all --tags

# Get current version with git tag.
Info "getting latest version from git remote tag..."
CURRENT_VERSION=$(git describe --tags --abbrev=0)
Info "current version: ${CURRENT_VERSION}" 

# Get latest semantic version commit comment.
Info "getting comment from latest commit (expecting semantic comment: eg.: 'fix: this is a fix')..."
COMMENT_FROM_LATEST_COMMIT=$(git log -1 --pretty=%B)
Info "comment from latest commit is: ${COMMENT_FROM_LATEST_COMMIT}" 

# Use getver tool to calculate next version.
Info "calculating the next version..."
NEXT_VERSION=$(${SCRIPTS_BIN}/getver --current-version $CURRENT_VERSION --comment "${COMMENT_FROM_LATEST_COMMIT}")

# Verify error level.
ERR=$?
if [[ $ERR != 0 ]]; then
    Error "error while getting next version!"
    Error "please verify if your last commit message is compliant with the semantic commit comment required by this repo and package."
    Error "last commit: ${COMMENT_FROM_LATEST_COMMIT}" 
    Hint "To correct the commit message you can run: <git commit --amend>"
    exit $ERR
fi 

Info "next version is: ${NEXT_VERSION}" 

# Generate new version with `git tag`.
Info "updating version..."
git tag -a $NEXT_VERSION -m "${COMMENT_FROM_LATEST_COMMIT}"
git push origin --tags $NEXT_VERSION
ERR=$?
if [[ $ERR != 0 ]]; then
    Error "error trying to update version!"
    exit $ERR
fi 

Info "Successfully updated to version ${NEXT_VERSION} !" 
