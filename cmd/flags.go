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

var produceFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "queue-name",
		Usage:   "Defines the name of the queue",
		EnvVars: []string{"SQS_QUEUE_NAME"},
	},
	&cli.StringFlag{
		Name:    "candidate-position",
		Usage:   "The political position to be filled",
		EnvVars: []string{"CANDIDATE_POSITION"},
	},
	&cli.StringFlag{
		Name:    "candidate-id",
		Usage:   "The id of the candidate",
		EnvVars: []string{"CANDIDATE_ID"},
	},
}

var retryFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "queue-name",
		Usage:   "The name of the queue",
		EnvVars: []string{"SQS_QUEUE_NAME"},
	},
	&cli.StringFlag{
		Name:    "retry-queue-name",
		Usage:   "The name of the retry queue",
		EnvVars: []string{"SQS_RETRY_QUEUE_NAME"},
	},
	&cli.IntFlag{
		Name:    "retry-count",
		Usage:   "The number of retry attempts for a failed message",
		Value:   1,
		EnvVars: []string{"RETRY_COUNT"},
	},
}
