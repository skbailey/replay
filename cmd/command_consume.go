package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
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
			consume(c)
			return nil
		},
	}
}

func consume(c *cli.Context) error {
	fmt.Println("Consume message...")
	queueName := c.String("queue-name")

	if queueName == "" {
		log.Fatalln("You must supply the name of a queue")
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := sqs.New(sess)
	urlResult, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &queueName,
	})

	if err != nil {
		log.Fatalln("could not fetch queue url", err)
	}

	// Poll queue
	quitChannel := make(chan os.Signal)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	shutdownChannel := make(chan bool)

	var waitGroup sync.WaitGroup
	waitGroup.Add(1)

	queueURL := urlResult.QueueUrl
	go func(wg *sync.WaitGroup, shutdownChan <-chan bool) {
		defer wg.Done()
		for {
			select {
			case _ = <-shutdownChan:
				log.Println("Shutdown signal received, retrieval of queue messages will stop soon.")
				return
			default:
				msgResult, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
					AttributeNames: []*string{
						aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
					},
					MessageAttributeNames: []*string{
						aws.String(sqs.QueueAttributeNameAll),
					},
					QueueUrl:            queueURL,
					MaxNumberOfMessages: aws.Int64(5),
					WaitTimeSeconds:     aws.Int64(waitTimeInSeconds),
					VisibilityTimeout:   aws.Int64(visibilityTimeoutInSeconds),
				})

				if err != nil {
					log.Println("error receiving messages")
				}

				for _, message := range msgResult.Messages {
					fmt.Println("Message Handle: " + *message.ReceiptHandle)
					fmt.Println("Message Body: " + *message.Body)

					for key, attr := range message.MessageAttributes {
						fmt.Printf("Key: %s, Value: %s\n", key, *attr.StringValue)
					}

					var vote Vote
					err = json.Unmarshal([]byte(*message.Body), &vote)
					if err != nil {
						log.Println("error unmarshalling json", err)
						return
					}

					log.Printf("Vote is %+v\n", vote)

					if vote.ID == "andrew-yang" {
						log.Println("Deleting vote for ", vote.ID)
						_, err = svc.DeleteMessage(&sqs.DeleteMessageInput{
							QueueUrl:      queueURL,
							ReceiptHandle: message.ReceiptHandle,
						})

						if err != nil {
							log.Println("error deleting message from queue", err)
							return
						}
					}
				}
			}
		}

	}(&waitGroup, shutdownChannel)

	<-quitChannel
	close(shutdownChannel)

	waitGroup.Wait()

	return nil
}
