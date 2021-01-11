package main

import (
	"encoding/json"
	"io/ioutil"
	"fmt"
	"net/http"
	"time"
	//"database/sql"
	"strings"

    "github.com/ONSdigital/ras-rm-survey/logger"
    "github.com/gofrs/uuid"
	"github.com/ONSdigital/ras-rm-survey/models"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	//"github.com/julienschmidt/httprouter"
)

func handleEndpoints(r *mux.Router) {
	r.HandleFunc("/info", showInfo).Methods("GET")
	r.HandleFunc("/health", showHealth).Methods("GET")
	r.HandleFunc("/survey", getSurvey).Methods("GET")
	r.HandleFunc("/survey", postSurvey).Methods("POST")
	r.HandleFunc("/survey/{reference}", getSurveyByRef).Methods("GET")
	r.HandleFunc("/survey/{reference}", deleteSurveyByRef).Methods("DELETE")
	r.HandleFunc("/survey/{reference}", updateSurveyByRef).Methods("PATCH")
}

func showInfo(w http.ResponseWriter, r *http.Request) {
	serviceInfo := models.Info{Name: viper.GetString("service_name"), AppVersion: viper.GetString("app_version")}
	json.NewEncoder(w).Encode(serviceInfo)
}

func showHealth(w http.ResponseWriter, r *http.Request) {
	// Rabbit data is dummy until implemented
	dbStatus := "DOWN"
	start := time.Now()
	err := db.Ping()
	if err == nil {
		latency := time.Since(start)
		dbStatus = fmt.Sprintf("UP %s", latency.Truncate(time.Millisecond))
	}
	healthInfo := models.Health{Database: dbStatus, RabbitMQ: viper.GetString("dummy_health_rabbitmq")}
	json.NewEncoder(w).Encode(healthInfo)
}

//Find survey by reference, short name, long name, or any combination of the three
func getSurvey(w http.ResponseWriter, r *http.Request) {

    //fmt.Println(db)

    if db == nil {
        w.WriteHeader(http.StatusInternalServerError)
        errorString := models.Error{
            Error: "Database connection could not be found",
        }
        json.NewEncoder(w).Encode(errorString)
        return
    }

    queryParams := r.URL.Query()
    if len(queryParams) == 0 {
        w.WriteHeader(http.StatusBadRequest)
        errorString := models.Error{
            Error: "No query parameters provided for search",
        }
        json.NewEncoder(w).Encode(errorString)
        return
    }

    var sb strings.Builder
    sb.WriteString(" WHERE 1=1")
    for k := range queryParams {
        switch k {
        case "surveyRef":
            sb.WriteString(" AND survey_ref='")
            sb.WriteString(queryParams.Get("surveyRef"))
            sb.WriteString("'")
        case "shortName":
            sb.WriteString(" AND short_name='")
            sb.WriteString(queryParams.Get("shortName"))
            sb.WriteString("'")
        case "longName":
            sb.WriteString(" AND long_name='")
            sb.WriteString(queryParams.Get("longName"))
            sb.WriteString("'")
        default:
            w.WriteHeader(http.StatusBadRequest)
            errorString := models.Error{
                Error: "Invalid query parameter " + k,
            }
            json.NewEncoder(w).Encode(errorString)
            return
        }
    }

    //change test_schema.survey to surveyv2.survey when done

    queryString := "SELECT survey_ref, short_name, long_name, legal_basis, survey_mode FROM test_schema.survey" + sb.String()

    rows, err := db.Query(queryString)

    if err != nil {
        http.Error(w, "get survey query failed", http.StatusInternalServerError)
        return
    }

    var listOfSurveys = []models.Survey{}

    for rows.Next(){
        survey := models.Survey{}

        rows.Scan(
            &survey.Reference,
            &survey.ShortName,
            &survey.LongName,
            &survey.LegalBasis,
            &survey.SurveyMode,
        )

        listOfSurveys = append(listOfSurveys, survey)
    }

    if len(listOfSurveys) == 0 {
        re := models.NewRESTError("404", "Survey not found")
        data, err := json.Marshal(re)
        if err != nil {
            http.Error(w, "Error marshaling NewRestError JSON", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json; charset=UTF-8")
        w.WriteHeader(http.StatusNotFound)
        w.Write(data)

        return
    }

    data, err := json.Marshal(listOfSurveys)
    if err != nil {
        http.Error(w, "Failed to marshal survey JSON", http.StatusInternalServerError)
        return
    }

    logger.Logger.Info("Successfully retrieved survey")
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusOK)
    w.Write(data)
}

//Create survey based on JSON request
func postSurvey (w http.ResponseWriter, r *http.Request) {

    if db == nil {
        w.WriteHeader(http.StatusInternalServerError)
        errorString := models.Error{
            Error: "Database connection could not be found",
        }
        json.NewEncoder(w).Encode(errorString)
        return
    }

    //replace with schemav2 when done

    stmt, err := db.Prepare("INSERT INTO test_schema.survey VALUES($1, $2, $3, $4, $5)")
    if err != nil {
        http.Error(w, "SQL statement not prepared", http.StatusInternalServerError)
        return
    }

    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Couldn't read message body", http.StatusInternalServerError)
        return
    }

    var survey models.Survey

    err = json.Unmarshal(body, &survey)

    if err != nil {
        http.Error(w, "Error unmarshalling JSON", http.StatusBadRequest)
        return
    }

    // Generate a UUID to uniquely identify the new survey
    surveyRef, err := uuid.NewV4()
    if err != nil {
        http.Error(w, "Error generating random uuid", http.StatusInternalServerError)
        return
    }

    survey.Reference = surveyRef.String()

    _, err = stmt.Exec(survey.Reference, survey.ShortName, survey.LongName, survey.LegalBasis, survey.SurveyMode)
    if err != nil {
        http.Error(w, "SQL statement error", http.StatusInternalServerError)
        return
    }

    var js []byte
    js, err = json.Marshal(&survey)

    logger.Logger.Info("Successfully posted survey")
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusCreated)
    w.Write(js)
}

