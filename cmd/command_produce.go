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
		Flags:   produceFlags,
		Action: func(c *cli.Context) error {
			produce(c)
			return nil
		},
	}
}

// Vote represents a vote to be added to the queue
type Vote struct {
	ID       string `json:"id"`
	Position string `json:"position"`
}

func produce(c *cli.Context) error {
	fmt.Println("Produce message...")
	queueName := c.String("queue-name")
	candidatePosition := c.String("candidate-position")
	candidateID := c.String("candidate-id")

	if queueName == "" {
		log.Fatalln("You must supply the name of a queue")
	}

	if candidatePosition == "" {
		log.Fatalln("You must provide the candidate position")
	}

	if candidateID == "" {
		log.Fatalln("You must provide the candidate id")
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

	message := Vote{
		ID:       candidateID,
		Position: candidatePosition,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	queueURL := urlResult.QueueUrl
	_, err = svc.SendMessage(&sqs.SendMessageInput{
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"RetryCount": &sqs.MessageAttributeValue{
				DataType:    aws.String("Number"),
				StringValue: aws.String("0"),
			},
		},
		MessageBody: aws.String(string(messageBytes)),
		QueueUrl:    queueURL,
	})

	return err
}
