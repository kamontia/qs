#!/bin/bash
#
# test script for qs
#
# USAGE: ./test.sh [setup]
# if you only create testdir, please set "setup" to arg

set -ex
set -o pipefail

readonly EXEC_COMMAND=$(basename $(pwd))
readonly ROOTDIR=$(pwd)

DIRNO=1

setup () {
  TESTDIR=test-"$$"-"${EXEC_COMMAND}"-"${DIRNO}"
  mkdir -p "$TESTDIR"
  go build
  cp ./"${EXEC_COMMAND}" "$TESTDIR"
  cd "$TESTDIR"

  git init
  git config --local user.email "git-fixup@example.com"
  git config --local user.name "git fixup"
  git commit --allow-empty -m "Initial commit"

  max=10
  for ((i=0; i <= "$max"; i++)); do
      echo file-"${i}" >> file-"${i}"
      git add file-"${i}"
      git commit -m "Add file-${i}"
  done
  echo "*** create $TESTDIR ***"
}

teardown () {
  cd "${ROOTDIR}"
  DIRNO=$(expr $DIRNO + 1)
}

test_squashed () {
  NUM="$2"
  if [[ "${NUM}" =~ ^([0-9]+)$ ]]; then
    EXPECTED_ADDED_FILE_NUM=$(( ${BASH_REMATCH[1]} + 1 ))
  elif [[ "${NUM}" =~ ^([0-9]+)\.\.([0-9]+)$ ]]; then
    EXPECTED_ADDED_FILE_NUM=$(( ${BASH_REMATCH[2]} + 1 ))
  else
    echo "invalid augument ${NUM}"
    exit 1
  fi

  setup

  git log --oneline

  ./"$EXEC_COMMAND" -n "$NUM" -f -d
  ret=$?

  if [[ "$ret" != "0" ]]; then
    echo "[failed] RUN ./$EXEC_COMMAND -n $NUM -f -d RESULT qs non-zero status code $ret" >> ./../test-$$-result
    return 1
  fi

  git log --oneline
  ADDED_FILE_NUM=$(git diff HEAD^..HEAD --name-only | wc -l | tr -d ' ')

  if [ "$ADDED_FILE_NUM" == "$EXPECTED_ADDED_FILE_NUM" ]; then
    echo "[passed] RUN ./$EXEC_COMMAND -n $NUM -f -d RESULT $ADDED_FILE_NUM" EXPECTED $EXPECTED_ADDED_FILE_NUM >> ./../test-$$-result
  else
    echo "[failed] RUN ./$EXEC_COMMAND -n $NUM -f -d RESULT $ADDED_FILE_NUM" EXPECTED $EXPECTED_ADDED_FILE_NUM >> ./../test-$$-result
  fi

  teardown
}

# https://github.com/kamontia/qs/pull/38
test_rebase_abort () {
   setup
   git revert HEAD~2 --no-edit
   echo file-11 >> file-11
   git add file-11
   git commit -m "Add file-11"

   set +e
   REFLOG_HASH_1=$(git log --oneline --format=%h|head -n1)
   ./"${EXEC_COMMAND}" -n 2..5 -f -d
   REFLOG_HASH_2=$(git log --oneline --format=%h|head -n 1)
   set -e
   if [[ "${REFLOG_HASH_1}" == "${REFLOG_HASH_2}" ]]; then
     echo "[passed] RUN ./${EXEC_COMMAND} -n 2..5 -f -d" >> ./../test-$$-result
   else
     echo "[failed] RUN ./${EXEC_COMMAND} -n 2..5 -f -d" >> ./../test-$$-result
   fi 
   teardown
}

# https://github.com/kamontia/qs/issues/17
test_message () {
  NUM="$2"
  MESSAGE="$6"
  if [[ "${NUM}" =~ ^([0-9]+)$ ]]; then
    TARGET=0
    PRETARGET=1
  elif [[ "${NUM}" =~ ^([0-9]+)\.\.([0-9]+)$ ]]; then
    TARGET="${BASH_REMATCH[1]}"
    PRETARGET=$(( ${BASH_REMATCH[1]} + 1 ))
  else
    echo "invalid augument ${NUM}"
    exit 1
  fi

  setup

  ./"$EXEC_COMMAND" -n "$NUM" -f -d -m "$MESSAGE"
  ret="$?"

  if [[ "$ret" != "0" ]]; then
    echo "[failed] RUN ./$EXEC_COMMAND -n $NUM -f -d RESULT qs non-zero status code $ret" >> ./../test-$$-result
    return 1
  fi

  ACTUAL_MESSAGE=$(git log HEAD~$PRETARGET..HEAD~$TARGET --oneline --format=%s)

  if [[ "$MESSAGE" == "$ACTUAL_MESSAGE" ]]; then
    echo "[passed] RUN ./$EXEC_COMMAND -n $NUM -f -d -m $MESSAGE RESULT $ACTUAL_MESSAGE EXPECTED $MESSAGE" >> ./../test-$$-result
  else
    echo "[failed] RUN ./$EXEC_COMMAND -n $NUM -f -d -m $MESSAGE RESULT $ACTUAL_MESSAGE EXPECTED $MESSAGE" >> ./../test-$$-result
  fi

  teardown
}

