package model

import (
	"regexp"
	"strings"
	"testing"
)

// GitCommanderMock is mock of exuecuter
type GitCommanderMock struct{}

func TestAddReflogHash(t *testing.T) {

	gci := SetGitExecuter(GitCommanderMock{}) // use mock for unit test
	gci.AddReflogHash()

	got := len(gci.ReflogHashList)
	want := 6
	if got != want {
		t.Fatalf("want %v, but %v:", want, got)
	}
}

func (g GitCommanderMock) Reflog(opt string) ([]byte, error) {
	return []byte("a517e3b\na517e3b\n0b860a1\na517e3b\na517e3b\n0b860a1"), nil
}

func TestAddcommitMessage(t *testing.T) {
	/*
	  AddcommitMessage function append the following property
	  from the result of `git log --oneline --format=%s`.
	  - gci.CommitMsgList ............ original commit messages
	  - gci.CommitNewMsgList ......... commit messages inserted at the beginning of "fixup!"
	  - gci.CommitSpecifiedMsgList ... commit messages inserted at the beginning of "fixup!"
	                                   and replaced original message with specified message
	*/
	gci := SetGitExecuter(GitCommanderMock{}) // use mock for unit test
	gci.AddCommitMessage("SpecifiedMsg")

	got := strings.Join(gci.CommitMsgList, " ")
	reg := regexp.MustCompile(`(test\d+\s*){6}`)
	if !reg.MatchString(got) {
		t.Fatalf("regexp %v do not match %v:", reg, got)
	}

	got = strings.Join(gci.CommitNewMsgList, " ")
	reg = regexp.MustCompile(`(fixup!\stest\d+\s*){6}`)
	if !reg.MatchString(got) {
		t.Fatalf("regexp %v do not match %v:", reg, got)
	}

	got = strings.Join(gci.CommitSpecifiedMsgList, " ")
	reg = regexp.MustCompile(`(fixup!\sSpecifiedMsg\s*){6}`)
	if !reg.MatchString(got) {
		t.Fatalf("regexp %v do not match %v:", reg, got)
	}
}

func TestAddCommitHash(t *testing.T) {
	/*
		AddCommitHash inserts gci.CommitHashList
		from the result of `git log --oneline --format=%h`
	*/
	gci := SetGitExecuter(GitCommanderMock{}) // use mock for unit test
	gci.AddCommitHash()

	got := strings.Join(gci.CommitHashList, " ")
	reg := regexp.MustCompile(`([a-z0-9]{7}\s*){6}`)
	if !reg.MatchString(got) {
		t.Fatalf("regexp %v do not match %v:", reg, got)
	}
}

func (g GitCommanderMock) Commitlog(opts ...string) ([]byte, error) {
	if opts[1] == "--format=%s" { // for TestAddcommitMessage
		return []byte("test1\ntest2\ntest3\ntest4\ntest5\ntest6"), nil
	} else if opts[1] == "--format=%h" { // for TestAddcommitHash
		return []byte("1a2b3c4\nd5e6f7g\n8h9i0j1\nk2l3m4n\n5o6p7q8\nr9s0t1u"), nil
	} else {
		return nil, nil
	}
}
