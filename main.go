package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
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
	}
	app.Commands = Commands
	app.CommandNotFound = CommandNotFound

	app.Action = func(c *cli.Context) error {
		// Intaractive
		var force bool = c.Bool("force")
		var num string = strconv.Itoa(c.Int("number"))
		var stdin string
		var rangeArray []string
		var iBreakNumber int
		iNum, _ := strconv.Atoi(num)
		var error error

		/* Pick up squash range */
		/* TODO: Check error strictly */
		for _, v := range os.Args {
			if strings.Contains(v, "..") {
				rangeArray = strings.Split(v, "..")
				iNum, error = strconv.Atoi(rangeArray[0])
				if error != nil {

				}
				iBreakNumber, error = strconv.Atoi(rangeArray[1])
				if error != nil {
					fmt.Println(error)
					os.Exit(1)
				}

				if iNum < iBreakNumber {
					tmp := iNum
					iNum = iBreakNumber
					iBreakNumber = tmp
				}

				break
			}
		}

		if len(rangeArray) == 0 {
			fmt.Println("ERROR: argument erorr")
			os.Exit(1)
		}
		/* (END) Pick up squash range */

		if force {
			fmt.Println("*** force update ***")
		} else {
			fmt.Println("*** Do you fixup the following commits?(y/n) ***")
			out, err := exec.Command("git", "log", "--oneline", "-n", num).Output()
			if err != nil {
				fmt.Print(out)
				os.Exit(1)
			}
			for {
				fmt.Scan(&stdin)
				switch stdin {
				case "y":
					fmt.Println("*** Fixup! ***")
				case "n":
					fmt.Println("*** Abort! ***")
					os.Exit(1)
				default:
					fmt.Println("*** You can input y or n ***")
					continue
				}
				break
			}
		}
		// Parse number(--number, -n) parameter
		switch iNum {
		case 0:
			// fix up commit
			fmt.Println("*** git commit --fixup ***")
			out, err := exec.Command("git", "commit", "--fixup=HEAD").Output()

			if err != nil {
				fmt.Println(string(out))
				os.Exit(1)
			}

			fmt.Println(string(out))
			// rebase
			os.Setenv("GIT_EDITOR", ":")
			cmd := exec.Command("git", "rebase", "-i", "--autosquash", "--autostash", "HEAD~2")
			// Transfer the command I/O to Standard I/O
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			fmt.Println("*** rebase with autosquash ***")
			if err = cmd.Run(); err != nil {
				fmt.Println(err)
			}

		default:
			var commitHashList []string
			var commitMsg []string
			var commitNewMsg []string
			var reflogHashList []string
			/* Get commit hash */
			out, err := exec.Command("git", "log", "--oneline", "--format=%h").Output()
			if err != nil {
				fmt.Print(string(out))
				fmt.Print(err)
				os.Exit(1)
			}

			for _, v := range regexp.MustCompile("\r\n|\n|\r").Split(string(out), -1) {
				commitHashList = append(commitHashList, v)
			}
			/* (END)Get commit hash */

			/* Get reflog hash */
			out, err = exec.Command("git", "reflog", "--format=%h").Output()
			if err != nil {
				fmt.Print(string(out))
				fmt.Print(err)
				os.Exit(1)
			}
			for _, v := range regexp.MustCompile("\r\n|\n|\r").Split(string(out), -1) {
				reflogHashList = append(reflogHashList, v)
			}
			/* (END)Get reflog hash */

			/* Get commit message */
			out, err = exec.Command("git", "log", "--oneline", "--format=%s").Output()
			if err != nil {
				fmt.Print(string(out))
				fmt.Print(err)
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
				if iNum > i && i >= iBreakNumber {
					fmt.Printf("[%2d] \x1b[35mpickup\x1b[0m -> \x1b[36msquash\x1b[0m %s %s\n", i, commitHashList[i], commitNewMsg[iNum])
				} else {
					fmt.Printf("[%2d] \x1b[35mpickup\x1b[0m -> \x1b[35mpickup\x1b[0m %s %s\n", i, commitHashList[i], commitMsg[i])
				}
			}
			/* (END)Display commit hash and message */

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
				if iNum > i && i >= iBreakNumber {
					speciedExec = fmt.Sprintf("--exec=git commit --amend -m\"%s\"", commitNewMsg[iNum])
				} else {
					speciedExec = fmt.Sprintf("--exec=git commit --amend -m\"%s\"", commitMsg[i])
				}

				cmd := exec.Command("git", "rebase", speciedHead, speciedExec)
				log.Printf("git rebaes HEAD~%d %s\n", i, speciedExec)

				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err = cmd.Run(); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
			/* (END) git rebase */

			/* git rebase with autosquash option */
			speciedHead := fmt.Sprintf("HEAD~%d", iNum+1)
			cmd := exec.Command("git", "rebase", "-i", "--autosquash", "--autostash", speciedHead)
			// Transfer the command I/O to Standard I/O
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err = cmd.Run(); err != nil {
				fmt.Println("*** rebase failed ***")
				fmt.Println(err)
			}
			fmt.Println("*** rebase completed ***")
		}

		return nil
	}
	app.Run(os.Args)
}
