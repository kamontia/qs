package model

// Executer is interface of git command
type Executer interface {
	Reflog(opts ...string) ([]byte, error)
	Commitlog(opts ...string) ([]byte, error)
}

type GlobalInfo struct {
	/* Reference to the implementation object for execute the command(Production/Mock) */
	GitExecuter    Executer
	GitInfo        []CommitInfo
	RangeArray     string
	StartHeadIndex int
	EndHeadIndex   int
}

// CommitInfo is git information struct
type CommitInfo struct {
	CommitHashList   string
	ReflogHashList   string
	CommitMessageRef CommitMessage
}

// CommitMessage is
type CommitMessage struct {
	Subject                string
	CommitNewMsgList       string
	CommitSpecifiedMsgList string
	Body                   []string
}
