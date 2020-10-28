package main

import (
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/ras-rm-survey/logger"
	"go.uber.org/zap"
	"net/http"
	"time"

	"github.com/ONSdigital/ras-rm-survey/models"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

func handleEndpoints(r *mux.Router) {
	r.HandleFunc("/info", showInfo)
	r.HandleFunc("/health", showHealth)
	r.HandleFunc("/survey/{surveyRef}", getSurveyByRef)
}

func showInfo(w http.ResponseWriter, r *http.Request) {
	serviceInfo := models.Info{Name: viper.GetString("service_name"), AppVersion: viper.GetString("app_version")}
	json.NewEncoder(w).Encode(serviceInfo)
}

func showHealth(w http.ResponseWriter, r *http.Request) {
	// Rabbit data is dummy until implemented
	dbStatus := "DOWN"
	d, err := db.DB()
	if err == nil {
		start := time.Now()
		err = d.Ping()
		if err == nil {
			latency := time.Since(start)
			dbStatus = fmt.Sprintf("UP %s", latency.Truncate(time.Millisecond))
		}
	}
	healthInfo := models.Health{Database: dbStatus, RabbitMQ: viper.GetString("dummy_health_rabbitmq")}
	json.NewEncoder(w).Encode(healthInfo)
}

func getSurveyByRef(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	surveyRef := params["surveyRef"]
	logger.Logger.Info(surveyRef)
	var survey models.Survey
	result := db.First(&survey, surveyRef)
	if result.Error != nil {
		logger.Logger.Error("Something went wrong")
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		logger.Logger.Error("Something went wrong")
	}
	logger.Logger.Info("Got result", zap.Any("result", resultJson))
	w.Header().Set("Content-Type", "application/json")
	w.Write(resultJson) // TODO handle error
}
