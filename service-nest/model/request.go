package model

import "time"

type ServiceRequest struct {
	ID                 string                   `json:"request_id" dynamodbav:"request_id"`
	HouseholderID      *string                  `json:"householder_id" bson:"HouseholderID" dynamodbav:"householder_id"`
	HouseholderName    string                   `json:"householder_name" bson:"HouseholderName" dynamodbav:"householder_name"`
	HouseholderAddress *string                  `json:"householder_address" bson:"HouseholderAddress"  dynamodbav:"householder_address"`
	HouseholderContact string                   `json:"householder_contact" bson:"HouseholderPhone" dynamodbav:"householder_contact"`
	ServiceName        string                   `json:"service_name" bson:"ServiceName" dynamodbav:"service_name"`
	ServiceID          string                   `json:"service_id" bson:"serviceID" dynamodbav:"service_id"`
	RequestedTime      time.Time                `json:"requested_time" bson:"requestedTime" dynamodbav:"requested_time"`
	ScheduledTime      time.Time                `json:"scheduled_time" bson:"scheduledTime" dynamodbav:"scheduled_time"`
	Status             string                   `json:"status" bson:"status" dynamodbav:"status"` // Pending, Accepted, Completed, Cancelled
	ApproveStatus      bool                     `json:"approve_status" bson:"approveStatus" dynamodbav:"approve_status"`
	Description        string                   `json:"description" bson:"description" dynamodbav:"description"`
	ProviderDetails    []ServiceProviderDetails `json:"provider_details,omitempty" bson:"providerDetails,omitempty" dynamodbav:"provider_details,omitempty"`
}
type ServiceProviderDetails struct {
	ServiceProviderID string  `json:"service_provider_id" bson:"serviceProviderID" dynamodbav:"service_provider_id"`
	Name              string  `json:"name" bson:"name" dynamodbav:"name"`
	Contact           string  `json:"contact" bson:"contact" dynamodbav:"contact"`
	Address           string  `json:"address" bson:"address" dynamodbav:"address"`
	Price             string  `json:"price" bson:"price" dynamodbav:"price"`
	Rating            float64 `json:"rating" bson:"rating" dynamodbav:"rating"`
	RatingCount       int64   `json:"rating_count" bson:"rating_count" dynamodbav:"rating_count"`
	Approve           int     `json:"approve" bson:"approve" dynamodbav:"approve"`
}
