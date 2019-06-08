package model

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/kamontia/qs/utils"
	log "github.com/sirupsen/logrus"
)

const ADJUST_LINE_NUMBER = 2

// SetGitExecuter is constructor for GitExecutor
func SetGitExecuter(e Executer) *GlobalInfo {
	return &GlobalInfo{
		GitExecuter: e,
	}
}

// AddReflogHash collect reflogHash, then append to ReflogHashList
func (g *GlobalInfo) AddReflogHash() {
	/* Reference to the implementation object for execute the command(Production/Mock) */
	// number := g.StartHeadIndex - g.EndHeadIndex + 1
	countNum := fmt.Sprintf("%d", g.StartHeadIndex-g.EndHeadIndex+1)
	out, err := g.GitExecuter.Reflog("--format=%h", "-n", countNum)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	for i, v := range regexp.MustCompile("\r\n|\n|\r").Split(string(out), -1) {
		if v != "" {
			g.GitInfo[i].ReflogHashList = v
		}
	}
}

// AddCommitMessage insert commit messages
func (g *GlobalInfo) AddCommitMessage(msg string) {
	for i := 0; i < len(g.GitInfo); i++ {
		headIndex := fmt.Sprintf("HEAD~%d..HEAD~%d", i+1, i)
		subject, sErr := g.GitExecuter.Commitlog("--pretty=format:%s", headIndex, "--no-decorate")
		body, bErr := g.GitExecuter.Commitlog("--pretty=format:%b", headIndex, "--no-decorate")

		if sErr != nil {
			log.Error(sErr)
			os.Exit(1)
		}
		if bErr != nil {
			log.Error(bErr)
			os.Exit(1)
		}

		g.GitInfo[i].CommitMessageRef.Subject = string(subject)
		g.GitInfo[i].CommitMessageRef.CommitNewMsgList = fmt.Sprintf("fixup! %s", string(subject))
		g.GitInfo[i].CommitMessageRef.CommitSpecifiedMsgList = fmt.Sprintf("fixup! %s", string(msg))

		// Count newlines
		bodyLines := strings.Count(string(body), "\r\n|\n|\r")
		g.GitInfo[i].CommitMessageRef.Body = make([]string, bodyLines)

		for _, v := range regexp.MustCompile("\r\n|\n|\r").Split(string(body), -1) {
			if v != "" {
				g.GitInfo[i].CommitMessageRef.Body = append(g.GitInfo[i].CommitMessageRef.Body, string(v))
			}
		}
	}

}

func (g *GlobalInfo) MakeMesasgeOption(index, target int, squashMessage string) bool {
	file, err := os.OpenFile(".qstemp", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Error("FileOpenError:", err)
	}

	defer file.Close()

	var message string
	if g.GitInfo[g.StartHeadIndex].CommitMessageRef.Subject != "" {
		if index == target { // リベースの最終コミット
			/*
				HEAD~N	: target message          <- index == target
				HEAD~N-1: fixup! target message
				...
				HEAD~0	:
			*/
			if squashMessage == "" {
				message = fmt.Sprintf("%s", g.GitInfo[target].CommitMessageRef.Subject)
			} else {
				message = fmt.Sprintf("%s", squashMessage)
			}
		} else { // リベースでまとめられる(消える)コミット
			if squashMessage == "" {
				message = fmt.Sprintf("fixup! %s", g.GitInfo[target].CommitMessageRef.Subject)
			} else {
				message = fmt.Sprintf("fixup! %s", squashMessage)
			}
		}
	} else {
		// allow empty commit
		message = fmt.Sprintf("%s", "")
	}

	if len(g.GitInfo[index].CommitMessageRef.Body) != 0 {
		message += fmt.Sprintf("\n\n")
	}

	for _, v := range g.GitInfo[index].CommitMessageRef.Body {
		message += fmt.Sprintf("%s\n", v)
	}
	fmt.Fprintln(file, message)
	return true
}

// AddCommitHash insert commit hash
func (g *GlobalInfo) AddCommitHash() {
	countNum := fmt.Sprintf("%d", g.StartHeadIndex-g.EndHeadIndex+1)
	out, err := g.GitExecuter.Commitlog("--format=%h", "-n", countNum)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	for i, v := range regexp.MustCompile("\r\n|\n|\r").Split(string(out), -1) {
		if v != "" {
			g.GitInfo[i].CommitHashList = v
		}
	}
}

func (gci *GlobalInfo) DisplayCommitHashAndMessage(beginNumber int, endNumber int) {
	/* Display commit hash and message. The [pickup|..] strings is colored */

	for i := endNumber; i >= beginNumber; i-- {
		/* Switch output corresponded to do squash */
		if utils.NeedsChangeMessage(i, beginNumber, endNumber) {
			if runtime.GOOS == "windows" {
				log.Infof("[%2d]\tpickup -> squash %s %s", i, gci.GitInfo[i].CommitHashList, gci.GitInfo[i].CommitMessageRef.Subject)
			} else {
				log.Infof("[%2d]\t\x1b[35mpickup\x1b[0m -> \x1b[36msquash \x1b[0m %s %s", i, gci.GitInfo[i].CommitHashList, gci.GitInfo[i].CommitMessageRef.Subject)
			}
		} else {
			if runtime.GOOS == "windows" {
				log.Infof("[%2d]\tpickup -> pickup %s %s", i, gci.GitInfo[i].CommitHashList, gci.GitInfo[i].CommitMessageRef.Subject)
			} else {
				log.Infof("[%2d]\t\x1b[35mpickup\x1b[0m -> \x1b[35mpickup \x1b[0m %s %s", i, gci.GitInfo[i].CommitHashList, gci.GitInfo[i].CommitMessageRef.Subject)
			}
		}
	}
}

func CheckCurrentCommit(f bool, beginNumber int, endNumber int, gci *GlobalInfo, specifiedMsg string) {
	var force = f
	var sNum = strconv.Itoa(beginNumber)
	gci.AddReflogHash()
	gci.AddCommitHash()
	gci.AddCommitMessage(specifiedMsg)

	headMax := len(gci.GitInfo) - 2

	/*
		$ qs -n 1..3
			HEAD~0:
			HEAD~1: beginNumber
			HEAD~2:
			HEAD~3: headMax
			HEAD~4:
	*/
	if !utils.RangeValidation(beginNumber, headMax) {
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

func DoRecovery(doneCh chan struct{}, gci *GlobalInfo) {
	log.Error("Error. QS try to recovery...")
	cmd := exec.Command("git", "rebase", "--abort")
	if err := cmd.Run(); err != nil {
		log.Error(err)
	}
	cmd = exec.Command("git", "reset", "--mixed", gci.GitInfo[0].ReflogHashList)
	if err := cmd.Run(); err != nil {
		log.Error(err)
	}
	doneCh <- struct{}{}
	log.Error("Completed. Please rebase manually.")
	os.Exit(1)
}

func SignalErorrHandling(gci *GlobalInfo) {
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
func StartSignalHandling(gci *GlobalInfo) (sync.WaitGroup, chan struct{}, chan os.Signal) {
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

	// fmt.Println(wg, doneCh, sigCh)
	return wg, doneCh, sigCh
}

// StopSignalHandling is topping signal handling
func StopSignalHandling(wg sync.WaitGroup, doneCh chan struct{}, sigCh chan os.Signal) {
	log.Debug("StopSingalHandling")
	//fmt.Println(wg, doneCh, sigCh)
	doneCh <- struct{}{}
}
