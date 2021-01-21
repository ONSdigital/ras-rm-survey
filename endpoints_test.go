package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"bytes"
	"database/sql"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ONSdigital/ras-rm-survey/models"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	// Yes, this import is weird but the mySQL driver offers passing a mock sql.DB and the postgres one doesn't.
)

var router *mux.Router
var resp *httptest.ResponseRecorder

var searchSurveyQueryColumns = []string{"id", "survey_ref", "short_name", "long_name", "legal_basis", "survey_mode"}

var findSurveyQuery = "SELECT (.+) FROM*"
var postSurveyExec = "INSERT INTO (.+)*"
var deleteSurveyExec = "DELETE FROM (.+)*"
var updateSurveyExec = "UPDATE (.+)*"

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

	assert.Equal(t, "UP 100ms", health.Database)
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

    returnRows.AddRow("8eb7bdf5-92c2-4c52-8cc8-8f6525301bc5", "123", "TS", "Test Survey", "Test Legal Basis", "Test Survey Mode")

    mock.ExpectQuery(findSurveyQuery).WillReturnRows(returnRows)

    req := httptest.NewRequest("GET", "/survey?shortName=TS", nil)
    router.ServeHTTP(resp, req)

    var surveys []models.Survey

    err = json.NewDecoder(resp.Body).Decode(&surveys)
    if err != nil {
        t.Fatal("Error decoding JSON response from 'GET /survey', ", err.Error())
    }

    assert.Equal(t, http.StatusOK, resp.Code)
    assert.Equal(t, surveys[0].SurveyRef, "123")

}


func TestPostSurveyEndpoint (t *testing.T) {
    setup()

    var mock sqlmock.Sqlmock
    var err error

    db, mock, err = sqlmock.New()
    if err != nil {
        t.Fatal("Error setting up an SQL mock" + err.Error())
    }

    var jsonStr = []byte(`{"surveyRef":"156","shortName":"NEWPOST3333","longName":"postsurvey","legalBasis":"Ltest2","surveyMode":"aTEST2"}`)

    mock.ExpectBegin()
    mock.ExpectPrepare(postSurveyExec).ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectCommit()

    req := httptest.NewRequest("POST", "/survey", bytes.NewReader(jsonStr))
    router.ServeHTTP(resp, req)

    assert.Equal(t, http.StatusCreated, resp.Code)

    var survey models.Survey

    err = json.NewDecoder(resp.Body).Decode(&survey)
    if err != nil {
        t.Fatal("Error decoding JSON response from 'POST /survey', ", err.Error())
    }

    assert.Equal(t, survey.ShortName, "NEWPOST3333")
}

func TestGetSurveyByRefEndpoint(t *testing.T) {
    setup()

    var mock sqlmock.Sqlmock
    var err error

    db, mock, err = sqlmock.New()
    if err != nil {
        t.Fatal("Error setting up an SQL mock" + err.Error())
    }

    returnRows := mock.NewRows(searchSurveyQueryColumns)

    returnRows.AddRow("8eb7bdf5-92c2-4c52-8cc8-8f6525301bc5", "123", "TS", "Test Survey", "Test Legal Basis", "Test Survey Mode")

    mock.ExpectQuery(findSurveyQuery).WillReturnRows(returnRows)

    req := httptest.NewRequest("GET", "/survey/123", nil)
    router.ServeHTTP(resp, req)

    var surveys []models.Survey

    err = json.NewDecoder(resp.Body).Decode(&surveys)
    if err != nil {
        t.Fatal("Error decoding JSON response from 'GET /survey', ", err.Error())
    }

    assert.Equal(t, http.StatusOK, resp.Code)
    assert.Equal(t, surveys[0].SurveyRef, "123")
}

func TestDeleteSurveyEndpoint (t *testing.T) {
    setup()

    var mock sqlmock.Sqlmock
    var err error

    db, mock, err = sqlmock.New()
    if err != nil {
        t.Fatal("Error setting up an SQL mock" + err.Error())
    }

    returnRows := mock.NewRows(searchSurveyQueryColumns)

    returnRows.AddRow("8eb7bdf5-92c2-4c52-8cc8-8f6525301bc5", "123", "TS", "Test Survey", "Test Legal Basis", "Test Survey Mode")

    mock.ExpectBegin()
    mock.ExpectQuery(findSurveyQuery).WillReturnRows(returnRows)
    mock.ExpectPrepare(deleteSurveyExec).ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectCommit()

    req := httptest.NewRequest("DELETE", "/survey/123", nil)
    router.ServeHTTP(resp, req)

    assert.Equal(t, http.StatusNoContent, resp.Code)
}

func TestUpdateSurveyEndpoint (t *testing.T) {
    setup()

    var mock sqlmock.Sqlmock
    var err error

    db, mock, err = sqlmock.New()
    if err != nil {
        t.Fatal("Error setting up an SQL mock" + err.Error())
    }

    returnRows := mock.NewRows(searchSurveyQueryColumns)

    returnRows.AddRow("8eb7bdf5-92c2-4c52-8cc8-8f6525301bc5", "123", "TS", "Test Survey", "Test Legal Basis", "Test Survey Mode")

    var jsonStr = []byte(`{"shortName":"NEWPOST3333","longName":"postsurvey","legalBasis":"Ltest2","surveyMode":"aTEST2"}`)

    mock.ExpectBegin()
    mock.ExpectQuery(findSurveyQuery).WillReturnRows(returnRows)
    mock.ExpectPrepare(updateSurveyExec).ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectCommit()

    req := httptest.NewRequest("PATCH", "/survey/123", bytes.NewReader(jsonStr))
    router.ServeHTTP(resp, req)

    assert.Equal(t, http.StatusOK, resp.Code)

    var survey models.Survey

    err = json.NewDecoder(resp.Body).Decode(&survey)
    if err != nil {
        t.Fatal("Error decoding JSON response from 'PATCH /survey', ", err.Error())
    }

    assert.Equal(t, survey.ShortName, "NEWPOST3333")
}

