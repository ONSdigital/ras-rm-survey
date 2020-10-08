package main

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/ras-rm-survey/models"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

func handleEndpoints(r *mux.Router) {
	r.HandleFunc("/info", showInfo)
	r.HandleFunc("/health", showHealth)
}

func showInfo(w http.ResponseWriter, r *http.Request) {
	serviceInfo := models.Info{Name: viper.GetString("service_name"), AppVersion: viper.GetString("app_version")}
	json.NewEncoder(w).Encode(serviceInfo)
}

func showHealth(w http.ResponseWriter, r *http.Request) {
	//This endpoint returns a dummy response for now
	healthInfo := models.Health{Database: viper.GetString("dummy_health_database"), RabbitMQ: viper.GetString("dummy_health_rabbitmq")}
	json.NewEncoder(w).Encode(healthInfo)
}
