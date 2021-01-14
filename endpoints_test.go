package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"bytes"
	//"fmt"
//"github.com/ONSdigital/ras-rm-survey/logger"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ONSdigital/ras-rm-survey/models"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	// Yes, this import is weird but the mySQL driver offers passing a mock sql.DB and the postgres one doesn't.
)

var router *mux.Router
var resp *httptest.ResponseRecorder

var searchSurveyQueryColumns = []string{"survey_ref", "short_name", "long_name", "legal_basis", "survey_mode"}

var findSurveyQuery = "SELECT (.+) FROM*"
var postSurveyExec = "INSERT INTO (.+)*"

func setup() {
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

	assert.Equal(t, "UP 100ms", health.Database)    //This assertion statement randomly fails from time to time
	assert.Equal(t, viper.GetString("dummy_health_rabbitmq"), health.RabbitMQ)
}

func TestGetSurveyEndpoint (t *testing.T) {
    setup()

    var mock sqlmock.Sqlmock
    var err error

    db, mock, err = sqlmock.New()
    if err != nil {
        t.Fatal("Error setting up an SQL mock" + err.Error())
    }

    returnRows := mock.NewRows(searchSurveyQueryColumns)

    returnRows.AddRow("8eb7bdf5-92c2-4c52-8cc8-8f6525301bc5", "TS", "Test Survey", "Test Legal Basis", "Test Survey Mode")

    mock.ExpectQuery(findSurveyQuery).WillReturnRows(returnRows)

    req := httptest.NewRequest("GET", "/survey?shortName=TS", nil)
    router.ServeHTTP(resp, req)

    var surveys []models.Survey

    err = json.NewDecoder(resp.Body).Decode(&surveys)
    if err != nil {
        t.Fatal("Error decoding JSON response from 'GET /survey', ", err.Error())
    }

    assert.Equal(t, http.StatusOK, resp.Code)
    assert.Equal(t, surveys[0].Reference, "8eb7bdf5-92c2-4c52-8cc8-8f6525301bc5")

}

/*

func TestPostSurveyEndpoint2 (t *testing.T) {
    setup()


    var jsonStr = []byte(`{"shortName":"NEWPOST3333","longName":"postsurvey","legalBasis":"Ltest2","surveyMode":"aTEST2"}`)


    req := httptest.NewRequest("POST", "/survey", bytes.NewReader(jsonStr))
    router.ServeHTTP(resp, req)

    assert.Equal(t, http.StatusCreated, resp.Code)

    var survey models.Survey

    err := json.NewDecoder(resp.Body).Decode(&survey)
    if err != nil {
        t.Fatal("Error decoding JSON response from 'POST /survey', ", err.Error())
    }

    assert.Equal(t, survey.ShortName, "TS")

}

*/



func TestPostSurveyEndpoint (t *testing.T) {
    setup()

    var mock sqlmock.Sqlmock
    var err error

    db, mock, err = sqlmock.New()
    if err != nil {
        t.Fatal("Error setting up an SQL mock" + err.Error())
    }

    var jsonStr = []byte(`{"shortName":"NEWPOST3333","longName":"postsurvey","legalBasis":"Ltest2","surveyMode":"aTEST2"}`)

   //mock.ExpectPrepare(postSurveyExec).ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))

    //mock.ExpectPrepare("INSERT INTO test_schema.survey VALUES($1, $2, $3, $4, $5)")

    mock.ExpectPrepare(postSurveyExec)

    req := httptest.NewRequest("POST", "/survey", bytes.NewReader(jsonStr))
    router.ServeHTTP(resp, req)

    assert.Equal(t, http.StatusCreated, resp.Code)

    var survey models.Survey

    err = json.NewDecoder(resp.Body).Decode(&survey)
    if err != nil {
        t.Fatal("Error decoding JSON response from 'POST /survey', ", err.Error())
    }

    assert.Equal(t, survey.ShortName, "TS")

}