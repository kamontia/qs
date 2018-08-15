package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"

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
		switch number := c.Int("number"); number {
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

			/* Get commit message */
			out, err = exec.Command("git", "log", "--oneline", "--format=%s").Output()
			if err != nil {
				fmt.Print(string(out))
				fmt.Print(err)
				os.Exit(1)
			}
			for _, v := range regexp.MustCompile("\r\n|\n|\r").Split(string(out), -1) {
				commitMsg = append(commitMsg, v)
			}
			/* (END)Get commit message */

			/* Display commit hash and message. The [pickup|..] strings is colored */
			for i := 0; i < len(commitMsg)-1; i++ {

				/* (WIP) Switch output corresponded to do squash */
				if i > 1 && i < 8 {
					fmt.Printf("[%2d] \x1b[35mpickup\x1b[0m -> \x1b[36msquash\x1b[0m %s\tsquash! %s\n", i, commitHashList[i], commitMsg[i])
				} else {
					fmt.Printf("[%2d] \x1b[35mpickup\x1b[0m -> \x1b[35mpickup\x1b[0m %s\t%s\n", i, commitHashList[i], commitMsg[i])
				}
			}
			/* (END)Display commit hash and message */
			os.Exit(1) // This line will be removed.

			/* (WIP) */
			for i := 0; i < number; i++ {
				fixUpStr := fmt.Sprintf("--squash=%s", commitHashList[i])
				cmd := exec.Command("git", "commit", fixUpStr)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err = cmd.Run(); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
			/* (END) */

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
		}

		return nil
	}
	app.Run(os.Args)
}
