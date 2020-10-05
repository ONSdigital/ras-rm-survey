package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ONSdigital/ras-rm-survey/logger"
	"github.com/ONSdigital/ras-rm-survey/models"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func main() {
	viper.AutomaticEnv()
	setDefaults()
	err := logger.ConfigureLogger()
	if err != nil {
		log.Fatalln("Couldn't set up a logger, exiting", err)
	}

	logger.Logger.Info("Starting ras-rm-survey...")

	dbURI := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable searchpath=%s", viper.GetString("db_host"), viper.GetString("db_port"), viper.GetString("db_name"), viper.GetString("db_username"), viper.GetString("db_password"), viper.GetString("db_schema"))
	db, err := gorm.Open(postgres.Open(dbURI), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "surveyv2.",
		},
	})
	if err != nil {
		logger.Logger.Fatal("Couldn't connect to postgres, " + err.Error())
	}

	db.Exec("CREATE SCHEMA IF NOT EXISTS " + viper.GetString("db_schema"))
	db.Exec("SET search_path TO " + viper.GetString("db_schema"))
	db.AutoMigrate(&models.Survey{}, &models.CollectionExercise{}, &models.CollectionInstrument{}, &models.Email{})

	router := mux.NewRouter()
	logger.Logger.Info("ras-rm-survey started")
	http.ListenAndServe(":8080", router)
}
