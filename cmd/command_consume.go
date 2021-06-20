package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"replay/queue"
	"sync"
	"syscall"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/urfave/cli/v2"
)

func commandConsume() *cli.Command {
	return &cli.Command{
		Name:    "consume",
		Aliases: []string{"c"},
		Usage:   "Consume a message from a queue",
		Flags:   consumeFlags,
		Action: func(c *cli.Context) error {
			fmt.Println("Consume message...")
			return consume(c)
		},
	}
}

func consume(c *cli.Context) error {
	err := queue.Initialize(c)
	if err != nil {
		log.Fatalln("could not initialize queue")
		return err
	}

	// Poll queue
	quitChannel := make(chan os.Signal)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	shutdownChannel := make(chan bool)

	var waitGroup sync.WaitGroup
	waitGroup.Add(1)

	go func(wg *sync.WaitGroup, shutdownChan <-chan bool) {
		defer wg.Done()
		for {
			select {
			case _ = <-shutdownChan:
				log.Println("Shutdown signal received, retrieval of queue messages will stop soon.")
				return
			default:
				err := queue.FetchMessages(func(message *sqs.Message) error {
					var vote Vote
					err = json.Unmarshal([]byte(*message.Body), &vote)
					if err != nil {
						log.Println("error unmarshalling json", err)
						return err
					}

					log.Printf("Vote is %+v\n", vote)
					if vote.ID == "" || vote.Position == "" {
						return errors.New("invalid vote: missing required data")
					}

					if vote.ID == "scott-stringer" {
						return errors.New("do not count votes for this candidate")
					}

					return nil
				})

				if err != nil {
					log.Println("failed while fetching messages", err)
					return
				}
			}
		}

	}(&waitGroup, shutdownChannel)

	<-quitChannel
	close(shutdownChannel)

	waitGroup.Wait()

	return nil
}
