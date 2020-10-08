package main

type infoResponse struct{
    Name string `json:"name"`
    AppVersion string `json:"appVersion"`
}

type healthResponse struct{
    Database string `json:"database"`
    Rabbitmq string `json:"rabbitmq"`
}