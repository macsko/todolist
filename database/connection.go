package database

import (
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

// Initialization of database
func Init() {
	// Setup the mgm default config
	err := mgm.SetDefaultConfig(nil, "todolist", options.Client().ApplyURI("mongodb://" + os.Getenv("DB_PATH")))
	if err != nil {
		panic(err)
	}
}
