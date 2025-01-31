package model

type Service struct {
	ID              string  `json:"id" bson:"id" dynamodbav:"id"`
	Name            string  `json:"name" bson:"name" dynamodbav:"name"`
	Description     string  `json:"description" bson:"description" dynamodbav:"description"`
	Price           float64 `json:"price" bson:"price" dynamodbav:"price"`
	ProviderID      string  `json:"provider_id" bson:"provider_id" dynamodbav:"provider_id"`
	Category        string  `json:"category" bson:"category" dynamodbav:"category"`
	AvgRating       float64 `json:"avg_rating" bson:"avg_rating" dynamodbav:"avg_rating"`
	RatingCount     int64   `json:"rating_count" bson:"rating_count" dynamodbav:"rating_count"`
	ProviderName    string  `json:"provider_name" bson:"provider_name" dynamodbav:"provider_name"`
	ProviderContact string  `json:"provider_contact" bson:"provider_contact" dynamodbav:"provider_contact"`
	ProviderAddress string  `json:"provider_address" bson:"provider_address" dynamodbav:"provider_address"`
}
