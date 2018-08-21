#!/bin/bash
set -ex
set -o pipefail


# if you only do prepare
# ./test.sh prepare
PREPARE=$1
DIRNO=1
ExecComamnd=$(basename $(pwd))
ROOTDIR=$(pwd)
prepare_env () {
  TESTDIR=test-${ExecComamnd}-${DIRNO}
  mkdir -p $TESTDIR
  go build
  cp ./${ExecComamnd} $TESTDIR
  cd $TESTDIR
}

prepare_git () {
  git init
  git config --local user.email "git-fixup@example.com"
  git config --local user.name "git fixup"
  git commit --allow-empty -m "Initial commit"

  max=10
  for ((i=0; i <= $max; i++)); do
      touch file-${i}
      git add file-${i}
      git commit -m "Add file-${i}"
  done
}

testcase_end () {
  cd ${ROOTDIR}
  DIRNO=$(expr $DIRNO + 1)
}

test_run () {
  prepare_env
  prepare_git
  echo "*** START $1 ***"
  $1 "$1"
  echo "*** FINISH $1 ***"
  testcase_end
}

testcase1 () {
  NUM=5
  EXPECTED_ADDED_FILE_NUM=$(( $NUM + 1 ))

  git log --oneline
  ./${ExecComamnd}  -n 0..5 -d -f
  git log --oneline
  ADDED_FILE_NUM=`git diff HEAD^..HEAD --name-only | wc -l | tr -d ' '`

  if [ "$ADDED_FILE_NUM" == "$EXPECTED_ADDED_FILE_NUM" ]; then
    echo "*** test passed ***"
  else
    echo "*** test failed ***"
  fi
}

testcase2 () {
  NUM=5
  EXPECTED_ADDED_FILE_NUM=$(( $NUM + 1 ))

  git log --oneline
  ./${ExecComamnd}  -n 5 -d -f
  git log --oneline
  ADDED_FILE_NUM=`git diff HEAD^..HEAD --name-only | wc -l | tr -d ' '`

  if [ "$ADDED_FILE_NUM" == "$EXPECTED_ADDED_FILE_NUM" ]; then
    echo "*** test passed ***"
  else
    echo "*** test failed ***"
  fi
}

if [ "prepare" != "$PREPARE" ]; then
  test_run "testcase1"
  test_run "testcase2"
fi
