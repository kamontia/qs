package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/codegangsta/cli"
)

func main() {

	app := cli.NewApp()
	app.Name = Name
	app.Version = Version
	app.Author = "Tatsuya Kamohara<kamontia@gmail.com>\n   Takeshi Kondo<take.she12@gmail.com>"
	app.Email = ""
	app.Usage = ""

	app.Flags = GlobalFlags
	app.Commands = Commands
	app.CommandNotFound = CommandNotFound

	app.Action = func(c *cli.Context) error {
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
		fmt.Println("*** rebase with autosquash ***")
		if err = cmd.Run(); err != nil {
			fmt.Println(err)
		}

		return nil
	}
	app.Run(os.Args)
}
