package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"log"
)

var EventQueue = make(chan BatchMessage, 100)

func Poll(queueName string) {
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
		log.Fatalln(err)
	}
	queueUrl := queueUrlOutput.QueueUrl
	for {
		receiveMessageOuput, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            queueUrl,
			MaxNumberOfMessages: aws.Int64(10),
			WaitTimeSeconds:     aws.Int64(20),
			VisibilityTimeout:   aws.Int64(60),
		})
		if err != nil {
			log.Fatalln(err)
		}

		if len(receiveMessageOuput.Messages) == 0 {
			fmt.Println("Received no messages")
		} else {
			for _, v := range receiveMessageOuput.Messages {
				event := BatchEvent{}
				err := json.Unmarshal([]byte(*v.Body), &event)
				if err != nil {
					log.Fatalln(err)
				}
				EventQueue <- BatchMessage{
					ReceiptHandle: *v.ReceiptHandle,
					Event:         event,
				}
				fmt.Printf("Added event to queue. Queue is now size %d.\n", len(EventQueue))
			}
		}
	}
}
