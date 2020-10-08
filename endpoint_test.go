package main

import (
    "encoding/json"
    "io/ioutil"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/spf13/viper"
    "github.com/stretchr/testify/assert"
    "github.com/gorilla/mux"
)

var router *mux.Router
var resp *httptest.ResponseRecorder

func setup(){
    setDefaults()
    router = mux.NewRouter()
    resp = httptest.NewRecorder()
    handleEndpoints(router)
}

func TestInfoEndpoint(t *testing.T) {
    setup()

    req := httptest.NewRequest("GET", "/info", nil)
    router.ServeHTTP(resp, req)

    assert.Equal(t, http.StatusOK, resp.Code)

    body, _ := ioutil.ReadAll(resp.Body)

    var info infoResponse
    err := json.Unmarshal(body, &info)
    if err != nil {
        t.Fatal("Error decoding JSON response from 'GET /info', ", err.Error())
    }

    assert.Equal(t, viper.GetString("service_name"), info.Name)
}

func TestHealthEndpoint(t *testing.T) {
    setup()

    req := httptest.NewRequest("GET", "/health", nil)
    router.ServeHTTP(resp, req)

    assert.Equal(t, http.StatusOK, resp.Code)

    body, _ := ioutil.ReadAll(resp.Body)

    var health healthResponse
    err := json.Unmarshal(body, &health)
    if err != nil {
        t.Fatal("Error decoding JSON response from 'GET /health', ", err.Error())
    }

    assert.Equal(t, viper.GetString("dummy_health_database"), health.Database)
}