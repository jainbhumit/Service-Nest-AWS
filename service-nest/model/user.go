package model

type User struct {
	ID       string `json:"id" dynamodbav:"id"`
	Name     string `json:"name" dynamodbav:"name"`
	Email    string `json:"email" dynamodbav:"email"`
	Password string `json:"password" dynamodbav:"password"`
	Role     string `json:"role" dynamodbav:"role"` // Householder or ServiceProvider
	Address  string `json:"address" dynamodbav:"address"`
	Contact  string `json:"contact" dynamodbav:"contact"`
	IsActive bool   `json:"is_active" dynamodbav:"is_active"`
}
