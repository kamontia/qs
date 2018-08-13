#!/bin/bash
set -eux

TESTDIR=test-git-fixup-$$

mkdir -p $TESTDIR
cp ./git-fixup $TESTDIR
cd $TESTDIR

git init
git config --local user.email "git-fixup@example.com"
git config --local user.name "git fixup"

touch fileA
git add fileA
git commit -m "add fileA"
touch fileB
git add fileB
./git-fixup -f -n 1
