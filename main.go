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
var iNum int
var iBreakNumber int
var rangeArray []string
var commitHashList []string
var commitMsg []string
var commitNewMsg []string
var commitSpecifiedMsg []string
var reflogHashList []string
var specifiedMsg string

func logrus_init(d bool) {
	var debug bool = d
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

func check_current_commit(f bool, iNum int, iBreakNumber int) {
	var force bool = f
	var sNum = strconv.Itoa(iNum)
	log.SetOutput(os.Stdout)

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

	/* Get reflog hash */
	out, err = exec.Command("git", "reflog", "--format=%h").Output()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	for _, v := range regexp.MustCompile("\r\n|\n|\r").Split(string(out), -1) {
		reflogHashList = append(reflogHashList, v)
	}
	/* (END)Get reflog hash */

	/* Get commit message */
	out, err = exec.Command("git", "log", "--oneline", "--format=%s").Output()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	for _, v := range regexp.MustCompile("\r\n|\n|\r").Split(string(out), -1) {
		commitMsg = append(commitMsg, v)
		commitNewMsg = append(commitNewMsg, fmt.Sprintf("fixup! %s", v))
		commitSpecifiedMsg = append(commitSpecifiedMsg, fmt.Sprintf("fixup! %s", specifiedMsg))
	}
	/* (END)Get commit message */
	if force {
		log.Info("force update")
	} else {
		/* Display commit hash and message. The [pickup|..] strings is colored */
		for i := len(commitMsg) - 2; i >= 0; i-- {
			/* Switch output corresponded to do squash */
			if iNum > i && i >= iBreakNumber {
				log.Warnf("[%2d]\t\x1b[35mpickup\x1b[0m -> \x1b[36msquash \x1b[0m %s %s", i, commitHashList[i], commitNewMsg[iNum])
			} else {
				log.Warnf("[%2d]\t\x1b[35mpickup\x1b[0m -> \x1b[35mpickup \x1b[0m %s %s", i, commitHashList[i], commitMsg[i])
			}
		}
		/* (END) Display commit hash and message */
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

		/* Pick up squash range */
		/* TODO: Check error strictly */
		var error error

		if strings.Contains(c.String("number"), "..") {
			/* Specify the range you aggregate */
			rangeArray := strings.Split(c.String("number"), "..")
			iNum, error = strconv.Atoi(rangeArray[0])
			if error != nil {
				log.Error(error)
				os.Exit(1)
			}
			iBreakNumber, error = strconv.Atoi(rangeArray[1])
			if error != nil {
				log.Error(error)
				os.Exit(1)
			}
			if iNum < iBreakNumber {
				tmp := iNum
				iNum = iBreakNumber
				iBreakNumber = tmp
			}
		} else {
			/* Specify the one parameter you aggregate */
			iBreakNumber = 0
			iNum, error = strconv.Atoi(c.String("number"))
			if error != nil {
				log.Error(error)
			}
		}
		/* (END) Pick up squash range */
		logrus_init(c.Bool("debug"))
		check_current_commit(c.Bool("force"), iNum, iBreakNumber)

		// Parse number(--number, -n) parameter

		/* (WIP) git rebase */
		/**
		git rebase HEAD~N --exec="git commit -m"squash! commit messages" "
		*/
		/* Suppress vim editor launching */
		os.Setenv("GIT_EDITOR", ":")
		/* (END) Suppress vim editor launching */
		for i := iNum; i >= 0; i-- {
			speciedHead := fmt.Sprintf("HEAD~%d", i+1)
			var speciedExec string

			if specifiedMsg != "" {
				if iNum == i {
					speciedExec = fmt.Sprintf("--exec=git commit --amend -m\"%s\"", specifiedMsg)
				} else if iNum > i && i >= iBreakNumber {
					speciedExec = fmt.Sprintf("--exec=git commit --amend -m\"%s\"", commitSpecifiedMsg[iNum])
				} else {
					speciedExec = fmt.Sprintf("--exec=git commit --amend -m\"%s\"", commitMsg[i])
				}
			} else {
				if iNum > i && i >= iBreakNumber {
					speciedExec = fmt.Sprintf("--exec=git commit --amend -m\"%s\"", commitNewMsg[iNum])
				} else {
					speciedExec = fmt.Sprintf("--exec=git commit --amend -m\"%s\"", commitMsg[i])
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
		speciedHead := fmt.Sprintf("HEAD~%d", iNum+1)
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
