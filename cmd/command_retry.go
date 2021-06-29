package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/urfave/cli/v2"
)

const waitTimeInSeconds = 5
const visibilityTimeoutInSeconds = 10

func commandRetry() *cli.Command {
	return &cli.Command{
		Name:    "retry",
		Aliases: []string{"r"},
		Usage:   "Retry consuming a message",
		Flags:   retryFlags,
		Action: func(c *cli.Context) error {
			retry(c)
			return nil
		},
	}
}

func retry(c *cli.Context) error {
	fmt.Println("Consume message...")
	queueName := c.String("queue-name")
	retryQueueName := c.String("retry-queue-name")

	// Setup
	if queueName == "" {
		log.Fatalln("You must supply the name of a queue")
	}

	if retryQueueName == "" {
		log.Fatalln("You must supply the name of a retry queue")
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

	retryURLResult, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &retryQueueName,
	})

	if err != nil {
		log.Fatalln("could not fetch retry queue url", err)
	}

	queueURL := urlResult.QueueUrl
	retryQueueURL := retryURLResult.QueueUrl

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
				msgResult, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
					AttributeNames: []*string{
						aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
					},
					MessageAttributeNames: []*string{
						aws.String(sqs.QueueAttributeNameAll),
					},
					QueueUrl:            retryQueueURL,
					MaxNumberOfMessages: aws.Int64(5),
					WaitTimeSeconds:     aws.Int64(waitTimeInSeconds),
					VisibilityTimeout:   aws.Int64(visibilityTimeoutInSeconds),
				})

				if err != nil {
					log.Println("error receiving messages")
					continue
				}

				for _, message := range msgResult.Messages {
					fmt.Println("Message Handle: " + *message.ReceiptHandle)
					fmt.Println("Message Body: " + *message.Body)

					var count int
					for key, attr := range message.MessageAttributes {
						fmt.Printf("Key: %s, Value: %s\n", key, *attr.StringValue)

						if key == "RetryCount" {
							value := *attr.StringValue
							count, _ = strconv.Atoi(value) // if value isn't a valid integer, it's fine to default to 0
						}
					}

					if count < 1 {
						count++

						_, err = svc.SendMessage(&sqs.SendMessageInput{
							MessageAttributes: map[string]*sqs.MessageAttributeValue{
								"RetryCount": &sqs.MessageAttributeValue{
									DataType:    aws.String("Number"),
									StringValue: aws.String(fmt.Sprintf("%d", count)),
								},
							},
							MessageBody: aws.String(string(*message.Body)),
							QueueUrl:    queueURL,
						})
					} else {
						log.Println("this message has failed 2 retries")
					}

					_, err := svc.DeleteMessage(&sqs.DeleteMessageInput{
						QueueUrl:      retryQueueURL,
						ReceiptHandle: message.ReceiptHandle,
					})

					if err != nil {
						log.Println("failed to delete message", message.ReceiptHandle)
					}

					continue
				}
			}
		}

	}(&waitGroup, shutdownChannel)

	<-quitChannel
	close(shutdownChannel)

	waitGroup.Wait()

	return nil
}
