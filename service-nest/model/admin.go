package model

type Admin struct {
	User *User `json:"user" bson:"user"`
}