# https://github.com/kamontia/qs/issues/61
test_ls() {
  NUM="$2"
  if [[ "${NUM}" =~ ^([0-9]+)$ ]]; then
    EXPECTED="$NUM"
  elif [[ ${NUM} =~ ^([0-9]+)\.\.([0-9]+)$ ]]; then
    EXPECTED=$(( ${BASH_REMATCH[2]} - ${BASH_REMATCH[1]} ))
  else
    echo "invalid augument ${NUM}"
    exit 1
  fi

  setup

  RESULT="$(./$EXEC_COMMAND ls -n $NUM)"
  ret="$?"

  if [[ "$ret" != "0" ]]; then
    echo "[failed] RUN ./$EXEC_COMMAND -n $NUM -f -d RESULT qs non-zero status code $ret" >> ./../test-$$-result
    return 1
  fi

  SQUASHED_COMMITS=$(echo "$RESULT" | grep -c "squash")

  if [[ "$SQUASHED_COMMITS" == "$EXPECTED" ]]; then
    echo "[passed] RUN ./$EXEC_COMMAND ls -n $NUM RESULT $SQUASHED_COMMITS EXPECTED $EXPECTED" >> ./../test-$$-result
  else
    echo "[failed] RUN ./$EXEC_COMMAND ls -n $NUM RESULT $SQUASHED_COMMITS EXPECTED $EXPECTED" >> ./../test-$$-result
  fi

  teardown
}

# https://github.com/kamontia/qs/issues/71
test_range_validation() {
  FLAGS="$1 $2"
  RANGE="$3"
  MESSAGE="$4" # Expect to finish leaving this message(test target)
  VALIDATION_FILE=$(mktemp tmp-XXXXX)  # Temporary file to validate this test.
  VALIDATION_FILE_ABS_PATH=$(pwd)/${VALIDATION_FILE}
  trap 'rm -v ${VALIDATION_FILE_ABS_PATH}; exit 1' 1 2 3 15
  
  setup

  set +e
  ./"$EXEC_COMMAND" ${FLAGS} ${RANGE} | tee ${VALIDATION_FILE_ABS_PATH}
  RESULT=$(grep -c "${MESSAGE}" ${VALIDATION_FILE_ABS_PATH})
  set -e
  rm ${VALIDATION_FILE_ABS_PATH}
  if [[ ${RESULT} -ne 0 ]]; then # MESSAGE matches
    echo "[passed] ./"$EXEC_COMMAND" ${FLAGS} ${RANGE} EXPECTED MESSAGE ${MESSAGE}" >> ./../test-$$-result
  else # Message not matches
    echo "[failed] ./"$EXEC_COMMAND" ${FLAGS} ${RANGE} EXPECTED MESSAGE ${MESSAGE}" >> ./../test-$$-result
  fi
  teardown
}

# https://github.com/kamontia/qs/issues/59
test_signal_handling() {
  FLAGS="$1 $2"
  RANGE="$3"
  MESSAGE="$4" # Expect to finish leaving this message(test target)
  SIGNAL="$5" # 2:SIGINT 15:SIGTERM
  VALIDATION_FILE=$(mktemp tmp-XXXXX)  # Temporary file to validate this test.
  VALIDATION_FILE_ABS_PATH=$(pwd)/${VALIDATION_FILE}
  
  setup
  set +e
  REFLOG_HASH_1=$(git log --oneline --format=%h|head -n 1)
  ./"$EXEC_COMMAND" -f -n 1..10 > "${VALIDATION_FILE_ABS_PATH}" &
  sleep 1 # Wait for the qs command to be executed
  kill -${SIGNAL} $!
  sleep 1 # Wait for qs command to catch sigint
  RESULT=$(grep -c "${MESSAGE}" ${VALIDATION_FILE_ABS_PATH})
  REFLOG_HASH_2=$(git log --oneline --format=%h|head -n 1)
  rm ${VALIDATION_FILE_ABS_PATH}
  set -e
  if [[ ${RESULT} -ne 0 ]] && [[ "${REFLOG_HASH_1}==${REFLOG_HASH_2}" ]]; then # MESSAGE matches
    echo "[passed] ./"$EXEC_COMMAND" ${FLAGS} ${RANGE} EXPECTED MESSAGE ${MESSAGE}" >> ./../test-$$-result
  else # Message not matches
    echo "[failed] ./"$EXEC_COMMAND" ${FLAGS} ${RANGE} EXPECTED MESSAGE ${MESSAGE}" >> ./../test-$$-result
  fi
  teardown
}

main() {
  test_squashed -n 5 -f -d
  test_squashed -n 0..5 -f -d
  test_rebase_abort
  test_message -n 5 -f -d -m "test message"
  test_message -n 3..5 -f -d -m "test message"
  test_ls -n 5
  test_ls -n 3..5
  test_range_validation -f -n 9..11 "The first commit is included in the specified range."
  test_range_validation -f -n 9..12 "QS cannot rebase out of range."
  test_signal_handling -f -n 1..10 "Completed. Please rebase manually." 2

  echo "*** test result ***"
  cat ./test-"$$"-result
  ! grep 'failed' test-"$$"-result
}

if [[ "$1" == setup ]]; then
  setup
else
  main "$@"
fi
