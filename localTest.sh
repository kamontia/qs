#!/bin/bash
set -xe
set -eo pipefail

ExecComamnd=$(basename $(pwd))
TESTDIR=test-${ExecComamnd}-$$

mkdir -p $TESTDIR
go build
cp ./${ExecComamnd} $TESTDIR
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
max=10
for ((i=0; i <= $max; i++)); do
    touch file-${i}
    git add file-${i}
    git commit -m "Add file-${i}"
done


git log --oneline
./${ExecComamnd} -n 5
git log --oneline
git diff HEAD^..HEAD --name-only