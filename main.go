package main

import (
	"fmt"
	"net/http"
	"log"
	"github.com/spf13/viper"
	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	viper.AutomaticEnv()
	setDefaults()
    handleEndpoints(router)
	fmt.Println("Application started")

    log.Fatal(http.ListenAndServe(":8080", router))
}
