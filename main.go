package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	viper.AutomaticEnv()

	log.Println("Starting application...")

	dbURI := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable ", viper.GetString("db_host"), viper.GetString("db_port"), viper.GetString("db_name"), viper.GetString("db_username"), viper.GetString("db_password"))
	db, err := gorm.Open(postgres.Open(dbURI), &gorm.Config{})
	if err != nil {
		log.Fatalln("Couldn't connect to postgres, " + err.Error())
	}

	db.AutoMigrate(&Survey{})

	router := mux.NewRouter()
	log.Println("Application started")
	http.ListenAndServe(":8080", router)

}
