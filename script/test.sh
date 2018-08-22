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

git_init () {
  git init
  git config --local user.email "git-fixup@example.com"
  git config --local user.name "git fixup"
  git commit --allow-empty -m "Initial commit"
}

git_pre_commit () {
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
  ret=$?

  if [ "$ret" != "0" ]; then
    echo "[failed] RUN ./$ExecComamnd -n $NUM -f -d RESULT qs non-zero status code $ret" >> ./../test-$$-result
    return 1
  fi

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
  git_init
  git_pre_commit
  echo "*** START $1 ***"
  test_check $1 $2
  echo "*** FINISH $1 ***"
  teardown
}

:
: main
:
if [ "prepare" == "$PREPARE" ]; then
  prepare_env
  git_init
  git_pre_commit
  echo "*** create $TESTDIR ***"
else
  # test for `./qs -n 5 -f -d` and expected result value is 6
  test_run 5 6
  # test for `./qs -n 0..5 -f -d` and expected result value is 6
  test_run 0..5 6
  # test for `./qs -n 5 -f -d` and validate ok`
  test_run 5 6
  # test for `./qs -n 0..5 -f -d` and validate ok`
  test_run 0..5 6
  # to test for failure
  set +e
  # test for `./qs -n s -f -d` and expected failed`
  test_run s 6
  # test for `./qs -n 0..5. -f -d` and expected failed`
  test_run 0..5. 6
  set -e
  echo "*** test result ***"
  cat ./test-$$-result
fi

