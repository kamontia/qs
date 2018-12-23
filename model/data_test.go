package model

import (
	"testing"
)

func TestAddReflogHash(t *testing.T) {

	gci := NewGitCommitInfo(testExecuter{}) // use mock for unit test
	gci.AddReflogHash()

	got := len(gci.ReflogHashList)
	want := 6
	if got != want {
		t.Fatalf("want %v, but %v:", want, got)
	}
}

// testExecuter is mock of exuecuter
type testExecuter struct{}

func (g testExecuter) Reflog(opt string) ([]byte, error) {
	return []byte("a517e3b\na517e3b\n0b860a1\na517e3b\na517e3b\n0b860a1"), nil
}
