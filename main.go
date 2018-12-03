package main

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

	"github.com/kamontia/qs/model"
	colorable "github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

/* Definition */
var stdin string
var beginNumber int
var endNumber int
var headMax int
var specifiedMsg string

func logrusInit(d bool) {
	var debug = d
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
	log.SetOutput(colorable.NewColorableStdout())
	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func validate(n string) {
	r := regexp.MustCompile(`^[0-9]+$|^[0-9]+..[0-9]+$`)
	isNum := r.MatchString(n)
	if !isNum {
		log.Error("invalid number flag")
		os.Exit(1)
	}
}

func displayCommitHashAndMessage(gci *model.GitCommitInfo) {
	/* Display commit hash and message. The [pickup|..] strings is colored */

	/* set limit to display the history */
	var limit int
	if beginNumber+2 > len(gci.CommitMsgList)-2 {
		limit = len(gci.CommitMsgList) - 2
	} else {
		limit = beginNumber + 2
	}

	for i := limit; i >= 0; i-- {
		/* Switch output corresponded to do squash */
		if needsChangeMessage(i, beginNumber, endNumber) {
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

func rangeValidation() {
	if beginNumber > headMax {
		log.Error("QS cannot rebase out of range.")
		os.Exit(1)
	} else if beginNumber == headMax {
		log.Error("The first commit is included in the specified range. If necessary, please rebase with --root option manually.")
		os.Exit(1)
	} else {
		log.Debug("Specified range is OK.")
	}
}

func getCommitHash(gci *model.GitCommitInfo) {
	out, err := exec.Command("git", "log", "--oneline", "--format=%h").Output()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	for _, v := range regexp.MustCompile("\r\n|\n|\r").Split(string(out), -1) {
		gci.CommitHashList = append(gci.CommitHashList, v)
	}

	headMax = len(gci.CommitHashList) - 2
}

func getReflogHash(gci *model.GitCommitInfo) {
	out, err := exec.Command("git", "reflog", "--format=%h").Output()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	for _, v := range regexp.MustCompile("\r\n|\n|\r").Split(string(out), -1) {
		gci.ReflogHashList = append(gci.ReflogHashList, v)
	}
}

func getCommitMessage(gci *model.GitCommitInfo) {
	out, err := exec.Command("git", "log", "--oneline", "--format=%s").Output()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	for _, v := range regexp.MustCompile("\r\n|\n|\r").Split(string(out), -1) {
		gci.CommitMsgList = append(gci.CommitMsgList, v)
		gci.CommitNewMsgList = append(gci.CommitNewMsgList, fmt.Sprintf("fixup! %s", v))
		gci.CommitSpecifiedMsgList = append(gci.CommitSpecifiedMsgList, fmt.Sprintf("fixup! %s", specifiedMsg))
	}
}

func checkCurrentCommit(f bool, beginNumber int, endNumber int, gci *model.GitCommitInfo) {
	var force = f
	var sNum = strconv.Itoa(beginNumber)

	getCommitHash(gci)
	getReflogHash(gci)
	getCommitMessage(gci)
	rangeValidation()

	if force {
		log.Debug("force update")
	} else {
		displayCommitHashAndMessage(gci)

		fmt.Println("Do you squash the above commits?(y/n)")
		out, err := exec.Command("git", "log", "--oneline", "-n", sNum).Output()
		if err != nil {
			log.Error(out)
			os.Exit(1)
		}
		for {
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

func pickupSquashRange(num string) (int, int){
	/* TODO: Check error strictly */
	var error error

	if strings.Contains(num, "..") {
		/* Specify the range you aggregate */
		rangeArray := strings.Split(num, "..")
		beginNumber, error = strconv.Atoi(rangeArray[0])
		if error != nil {
			log.Error(error)
			os.Exit(1)
		}
		endNumber, error = strconv.Atoi(rangeArray[1])
		if error != nil {
			log.Error(error)
			os.Exit(1)
		}
		if beginNumber < endNumber {
			tmp := beginNumber
			beginNumber = endNumber
			endNumber = tmp
		}
	} else {
		/* Specify the one parameter you aggregate */
		endNumber = 0
		beginNumber, error = strconv.Atoi(num)
		if error != nil {
			log.Error(error)
		}
	}
  return beginNumber, endNumber
}

func needsChangeMessage(i int, begin int, end int) bool {
	if begin > i && i >= end {
		return true
	} else {
		return false
	}
}

func doRecovery(doneCh chan struct{}, gci *model.GitCommitInfo) {
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

func main() {

	app := cli.NewApp()
	app.Name = Name
	app.Version = Version
	app.Author = "Tatsuya Kamohara<kamontia@gmail.com>\n   Takeshi Kondo<take.she12@gmail.com>"
	app.Email = ""
	app.Usage = ""

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "number, n",
			Usage: "Specify suqash number",
		},
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "Force update",
		},
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "Show verbose logging",
		},
		cli.StringFlag{
			Name:  "message, m",
			Usage: "Commit message",
		},
	}
	app.Commands = Commands
	app.CommandNotFound = CommandNotFound

	app.Action = func(c *cli.Context) error {
		/* Create GitCommit Info */
		gci := new(model.GitCommitInfo)

		validate(c.String("number"))
		specifiedMsg = c.String("message")

		beginNumber, endNumber := pickupSquashRange(c.String("number"))
		logrusInit(c.Bool("debug"))
		checkCurrentCommit(c.Bool("force"), beginNumber, endNumber, gci)

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
						doRecovery(doneCh, gci)
					}

				case <-doneCh:
					return
				}
			}
		}()

		/* Suppress vim editor launching */
		os.Setenv("GIT_EDITOR", ":")

		for i := beginNumber; i >= 0; i-- {
			specifiedHead := fmt.Sprintf("HEAD~%d", i+1)
			var specifiedExec string

			if specifiedMsg != "" {
				if beginNumber == i {
					specifiedExec = fmt.Sprintf("--exec=git commit --amend -m\"%s\"", specifiedMsg)
				} else if needsChangeMessage(i, beginNumber, endNumber) {
					specifiedExec = fmt.Sprintf("--exec=git commit --amend -m\"%s\"", gci.CommitSpecifiedMsgList[beginNumber])
				} else {
					specifiedExec = fmt.Sprintf("--exec=git commit --amend -m\"%s\"", gci.CommitMsgList[i])
				}
			} else {
				if needsChangeMessage(i, beginNumber, endNumber) {
					specifiedExec = fmt.Sprintf("--exec=git commit --amend -m\"%s\"", gci.CommitNewMsgList[beginNumber])
				} else {
					specifiedExec = fmt.Sprintf("--exec=git commit --amend -m\"%s\"", gci.CommitMsgList[i])
				}
			}

			cmd := exec.Command("git", "rebase", specifiedHead, specifiedExec)
			log.Debugf("git rebase HEAD~%d %s\n", i, specifiedExec)

			if c.Bool("debug") {
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
			} else {
				cmd.Stdin = nil
				cmd.Stdout = nil
				cmd.Stderr = nil
			}
			if err := cmd.Run(); err != nil {
				doRecovery(doneCh, gci)
			}
		}

		/* git rebase with autosquash option */
		specifiedHead := fmt.Sprintf("HEAD~%d", beginNumber+1)
		cmd := exec.Command("git", "rebase", "-i", "--autosquash", "--autostash", specifiedHead, "--quiet", "--preserve-merges")

		/* Transfer the command I/O to Standard I/O */
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			doRecovery(doneCh, gci)
		}
		log.Debug("rebase completed")

		/* Stop gorutine */
		doneCh <- struct{}{}
		wg.Wait()

		return nil
	}
	app.Run(os.Args)
}
