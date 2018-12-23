package model

import (
	"fmt"
	"os"
	"regexp"

	log "github.com/sirupsen/logrus"
)

// Executer is interface of git command
type Executer interface {
	Reflog(opt string) ([]byte, error)
	Commitlog(opts ...string) ([]byte, error)
}

// GitCommitInfo is git information struct
type GitCommitInfo struct {
	RangeArray             []string
	CommitHashList         []string
	CommitMsgList          []string
	CommitNewMsgList       []string
	CommitSpecifiedMsgList []string
	ReflogHashList         []string
	/* Reference to the implementation object for execute the command(Production/Mock) */
	GitExecuter Executer
}

// SetGitExecuter is constructor for GitExecutor
func SetGitExecuter(e Executer) *GitCommitInfo {
	return &GitCommitInfo{
		GitExecuter: e,
	}
}

// AddReflogHash collect reflogHash, then append to ReflogHashList
func (g *GitCommitInfo) AddReflogHash() {
	/* Reference to the implementation object for execute the command(Production/Mock) */
	out, err := g.GitExecuter.Reflog("--format=%h")
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	for _, v := range regexp.MustCompile("\r\n|\n|\r").Split(string(out), -1) {
		g.ReflogHashList = append(g.ReflogHashList, v)
	}
}

// AddCommitMessage insert commit messages
func (g *GitCommitInfo) AddCommitMessage(msg string) {
	out, err := g.GitExecuter.Commitlog("--oneline", "--format=%s")
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	for _, v := range regexp.MustCompile("\r\n|\n|\r").Split(string(out), -1) {
		g.CommitMsgList = append(g.CommitMsgList, v)
		g.CommitNewMsgList = append(g.CommitNewMsgList, fmt.Sprintf("fixup! %s", v))
		g.CommitSpecifiedMsgList = append(g.CommitSpecifiedMsgList, fmt.Sprintf("fixup! %s", msg))
	}
}

// AddCommitHash insert commit hash
func (g *GitCommitInfo) AddCommitHash() int {
	out, err := g.GitExecuter.Commitlog("--oneline", "--format=%h")
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	for _, v := range regexp.MustCompile("\r\n|\n|\r").Split(string(out), -1) {
		g.CommitHashList = append(g.CommitHashList, v)
	}

	headMax := len(g.CommitHashList) - 2
	return headMax
}
