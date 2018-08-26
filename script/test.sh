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
      echo file-${i} >> file-${i}
      git add file-${i}
      git commit -m "Add file-${i}"
  done
}

teardown () {
  cd ${ROOTDIR}
  DIRNO=$(expr $DIRNO + 1)
}

test_squashed () {
  NUM=$2
  if [[ ${NUM} =~ ^([0-9]+)$ ]]; then
    EXPECTED_ADDED_FILE_NUM=$(( ${BASH_REMATCH[1]} + 1 ))
  elif [[ ${NUM} =~ ^([0-9]+)\.\.([0-9]+)$ ]]; then
    EXPECTED_ADDED_FILE_NUM=$(( ${BASH_REMATCH[2]} + 1 ))
  else
    echo "invalid augument ${NUM}"
    exit 1
  fi

  prepare_env
  git_init
  git_pre_commit

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
    echo "[passed] RUN ./$ExecComamnd -n $NUM -f -d RESULT $ADDED_FILE_NUM" EXPECTED $EXPECTED_ADDED_FILE_NUM >> ./../test-$$-result
  else
    echo "[failed] RUN ./$ExecComamnd -n $NUM -f -d RESULT $ADDED_FILE_NUM" EXPECTED EXPECTED_ADDED_FILE_NUM >> ./../test-$$-result
  fi

  teardown
}

test_PR38 () { # PR38: Feature error handling by git rebase --abort
   prepare_env
   git_init
   git_pre_commit
   git revert HEAD~2 --no-edit
   echo file-11 >> file-11
   git add file-11
   git commit -m "Add file-11"

   set +e
   REFLOG_HASH_1=$(git log --oneline --format=%h|head -n1)
   ./${ExecComamnd}  -n 2..5 -f -d
   REFLOG_HASH_2=$(git log --oneline --format=%h|head -n 1)
   set -e
   if [ "${REFLOG_HASH_1}" == "${REFLOG_HASH_2}" ]; then
     echo "[passed] RUN ./${ExecComamnd} -n 2..5 -f -d" >> ./../test-$$-result
   else
     echo "[failed] RUN ./${ExecComamnd} -n 2..5 -f -d" >> ./../test-$$-result
   fi 
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
  test_squashed -n 5 -f -d
  test_squashed -n 0..5 -f -d
  test_PR38
  echo "*** test result ***"
  cat ./test-$$-result
  ! grep 'failed' test-$$-result
fi

