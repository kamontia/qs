package main

import (
	"os"
	"fmt"
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
		out, err := exec.Command("git", "status").Output()

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(string(out))
		return nil
	  }
	app.Run(os.Args)
}
