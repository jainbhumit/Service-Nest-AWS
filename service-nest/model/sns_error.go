package model

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type ErrorNotifier struct {
	SnsClient *sns.Client
	TopicArn  string
}

func (n *ErrorNotifier) NotifyError(details map[string]interface{}, stack []byte) error {
	// Create a formatted message with all error details
	message := fmt.Sprintf(
		"Internal Server Error Details:\n"+
			"Method: %v\n"+
			"Path: %v\n"+
			"Status Code: %v\n"+
			"Request ID: %v\n"+
			"Duration: %v\n"+
			"Source IP: %v\n"+
			"Response Body: %v\n"+
			"Stack Trace:\n%s",
		details["method"],
		details["path"],
		details["statusCode"],
		details["requestID"],
		details["duration"],
		details["sourceIP"],
		details["responseBody"],
		string(stack),
	)

	input := &sns.PublishInput{
		Message:  aws.String(message),
		TopicArn: aws.String(n.TopicArn),
		Subject:  aws.String(fmt.Sprintf("Internal Server Error (500) - %v %v", details["method"], details["path"])),
	}

	_, err := n.SnsClient.Publish(context.Background(), input)
	return err
}
