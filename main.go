package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	log "github.com/sirupsen/logrus"
)

/* Definisiotn */
var stdin string
var beginNumber int
var endNumber int
var rangeArray []string
var commitHashList []string
var commitMsgList []string
var commitNewMsgList []string
var commitSpecifiedMsgList []string
var reflogHashList []string
var specifiedMsg string

func logrusInit(d bool) {
	var debug = d
	log.SetOutput(os.Stdout)
	if debug {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.WarnLevel)
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

func displayCommitHashAndMessage() {
	/* Display commit hash and message. The [pickup|..] strings is colored */
	for i := len(commitMsgList) - 2; i >= 0; i-- {
		/* Switch output corresponded to do squash */
		if needsChangeMessage(i) {
			log.Warnf("[%2d]\t\x1b[35mpickup\x1b[0m -> \x1b[36msquash \x1b[0m %s %s", i, commitHashList[i], commitMsgList[i])
		} else {
			log.Warnf("[%2d]\t\x1b[35mpickup\x1b[0m -> \x1b[35mpickup \x1b[0m %s %s", i, commitHashList[i], commitMsgList[i])
		}
	}
	/* (END) Display commit hash and message */
}

func getCommitHash() {
	/* Get commit hash */
	out, err := exec.Command("git", "log", "--oneline", "--format=%h").Output()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	for _, v := range regexp.MustCompile("\r\n|\n|\r").Split(string(out), -1) {
		commitHashList = append(commitHashList, v)
	}
	/* (END)Get commit hash */
}

func getCommitMessage() {
	/* Get commit message */
	out, err := exec.Command("git", "log", "--oneline", "--format=%s").Output()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	for _, v := range regexp.MustCompile("\r\n|\n|\r").Split(string(out), -1) {
		commitMsgList = append(commitMsgList, v)
		commitNewMsgList = append(commitNewMsgList, fmt.Sprintf("fixup! %s", v))
		commitSpecifiedMsgList = append(commitSpecifiedMsgList, fmt.Sprintf("fixup! %s", specifiedMsg))
	}
	/* (END)Get commit message */
}

func checkCurrentCommit(f bool, beginNumber int, endNumber int) {
	var force = f
	var sNum = strconv.Itoa(beginNumber)
	log.SetOutput(os.Stdout)

	getCommitHash()

	/* Get reflog hash */
	out, err := exec.Command("git", "reflog", "--format=%h").Output()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	for _, v := range regexp.MustCompile("\r\n|\n|\r").Split(string(out), -1) {
		reflogHashList = append(reflogHashList, v)
	}
	/* (END)Get reflog hash */

	getCommitMessage()

	if force {
		log.Info("force update")
	} else {
		displayCommitHashAndMessage()

		fmt.Println("Do you squash the above commits?(y/n)")
		out, err = exec.Command("git", "log", "--oneline", "-n", sNum).Output()
		if err != nil {
			log.Error(out)
			os.Exit(1)
		}
		for {
			fmt.Scan(&stdin)
			switch stdin {
			case "y":
				log.Info("Fixup!")
			case "n":
				log.Info("Abort!")
				os.Exit(1)
			default:
				log.Info("You can input y or n")
				continue
			}
			break
		}
	}
}

func pickupSquashRange(num string) {
	/* Pick up squash range */
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
	/* (END) Pick up squash range */

}

func needsChangeMessage(i int) bool {
	if beginNumber > i && i >= endNumber {
		return true
	} else {
		return false
	}
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
		validate(c.String("number"))
		specifiedMsg = c.String("message")

		pickupSquashRange(c.String("number"))
		logrusInit(c.Bool("debug"))
		checkCurrentCommit(c.Bool("force"), beginNumber, endNumber)

		// Parse number(--number, -n) parameter

		/* (WIP) git rebase */
		/**
		git rebase HEAD~N --exec="git commit -m"squash! commit messages" "
		*/
		/* Suppress vim editor launching */
		os.Setenv("GIT_EDITOR", ":")
		/* (END) Suppress vim editor launching */
		for i := beginNumber; i >= 0; i-- {
			speciedHead := fmt.Sprintf("HEAD~%d", i+1)
			var speciedExec string

			if specifiedMsg != "" {
				if beginNumber == i {
					speciedExec = fmt.Sprintf("--exec=git commit --amend -m\"%s\"", specifiedMsg)
				} else if needsChangeMessage(i) {
					speciedExec = fmt.Sprintf("--exec=git commit --amend -m\"%s\"", commitSpecifiedMsgList[beginNumber])
				} else {
					speciedExec = fmt.Sprintf("--exec=git commit --amend -m\"%s\"", commitMsgList[i])
				}
			} else {
				if needsChangeMessage(i) {
					speciedExec = fmt.Sprintf("--exec=git commit --amend -m\"%s\"", commitNewMsgList[beginNumber])
				} else {
					speciedExec = fmt.Sprintf("--exec=git commit --amend -m\"%s\"", commitMsgList[i])
				}
			}

			cmd := exec.Command("git", "rebase", speciedHead, speciedExec)
			log.Printf("git rebase HEAD~%d %s\n", i, speciedExec)

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
				log.Error("QS reseted this operation. Please git rebase manually")
				log.Error(err)
				cmd = exec.Command("git", "rebase", "--abort")
				if err := cmd.Run(); err != nil {
					log.Error(err)
				}
				cmd = exec.Command("git", "reset", "--hard", reflogHashList[0])
				if err := cmd.Run(); err != nil {
					log.Error(err)
				}
				os.Exit(1)
			}
		}
		/* (END) git rebase */

		/* git rebase with autosquash option */
		speciedHead := fmt.Sprintf("HEAD~%d", beginNumber+1)
		cmd := exec.Command("git", "rebase", "-i", "--autosquash", "--autostash", speciedHead, "--quiet")

		// Transfer the command I/O to Standard I/O
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Error("QS reseted this operation. Please git rebase manually")
			log.Error(err)
			cmd = exec.Command("git", "rebase", "--abort")
			if err := cmd.Run(); err != nil {
				log.Error(err)
			}
			cmd = exec.Command("git", "reset", "--hard", reflogHashList[0])
			if err := cmd.Run(); err != nil {
				log.Error(err)
			}
		}
		log.Info("rebase completed")

		return nil
	}
	app.Run(os.Args)
}
