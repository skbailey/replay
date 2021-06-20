package queue

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/urfave/cli/v2"
)

var sqsService *sqs.SQS
var sqsURL *string
var sqsRetryURL *string

const waitTimeInSeconds = 5
const visibilityTimeoutInSeconds = 10

// Initialize sets up the queue
func Initialize(c *cli.Context) error {
	queueName := c.String("queue-name")

	if queueName == "" {
		return errors.New("you must supply the name of a queue")
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	sqsService = sqs.New(sess)

	var err error
	sqsURL, err = getQueueURL(queueName)
	if err != nil {
		return err
	}

	sqsRetryURL, err = getQueueURL(fmt.Sprintf("%s-Retry", queueName))
	return err
}

func getQueueURL(name string) (*string, error) {
	urlResult, err := sqsService.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &name,
	})

	if err != nil {
		return nil, err
	}

	return urlResult.QueueUrl, nil
}

// FetchMessages polls the queue for messages
func FetchMessages(callback func(*sqs.Message) error) error {
	msgResult, err := sqsService.ReceiveMessage(&sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            sqsURL,
		MaxNumberOfMessages: aws.Int64(5),
		WaitTimeSeconds:     aws.Int64(waitTimeInSeconds),
		VisibilityTimeout:   aws.Int64(visibilityTimeoutInSeconds),
	})

	if err != nil {
		log.Println("error receiving messages")
		return err
	}

	for _, message := range msgResult.Messages {
		fmt.Println("Message Handle: " + *message.ReceiptHandle)
		fmt.Println("Message Body: " + *message.Body)

		for key, attr := range message.MessageAttributes {
			fmt.Printf("Key: %s, Value: %s\n", key, *attr.StringValue)
		}

		err = callback(message)
		if err != nil {
			log.Println("failed to process message", err)
			queueForRetry(message)
			continue
		}

		err = deleteMessage(message)
		if err != nil {
			log.Println("failed to delete message")
		}
	}

	return err
}

func queueForRetry(message *sqs.Message) error {
	_, err := sqsService.SendMessage(&sqs.SendMessageInput{
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"RetryCount": &sqs.MessageAttributeValue{
				DataType:    aws.String("Number"),
				StringValue: aws.String("0"),
			},
		},
		MessageBody: aws.String(string(*message.Body)),
		QueueUrl:    sqsRetryURL,
	})

	return err
}

func deleteMessage(message *sqs.Message) error {
	_, err := sqsService.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      sqsURL,
		ReceiptHandle: message.ReceiptHandle,
	})

	return err
}
