#!/bin/bash
set -eux
set -o pipefail

TESTDIR=test-git-fixup-$$

mkdir -p $TESTDIR
cp ./git-fixup $TESTDIR
cd $TESTDIR

:
: prepare git
:
git init
git config --local user.email "git-fixup@example.com"
git config --local user.name "git fixup"
git commit --allow-empty -m "Initial commit"

:
: prepare test
:
touch fileA
git add fileA
git commit -m "Add fileA"
touch fileB
git add fileB

:
: test
:
./git-fixup -f
ADDED_FILE_NUM=`git diff HEAD^..HEAD --name-only | wc -l | tr -d ' '`
if [ "$ADDED_FILE_NUM" == "2" ]; then
  echo "*** test passed ***"
  exit 0
else
  echo "*** test failed ***"
  exit 1
fi
