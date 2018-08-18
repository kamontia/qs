package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

func main() {

	app := cli.NewApp()
	app.Name = Name
	app.Version = Version
	app.Author = "Tatsuya Kamohara<kamontia@gmail.com>\n   Takeshi Kondo<take.she12@gmail.com>"
	app.Email = ""
	app.Usage = ""

	app.Flags = []cli.Flag{
		cli.IntFlag{
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
	}
	app.Commands = Commands
	app.CommandNotFound = CommandNotFound

	app.Action = func(c *cli.Context) error {
		// logrus init
		var debug bool = c.Bool("debug")
		log.SetOutput(os.Stdout)
		if debug {
			log.SetLevel(log.InfoLevel)
		} else {
			log.SetLevel(log.WarnLevel)
		}

		// Intaractive
		var force bool = c.Bool("force")
		var num string = strconv.Itoa(c.Int("number"))
		var stdin string

		if force {
			log.Info("*** force update ***")
		} else {
			fmt.Println("*** Do you fixup the following commits?(y/n) ***")
			out, err := exec.Command("git", "log", "--oneline", "-n", num).Output()
			if err != nil {
				log.Error(out)
				os.Exit(1)
			}
			for {
				fmt.Scan(&stdin)
				switch stdin {
				case "y":
					log.Info("*** Fixup! ***")
				case "n":
					log.Info("*** Abort! ***")
					os.Exit(1)
				default:
					log.Info("*** You can input y or n ***")
					continue
				}
				break
			}
		}

		// Parse number(--number, -n) parameter
		switch number := c.Int("number"); number {
		case 0:
			// fix up commit
			log.Info("*** git commit --fixup ***")
			out, err := exec.Command("git", "commit", "--fixup=HEAD", "--quiet").Output()

			if err != nil {
				log.Error(string(out))
				os.Exit(1)
			}

			log.Info(string(out))
			// rebase
			os.Setenv("GIT_EDITOR", ":")
			cmd := exec.Command("git", "rebase", "-i", "--autosquash", "--autostash", "HEAD~2", "--quiet")
			// Transfer the command I/O to Standard I/O
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			log.Info("*** rebase with autosquash ***")
			if err = cmd.Run(); err != nil {
				log.Error(err)
			}

		default:
			var commitHashList []string
			var commitMsg []string
			var commitNewMsg []string
			/* Get commit hash */
			out, err := exec.Command("git", "log", "--oneline", "--format=%h", "--quiet").Output()
			if err != nil {
				log.Error(string(out))
				log.Error(err)
				os.Exit(1)
			}

			for _, v := range regexp.MustCompile("\r\n|\n|\r").Split(string(out), -1) {
				commitHashList = append(commitHashList, v)
			}
			/* (END)Get commit hash */

			/* Get commit message */
			out, err = exec.Command("git", "log", "--oneline", "--format=%s", "--quiet").Output()
			if err != nil {
				log.Error(string(out))
				log.Error(err)
				os.Exit(1)
			}
			for _, v := range regexp.MustCompile("\r\n|\n|\r").Split(string(out), -1) {
				commitMsg = append(commitMsg, v)
				commitNewMsg = append(commitNewMsg, fmt.Sprintf("squash! %s", v))
			}
			/* (END)Get commit message */

			/* Display commit hash and message. The [pickup|..] strings is colored */
			for i := len(commitMsg) - 1; i >= 0; i-- {
				/* (WIP) Switch output corresponded to do squash */
				if c.Int("number") > i {
					log.Info("[%2d] \x1b[35mpickup\x1b[0m -> \x1b[36msquash\x1b[0m %s %s\n", i, commitHashList[i], commitNewMsg[number])
				} else {
					log.Info("[%2d] \x1b[35mpickup\x1b[0m -> \x1b[35mpickup\x1b[0m %s %s\n", i, commitHashList[i], commitMsg[i])
				}
			}
			/* (END)Display commit hash and message */

			/* (WIP) git rebase */
			/**
			git rebase HEAD~N --exec="git commit -m"squash! commit messages" "
			*/
			log.Info("Logged display")
			/* Suppress vim editor launching */
			os.Setenv("GIT_EDITOR", ":")
			/* (END) Suppress vim editor launching */
			for i := number; i >= 0; i-- {
				speciedHead := fmt.Sprintf("HEAD~%d", i)
				speciedExec := fmt.Sprintf("--exec=git commit --quiet --amend -m\"%s\"", commitNewMsg[number])
				cmd := exec.Command("git", "rebase", speciedHead, speciedExec, "--quiet")

				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err = cmd.Run(); err != nil {
					log.Error(err)
					os.Exit(1)
				}
			}
			/* (END) git rebase */

			log.Info("Logged display")
			/* git rebase with autosquash option */
			speciedHead := fmt.Sprintf("HEAD~%d", number+1)
			cmd := exec.Command("git", "rebase", "-i", "--autosquash", "--autostash", speciedHead, "--quiet")
			// Transfer the command I/O to Standard I/O
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err = cmd.Run(); err != nil {
				log.Error("*** rebase failed ***")
				log.Error(err)
			}
			log.Info("*** rebase completed ***")
		}

		return nil
	}
	app.Run(os.Args)
}
