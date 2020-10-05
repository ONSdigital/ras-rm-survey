package main

import (
	"fmt"
	"net/http"
	"log"
	"encoding/json"
	"github.com/spf13/viper"
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

func showInfo(w http.ResponseWriter, r *http.Request) {
    vi := viper.New()
    vi.SetConfigFile("_infra/helm/ras-rm-survey/Chart.yaml")
    vi.ReadInConfig()

    serviceInfo := infoResponse{Name: vi.GetString("Name"), AppVersion: vi.GetString("AppVersion")}
    json.NewEncoder(w).Encode(serviceInfo)
}

func showHealth(w http.ResponseWriter, r *http.Request) {
    //This endpoint returns a dummy response for now
    healthInfo := healthResponse{Database: "UP 100ms", Rabbitmq: "DOWN"}
    json.NewEncoder(w).Encode(healthInfo)
}

func main() {
	router := mux.NewRouter()
	fmt.Println("Application started")

	http.HandleFunc("/info", showInfo)
    router.HandleFunc("/health", showHealth)
    log.Fatal(http.ListenAndServe(":8080", router))
}
