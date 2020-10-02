package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func main() {
	viper.AutomaticEnv()

	log.Println("Starting application...")

	dbURI := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable search_path=%s", viper.GetString("db_host"), viper.GetString("db_port"), viper.GetString("db_name"), viper.GetString("db_username"), viper.GetString("db_password"), viper.GetString("db_schema"))
	db, err := gorm.Open(postgres.Open(dbURI), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "surveyv2.",
		},
	})
	if err != nil {
		log.Fatalln("Couldn't connect to postgres, " + err.Error())
	}

	db.AutoMigrate(&Survey{}, &CollectionExercise{}, &CollectionInstrument{}, &Email{})

	router := mux.NewRouter()
	log.Println("Application started")
	http.ListenAndServe(":8080", router)

}
