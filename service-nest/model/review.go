package model

import "time"

type Review struct {
	ID            string    `json:"review_id" bson:"id" dynamodbav:"review_id"`
	RequestId     string    `json:"request_id" bson:"request_id" dynamodbav:"request_id"`
	ServiceID     string    `json:"service_id" bson:"service_id" dynamodbav:"service_id"`
	HouseholderID string    `json:"householder_id" bson:"householder_id" dynamodbav:"householder_id"`
	ProviderID    string    `json:"provider_id" dynamodbav:"provider_id"`
	Rating        float64   `json:"rating" bson:"rating" dynamodbav:"rating"`
	Comments      string    `json:"comments" bson:"comments" dynamodbav:"comments"`
	ReviewDate    time.Time `json:"review_date" bson:"review_date" dynamodbav:"review_date"`
}
