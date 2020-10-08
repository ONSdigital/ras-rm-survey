package main

import (
	"fmt"
	"net/http"
	"log"
	"encoding/json"
	"github.com/spf13/viper"
	"github.com/gorilla/mux"
)

func showInfo(w http.ResponseWriter, r *http.Request) {
    serviceInfo := infoResponse{Name: viper.GetString("service_name"), AppVersion: viper.GetString("app_version")}
    json.NewEncoder(w).Encode(serviceInfo)
}

func showHealth(w http.ResponseWriter, r *http.Request) {
    //This endpoint returns a dummy response for now
    healthInfo := healthResponse{Database: viper.GetString("dummy_health_database"), Rabbitmq: viper.GetString("dummy_health_rabbitmq")}
    json.NewEncoder(w).Encode(healthInfo)
}

func main() {
	router := mux.NewRouter()

	viper.AutomaticEnv()
	setDefaults()
	fmt.Println("Application started")

	router.HandleFunc("/info", showInfo)
    router.HandleFunc("/health", showHealth)
    log.Fatal(http.ListenAndServe(":8080", router))
}
