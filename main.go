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
		return nil
	}
	app.Run(os.Args)
}