func TestDeleteSurveyEndpointReturns404WhenSurveyRefNotFound (t *testing.T) {
    setup()

    var mock sqlmock.Sqlmock
    var err error

    db, mock, err = sqlmock.New()
    if err != nil {
        t.Fatal("Error setting up an SQL mock" + err.Error())
    }

    returnRows := mock.NewRows(searchSurveyQueryColumns)

    returnRows.AddRow("8eb7bdf5-92c2-4c52-8cc8-8f6525301bc5", "123", "TS", "Test Survey", "Test Legal Basis", "Test Survey Mode")

    mock.ExpectBegin()
    mock.ExpectQuery(findSurveyQuery).WillReturnError(sql.ErrNoRows)

    req := httptest.NewRequest("DELETE", "/survey/555", nil)
    router.ServeHTTP(resp, req)

    assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestUpdateSurveyEndpointReturns404WhenSurveyRefNotFound (t *testing.T) {
    setup()

    var mock sqlmock.Sqlmock
    var err error

    db, mock, err = sqlmock.New()
    if err != nil {
        t.Fatal("Error setting up an SQL mock" + err.Error())
    }

    returnRows := mock.NewRows(searchSurveyQueryColumns)

    returnRows.AddRow("8eb7bdf5-92c2-4c52-8cc8-8f6525301bc5", "123", "TS", "Test Survey", "Test Legal Basis", "Test Survey Mode")

    var jsonStr = []byte(`{"shortName":"NEWPOST3333","longName":"postsurvey","legalBasis":"Ltest2","surveyMode":"aTEST2"}`)

    mock.ExpectBegin()
    mock.ExpectQuery(findSurveyQuery).WillReturnError(sql.ErrNoRows)

    req := httptest.NewRequest("PATCH", "/survey/555", bytes.NewReader(jsonStr))
    router.ServeHTTP(resp, req)

    assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetSurveyEndpointReturns400WhenNoParametersProvided (t *testing.T) {
    setup()

    var mock sqlmock.Sqlmock
    var err error

    db, mock, err = sqlmock.New()
    if err != nil {
        t.Fatal("Error setting up an SQL mock" + err.Error())
    }

    returnRows := mock.NewRows(searchSurveyQueryColumns)

    returnRows.AddRow("8eb7bdf5-92c2-4c52-8cc8-8f6525301bc5", "123", "TS", "Test Survey", "Test Legal Basis", "Test Survey Mode")

    req := httptest.NewRequest("GET", "/survey", nil)
    router.ServeHTTP(resp, req)

    assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestGetSurveyEndpointReturns400WhenInvalidParametersProvided (t *testing.T) {
    setup()

    var mock sqlmock.Sqlmock
    var err error

    db, mock, err = sqlmock.New()
    if err != nil {
        t.Fatal("Error setting up an SQL mock" + err.Error())
    }

    returnRows := mock.NewRows(searchSurveyQueryColumns)

    returnRows.AddRow("8eb7bdf5-92c2-4c52-8cc8-8f6525301bc5", "123", "TS", "Test Survey", "Test Legal Basis", "Test Survey Mode")

    req := httptest.NewRequest("GET", "/survey?invalidParameter=12345", nil)
    router.ServeHTTP(resp, req)

    assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestPostSurveyEndpointReturns400WhenInvalidJSONBody (t *testing.T) {
    setup()

    var mock sqlmock.Sqlmock
    var err error

    db, mock, err = sqlmock.New()
    if err != nil {
        t.Fatal("Error setting up an SQL mock" + err.Error())
    }

    mock.ExpectBegin()

    var jsonStr = []byte(`invalidjson`)

    req := httptest.NewRequest("POST", "/survey", bytes.NewReader(jsonStr))
    router.ServeHTTP(resp, req)

    assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestGetSurveyByRefEndpointReturns404WhenSurveyRefNotFound(t *testing.T) {
    setup()

    var mock sqlmock.Sqlmock
    var err error

    db, mock, err = sqlmock.New()
    if err != nil {
        t.Fatal("Error setting up an SQL mock" + err.Error())
    }

    returnRows := mock.NewRows(searchSurveyQueryColumns)

    mock.ExpectQuery(findSurveyQuery).WillReturnRows(returnRows)

    req := httptest.NewRequest("GET", "/survey/555", nil)
    router.ServeHTTP(resp, req)

    assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestUpdateSurveyEndpointReturns400WhenInvalidJSONBody (t *testing.T) {
    setup()

    var mock sqlmock.Sqlmock
    var err error

    db, mock, err = sqlmock.New()
    if err != nil {
        t.Fatal("Error setting up an SQL mock" + err.Error())
    }

    returnRows := mock.NewRows(searchSurveyQueryColumns)

    returnRows.AddRow("8eb7bdf5-92c2-4c52-8cc8-8f6525301bc5", "123", "TS", "Test Survey", "Test Legal Basis", "Test Survey Mode")

    var jsonStr = []byte(`invalidjson`)

    mock.ExpectBegin()

    req := httptest.NewRequest("PATCH", "/survey/555", bytes.NewReader(jsonStr))
    router.ServeHTTP(resp, req)

    assert.Equal(t, http.StatusBadRequest, resp.Code)
}