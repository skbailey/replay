package main

import "replay/cmd"

// func pollSqs(chn chan<- *sqs.Message) {
//
// 	for {
// 		output, err := sqsSvc.ReceiveMessage(&sqs.ReceiveMessageInput{
// 			QueueUrl:            &config.StackNotificationSqsUrl,
// 			MaxNumberOfMessages: aws.Int64(sqsMaxMessages),
// 			WaitTimeSeconds:     aws.Int64(sqsPollWaitSeconds),
// 		})
//
// 		if err != nil {
// 			log.Errorf("failed to fetch sqs message %v", err)
// 		}
//
// 		for _, message := range output.Messages {
// 			chn <- message
// 		}
//
// 	}
//
// }

func main() {
	cmd.Run()
}
