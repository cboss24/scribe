package main

import (
	"github.com/aws/aws-sdk-go/service/batch"
	"time"
)

type BatchEvent struct {
	Version    string          `json:"version"`
	ID         string          `json:"id"`
	DetailType string          `json:"detail-type"`
	Source     string          `json:"source"`
	Account    string          `json:"account"`
	Time       time.Time       `json:"time"`
	Region     string          `json:"region"`
	Resources  []string        `json:"resources"`
	Detail     batch.JobDetail `json:"detail"`
}
