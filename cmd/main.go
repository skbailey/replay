package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/urfave/cli/v2"
)

// Run launches the CLI program
func Run() {
	app := &cli.App{
		Name:  "send",
		Usage: "Sends a message to a queue",
		Flags: globalFlags,
		Action: func(c *cli.Context) error {
			fmt.Println("Send message...")
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

			queueURL := urlResult.QueueUrl
			msgResult, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
				AttributeNames: []*string{
					aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
				},
				MessageAttributeNames: []*string{
					aws.String(sqs.QueueAttributeNameAll),
				},
				QueueUrl:            queueURL,
				MaxNumberOfMessages: aws.Int64(5),
			})

			for _, message := range msgResult.Messages {
				fmt.Println("Message Handle: " + *message.ReceiptHandle)
				fmt.Println("Message Body: " + *message.Body)
			}
			return err
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
