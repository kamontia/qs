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

	//	app.Flags = GlobalFlags
	app.Commands = Commands
	app.CommandNotFound = CommandNotFound

	// flags
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "force option",
		},
	}

	app.Action = func(c *cli.Context) error {

		var force bool = c.Bool("f")
		if !force {
			os.Exit(1)
		}

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

		return nil
	}
	app.Run(os.Args)
}
