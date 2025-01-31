package model

import "time"

type Review struct {
	ID            string    `json:"id" bson:"id"`
	ServiceID     string    `json:"service_id" bson:"service_id"`
	HouseholderID string    `json:"householder_id" bson:"householder_id"`
	ProviderID    string    `json:"provider_id"`
	Rating        float64   `json:"rating" bson:"rating"`
	Comments      string    `json:"comments" bson:"comments"`
	ReviewDate    time.Time `json:"review_date" bson:"review_date"`
}
