package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func Poll(ctx context.Context, queueName string, messages chan<- BatchMessage) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String("us-east-1"),
			CredentialsChainVerboseErrors: aws.Bool(true),
		},
	}))
	svc := sqs.New(sess)

	queueUrlOutput, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		fmt.Printf("could not get SQS queue url: %s\n", err)
		close(messages)
		return
	}
	for {
		receiveMessageOuput, err := svc.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            queueUrlOutput.QueueUrl,
			MaxNumberOfMessages: aws.Int64(10),
			WaitTimeSeconds:     aws.Int64(20),
			VisibilityTimeout:   aws.Int64(60),
		})
		if err != nil {
			fmt.Printf("could not get messages from SQS queue: %s\n", err)
			close(messages)
			return
		}

		if len(receiveMessageOuput.Messages) == 0 {
			fmt.Println("Received no messages")
		} else {
			for _, v := range receiveMessageOuput.Messages {
				event := BatchEvent{}
				err := json.Unmarshal([]byte(*v.Body), &event)
				if err != nil {
					fmt.Printf("could not unmarshall SQS message: %s\n", err)
					close(messages)
					return
				}
				messages <- BatchMessage{
					ReceiptHandle: *v.ReceiptHandle,
					Event:         event,
				}
				fmt.Printf("Added event to queue. Queue is now size %d.\n", len(messages))
			}
		}
	}
}
