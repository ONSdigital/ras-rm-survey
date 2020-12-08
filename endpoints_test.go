package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ONSdigital/ras-rm-survey/models"
	//"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/julienschmidt/httprouter"
	// Yes, this import is weird but the mySQL driver offers passing a mock sql.DB and the postgres one doesn't.
)

var router *httprouter.Router
var resp *httptest.ResponseRecorder

func setup() {
	setDefaults()
	router = httprouter.New()
	resp = httptest.NewRecorder()
	handleEndpoints(router)
}

func TestInfoEndpoint(t *testing.T) {
	setup()

	req := httptest.NewRequest("GET", "/info", nil)
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	body, _ := ioutil.ReadAll(resp.Body)

	var info models.Info
	err := json.Unmarshal(body, &info)
	if err != nil {
		t.Fatal("Error decoding JSON response from 'GET /info', ", err.Error())
	}

	assert.Equal(t, viper.GetString("service_name"), info.Name)
	assert.Equal(t, viper.GetString("app_version"), info.AppVersion)
}

func TestHealthEndpoint(t *testing.T) {
	setup()
	var mock sqlmock.Sqlmock
	var err error

	db, mock, err = sqlmock.New(sqlmock.MonitorPingsOption(true))

	mock.ExpectPing().WillDelayFor(100 * time.Millisecond)

	req := httptest.NewRequest("GET", "/health", nil)
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	body, _ := ioutil.ReadAll(resp.Body)

	var health models.Health
	err = json.Unmarshal(body, &health)
	if err != nil {
		t.Fatal("Error decoding JSON response from 'GET /health', ", err.Error())
	}

	assert.Equal(t, "UP 100ms", health.Database)
	assert.Equal(t, viper.GetString("dummy_health_rabbitmq"), health.RabbitMQ)
}
