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

func commandProduce() *cli.Command {
	return &cli.Command{
		Name:    "produce",
		Aliases: []string{"p"},
		Usage:   "Produce a message to a queue",
		Flags:   consumeFlags,
		Action: func(c *cli.Context) error {
			produce(c)
			return nil
		},
	}
}

// Message represents a vote to be added to the queue
type Message struct {
	ID       string `json:"id"`
	Position string `json:"position"`
}

func produce(c *cli.Context) error {
	fmt.Println("Produce message...")
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

	message := Message{
		ID:       "maya-wiley",
		Position: "mayor",
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	queueURL := urlResult.QueueUrl
	_, err = svc.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String(string(messageBytes)),
		QueueUrl:    queueURL,
	})

	return err
}
