package model

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"syscall"

	"github.com/kamontia/qs/utils"
	log "github.com/sirupsen/logrus"
)

const ADJUST_LINE_NUMBER = 2

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
func (g *GitCommitInfo) AddCommitHash() {
	out, err := g.GitExecuter.Commitlog("--oneline", "--format=%h")
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	for _, v := range regexp.MustCompile("\r\n|\n|\r").Split(string(out), -1) {
		g.CommitHashList = append(g.CommitHashList, v)
	}
}

func (gci *GitCommitInfo) DisplayCommitHashAndMessage(beginNumber int, endNumber int) {
	/* Display commit hash and message. The [pickup|..] strings is colored */

	/* set limit to display the history */
	var limit int
	if beginNumber+2 > len(gci.CommitMsgList)-ADJUST_LINE_NUMBER {
		limit = len(gci.CommitMsgList) - 2
	} else {
		limit = beginNumber + 2
	}

	for i := limit; i >= 0; i-- {
		/* Switch output corresponded to do squash */
		if utils.NeedsChangeMessage(i, beginNumber, endNumber) {
			if runtime.GOOS == "windows" {
				log.Infof("[%2d]\tpickup -> squash %s %s", i, gci.CommitHashList[i], gci.CommitMsgList[i])
			} else {
				log.Infof("[%2d]\t\x1b[35mpickup\x1b[0m -> \x1b[36msquash \x1b[0m %s %s", i, gci.CommitHashList[i], gci.CommitMsgList[i])
			}
		} else {
			if runtime.GOOS == "windows" {
				log.Infof("[%2d]\tpickup -> pickup %s %s", i, gci.CommitHashList[i], gci.CommitMsgList[i])
			} else {
				log.Infof("[%2d]\t\x1b[35mpickup\x1b[0m -> \x1b[35mpickup \x1b[0m %s %s", i, gci.CommitHashList[i], gci.CommitMsgList[i])
			}
		}
	}
}

func CheckCurrentCommit(f bool, beginNumber int, endNumber int, gci *GitCommitInfo, specifiedMsg string) {
	var force = f
	var sNum = strconv.Itoa(beginNumber)

	gci.AddReflogHash()
	gci.AddCommitHash()
	gci.AddCommitMessage(specifiedMsg)

	headMax := len(gci.CommitHashList) - 2
	if !utils.RangeValidation(headMax, beginNumber) {
		log.Error("Range validagion is failed")
		os.Exit(1)
	}

	if force {
		log.Debug("force update")
	} else {
		gci.DisplayCommitHashAndMessage(beginNumber, endNumber)

		fmt.Println("Do you squash the above commits?(y/n)")
		out, err := exec.Command("git", "log", "--oneline", "-n", sNum).Output()
		if err != nil {
			log.Error(out)
			os.Exit(1)
		}
		for {
			var stdin string
			fmt.Scan(&stdin)
			switch stdin {
			case "y":
				log.Debug("Fixup!")
			case "n":
				log.Debug("Abort!")
				os.Exit(1)
			default:
				log.Debug("You can input y or n")
				continue
			}
			break
		}
	}
}

func DoRecovery(doneCh chan struct{}, gci *GitCommitInfo) {
	log.Error("Error. QS try to recovery...")
	cmd := exec.Command("git", "rebase", "--abort")
	if err := cmd.Run(); err != nil {
		log.Error(err)
	}
	cmd = exec.Command("git", "reset", "--hard", gci.ReflogHashList[0])
	if err := cmd.Run(); err != nil {
		log.Error(err)
	}
	doneCh <- struct{}{}
	log.Error("Completed. Please rebase manually.")
	os.Exit(1)
}

func SignalErorrHandling(gci *GitCommitInfo) {
	/* Create thread for handling signal */
	wg := sync.WaitGroup{}
	doneCh := make(chan struct{}, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh,
			syscall.SIGTERM,
			syscall.SIGINT,
			os.Interrupt) /* For Windows */

		defer signal.Stop(sigCh)

		for {
			select {
			case sig := <-sigCh:
				switch sig {
				case syscall.SIGTERM, syscall.SIGINT, os.Interrupt:
					log.Info("Catch signal.QS try to recorvery.")
					DoRecovery(doneCh, gci)
				}

			case <-doneCh:
				return
			}
		}
	}()
}

// StartSignalHandling is starting signal handling
func StartSignalHandling(gci *GitCommitInfo) (sync.WaitGroup, chan struct{}, chan os.Signal) {
	log.Debug("StartSignalHandling")
	/* Create thread for handling signal */
	wg := sync.WaitGroup{}
	doneCh := make(chan struct{}, 1)
	wg.Add(1)
	sigCh := make(chan os.Signal, 1)
	go func() {
		defer wg.Done()

		signal.Notify(sigCh,
			syscall.SIGTERM,
			syscall.SIGINT,
			os.Interrupt) /* For Windows */

		// Maybe deleted
		defer signal.Stop(sigCh)

		for {
			select {
			case sig := <-sigCh:
				switch sig {
				case syscall.SIGTERM, syscall.SIGINT, os.Interrupt:
					log.Info("Catch signal.QS try to recorvery.")
					DoRecovery(doneCh, gci)
				}

			case <-doneCh:
				return
			}
		}
	}()

	fmt.Println(wg, doneCh, sigCh)
	return wg, doneCh, sigCh
}

// StopSignalHandling is topping signal handling
func StopSignalHandling(wg sync.WaitGroup, doneCh chan struct{}, sigCh chan os.Signal) {
	log.Debug("StopSingalHandling")
	fmt.Println(wg, doneCh, sigCh)
	doneCh <- struct{}{}
}