//Create survey based on JSON request
func getSurveyByRef (w http.ResponseWriter, r *http.Request) {

    if db == nil {
        w.WriteHeader(http.StatusInternalServerError)
        errorString := models.Error{
            Error: "Database connection could not be found",
        }
        json.NewEncoder(w).Encode(errorString)
        return
    }

    vars := mux.Vars(r)

    surveyRef := vars["reference"]

    //change test_schema.survey to surveyv2.survey when done

    queryString := "SELECT survey_ref, short_name, long_name, legal_basis, survey_mode FROM test_schema.survey WHERE survey_ref = $1"

    rows, err := db.Query(queryString, surveyRef)

    if err != nil {
        http.Error(w, "get survey query failed", http.StatusInternalServerError)
        return
    }

    var listOfSurveys = []models.Survey{}

    for rows.Next(){
        survey := models.Survey{}

        rows.Scan(
            &survey.Reference,
            &survey.ShortName,
            &survey.LongName,
            &survey.LegalBasis,
            &survey.SurveyMode,
        )

        listOfSurveys = append(listOfSurveys, survey)
    }

    if len(listOfSurveys) == 0 {
        re := models.NewRESTError("404", "Survey not found")
        data, err := json.Marshal(re)
        if err != nil {
            http.Error(w, "Error marshaling NewRestError JSON", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json; charset=UTF-8")
        w.WriteHeader(http.StatusNotFound)
        w.Write(data)

        return
    }

    data, err := json.Marshal(listOfSurveys)
    if err != nil {
        http.Error(w, "Failed to marshal survey JSON", http.StatusInternalServerError)
        return
    }

    logger.Logger.Info("Successfully retrieved survey from reference")
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusOK)
    w.Write(data)

}

//Delete survey based on given reference
func deleteSurveyByRef (w http.ResponseWriter, r *http.Request) {

    if db == nil {
        w.WriteHeader(http.StatusInternalServerError)
        errorString := models.Error{
            Error: "Database connection could not be found",
        }
        json.NewEncoder(w).Encode(errorString)
        return
    }

    var params = mux.Vars(r)

    stmt, err := db.Prepare("DELETE FROM test_schema.survey WHERE survey_ref = $1")
    if err != nil {
        http.Error(w, "SQL statement not prepared", http.StatusInternalServerError)
        return
    }

    _, err = stmt.Exec(params["reference"])
    if err != nil {
        http.Error(w, "SQL statement error", http.StatusInternalServerError)
        return
    }

    logger.Logger.Info("Successfully deleted survey")
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusNoContent)
}

//Update survey based on JSON request
func updateSurveyByRef (w http.ResponseWriter, r *http.Request) {
    if db == nil {
        w.WriteHeader(http.StatusInternalServerError)
        errorString := models.Error{
            Error: "Database connection could not be found",
        }
        json.NewEncoder(w).Encode(errorString)
        return
    }

    var params = mux.Vars(r)

    stmt, err := db.Prepare("UPDATE test_schema.survey SET short_name = $1, long_name = $2, legal_basis = $3, survey_mode = $4 WHERE survey_ref = $5")
    if err != nil {
        http.Error(w, "SQL statement not prepared", http.StatusInternalServerError)
        return
    }

    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Couldn't read message body", http.StatusInternalServerError)
        return
    }

    var survey models.Survey

    err = json.Unmarshal(body, &survey)
    if err != nil {
        http.Error(w, "Error unmarshalling JSON", http.StatusBadRequest)
        return
    }

    survey.Reference = params["reference"]
    _, err = stmt.Exec(survey.ShortName, survey.LongName, survey.LegalBasis, survey.SurveyMode, survey.Reference)
    if err != nil {
        http.Error(w, "SQL statement error", http.StatusInternalServerError)
        return
    }

    var js []byte
    js, err = json.Marshal(&survey)

    logger.Logger.Info("Successfully updated survey")
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusOK)
    w.Write(js)
}