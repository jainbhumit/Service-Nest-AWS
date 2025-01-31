package model

type Category struct {
	ID          string `json:"id" dynamodbav:"id"`
	Name        string `json:"name" dynamodbav:"name"`
	Description string `json:"description" dynamodbav:"description"`
	ImageUrl    string `json:"imageUrl" dynamodbav:"imageUrl"`
}
