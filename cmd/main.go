package cmd

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

// Run launches the CLI program
func Run() {
	app := &cli.App{
		Name:  "replay",
		Usage: "Replays failed messages in a queue",
		Commands: []*cli.Command{
			commandConsume(),
			commandProduce(),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
