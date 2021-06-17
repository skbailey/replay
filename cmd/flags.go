package cmd

import (
	"github.com/urfave/cli/v2"
)

var consumeFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "queue-name",
		Usage:   "Defines the name of the queue",
		EnvVars: []string{"SQS_QUEUE_NAME"},
	},
}
