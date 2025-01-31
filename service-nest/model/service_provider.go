package model

type ServiceProvider struct {
	User
	ServicesOffered []Service `json:"services_offered" bson:"services_offered"`
	Rating          float64   `json:"rating" bson:"rating"`
	Reviews         []*Review `json:"reviews" bson:"reviews"`
	Availability    bool      `json:"availability" bson:"availability"`
	IsActive        bool      `json:"is_active" bson:"is_active"`
}
