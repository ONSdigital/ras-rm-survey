package main

import (
    "encoding/json"
    "github.com/spf13/viper"
    "net/http"
    "github.com/gorilla/mux"
)

type infoResponse struct{
    Name string `json:"name"`
    AppVersion string `json:"appVersion"`
}

type healthResponse struct{
    Database string `json:"database"`
    Rabbitmq string `json:"rabbitmq"`
}

func handleEndpoints(r *mux.Router) {
    r.HandleFunc("/info", showInfo)
    r.HandleFunc("/health", showHealth)
}

func showInfo(w http.ResponseWriter, r *http.Request) {
    serviceInfo := infoResponse{Name: viper.GetString("service_name"), AppVersion: viper.GetString("app_version")}
    json.NewEncoder(w).Encode(serviceInfo)
}

func showHealth(w http.ResponseWriter, r *http.Request) {
    //This endpoint returns a dummy response for now
    healthInfo := healthResponse{Database: viper.GetString("dummy_health_database"), Rabbitmq: viper.GetString("dummy_health_rabbitmq")}
    json.NewEncoder(w).Encode(healthInfo)
}