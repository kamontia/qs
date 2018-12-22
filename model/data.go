package model

import (
	"os"
	"regexp"

	log "github.com/sirupsen/logrus"
)

// Executer is interface of git command
type Executer interface {
	Reflog(opt string) ([]byte, error)
}

// GitCommitInfo is git information struct
type GitCommitInfo struct {
	RangeArray             []string
	CommitHashList         []string
	CommitMsgList          []string
	CommitNewMsgList       []string
	CommitSpecifiedMsgList []string
	ReflogHashList         []string
	/* コマンドを実行するための実装オブジェクトへの参照	*/
	GitExecuter Executer
}

// NewGitCommitInfo is constructor
func NewGitCommitInfo(e Executer) *GitCommitInfo {
	return &GitCommitInfo{
		GitExecuter: e,
	}
}

// AddReflogHash collect reflogHash, then append to REflogHashList
func (g *GitCommitInfo) AddReflogHash() {
	/* GitExecuterは NewGitCommtiInfoで初期化されており、本番向けかテスト向けの実装を参照する */
	out, err := g.GitExecuter.Reflog("--format=%h")
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	for _, v := range regexp.MustCompile("\r\n|\n|\r").Split(string(out), -1) {
		g.ReflogHashList = append(g.ReflogHashList, v)
	}
}
