package cmd

import (
	"encoding/json"
	"fmt"
	"log"

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

		var vote Vote
		err = json.Unmarshal([]byte(*message.Body), &vote)
		if err != nil {
			log.Println("error unmarshalling json", err)
			return err
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
				return err
			}
		}
	}

	return err
}
