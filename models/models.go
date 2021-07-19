package models

import "github.com/kamva/mgm/v3"

// List struct
type List struct {
	mgm.DefaultModel `bson:",inline"`
	Username         string `json:"-" bson:"username"`
	Title            string `json:"title" bson:"title"`
	Tasks            []Task `json:"tasks" bson:"tasks"`
}

// Task struct
type Task struct {
	mgm.DefaultModel `bson:",inline"`
	Title            string `json:"title" bson:"title"`
	Body             string `json:"description" bson:"description"`
	Status           int    `json:"status" bson:"status"`
}

// User (List owner) struct used as database model and API output
type User struct {
	mgm.DefaultModel `bson:",inline"`
	Username         string `json:"username" bson:"username"`
	Password         []byte `json:"-" bson:"password"`
}

// UserInput struct used as API input
type UserInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
