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
  TESTDIR=test-$$-${ExecComamnd}-${DIRNO}
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

teardown () {
  cd ${ROOTDIR}
  DIRNO=$(expr $DIRNO + 1)
}

test_check () {
  NUM=$1
  EXPECTED_ADDED_FILE_NUM=$2
  git log --oneline
  ./$ExecComamnd -n $NUM -f -d
  git log --oneline
  ADDED_FILE_NUM=`git diff HEAD^..HEAD --name-only | wc -l | tr -d ' '`

  if [ "$ADDED_FILE_NUM" == "$EXPECTED_ADDED_FILE_NUM" ]; then
    echo "[passed] RUN ./$ExecComamnd -n $NUM -f -d RESULT $EXPECTED_ADDED_FILE_NUM" >> ./../test-$$-result
  else
    echo "[failed] RUN ./$ExecComamnd -n $NUM -f -d RESULT $ADDED_FILE_NUM" >> ./../test-$$-result
  fi
}

test_run () {
  prepare_env
  prepare_git
  echo "*** START $1 ***"
  test_check $1 $2
  echo "*** FINISH $1 ***"
  teardown
}

:
: main
:
if [ "prepare" != "$PREPARE" ]; then
  # test for `./qs -n 5 -f -d` and expected result value is 6
  test_run 5 6
  # test for `./qs -n 0..5 -f -d` and expected result value is 6
  test_run 0..5 6
  echo "*** test result ***"
  cat ./test-$$-result
fi

