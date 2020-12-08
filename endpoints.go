package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ONSdigital/ras-rm-survey/models"
	//"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"github.com/julienschmidt/httprouter"
)

func handleEndpoints(r *httprouter.Router) {
    r.GET("/info", showInfo)
    r.GET("/health", showHealth)
	//r.HandleFunc("/info", showInfo)
	//r.HandleFunc("/health", showHealth)
}

func showInfo(w http.ResponseWriter, r *http.Request) {
	serviceInfo := models.Info{Name: viper.GetString("service_name"), AppVersion: viper.GetString("app_version")}
	json.NewEncoder(w).Encode(serviceInfo)
}

func showHealth(w http.ResponseWriter, r *http.Request) {
	// Rabbit data is dummy until implemented
	dbStatus := "DOWN"
	start := time.Now()
	err := db.Ping()
	if err == nil {
		latency := time.Since(start)
		dbStatus = fmt.Sprintf("UP %s", latency.Truncate(time.Millisecond))
	}
	healthInfo := models.Health{Database: dbStatus, RabbitMQ: viper.GetString("dummy_health_rabbitmq")}
	json.NewEncoder(w).Encode(healthInfo)
}
