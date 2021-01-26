package main

import (
	"encoding/json"
	"io/ioutil"
	"fmt"
	"net/http"
	"time"
	"database/sql"
	"strings"
	"strconv"

    "github.com/ONSdigital/ras-rm-survey/logger"
    "github.com/gofrs/uuid"
	"github.com/ONSdigital/ras-rm-survey/models"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

func handleEndpoints(r *mux.Router) {
	r.HandleFunc("/info", showInfo).Methods("GET")
	r.HandleFunc("/health", showHealth).Methods("GET")
	r.HandleFunc("/survey", getSurvey).Methods("GET")
	r.HandleFunc("/survey", postSurvey).Methods("POST")
	r.HandleFunc("/survey/{surveyRef}", getSurveyByRef).Methods("GET")
	r.HandleFunc("/survey/{surveyRef}", deleteSurveyByRef).Methods("DELETE")
	r.HandleFunc("/survey/{surveyRef}", updateSurveyByRef).Methods("PATCH")
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

    var paramArray []string
    var placeholderNumber int
    var sb strings.Builder

    sb.WriteString("SELECT id, survey_ref, short_name, long_name, legal_basis, survey_mode FROM " + viper.GetString("db_schema") + ".survey WHERE 1=1")

    for params := range queryParams {
        switch params {
        case "surveyRef":
            paramArray = append(paramArray, queryParams.Get("surveyRef"))
            placeholderNumber = placeholderNumber + 1
            sb.WriteString(" AND survey_ref = $" + strconv.Itoa(placeholderNumber))
        case "shortName":
            paramArray = append(paramArray, queryParams.Get("shortName"))
            placeholderNumber = placeholderNumber + 1
            sb.WriteString(" AND short_name = $" + strconv.Itoa(placeholderNumber))
        case "longName":
            paramArray = append(paramArray, queryParams.Get("longName"))
            placeholderNumber = placeholderNumber + 1
            sb.WriteString(" AND long_name = $" + strconv.Itoa(placeholderNumber))
        default:
            w.WriteHeader(http.StatusBadRequest)
            errorString := models.Error{
                Error: "Invalid query parameter " + params,
            }
            json.NewEncoder(w).Encode(errorString)
            return
        }
    }

    queryString := sb.String()
    var rows *sql.Rows
    var err error

    if len(paramArray) == 1 {
        rows, err = db.Query(queryString, paramArray[0])
    } else if len(paramArray) == 2 {
        rows, err = db.Query(queryString, paramArray[0], paramArray[1])
    } else {
        rows, err = db.Query(queryString, paramArray[0], paramArray[1], paramArray[2])
    }

    if err != nil {
        http.Error(w, "get survey query failed", http.StatusInternalServerError)
        return
    }

    var listOfSurveys = []models.Survey{}

    for rows.Next(){
        survey := models.Survey{}

        err = rows.Scan(
            &survey.ID,
            &survey.SurveyRef,
            &survey.ShortName,
            &survey.LongName,
            &survey.LegalBasis,
            &survey.SurveyMode,
        )
        if err != nil {
            http.Error(w, "Error scanning database rows", http.StatusInternalServerError)
            return
        }

        listOfSurveys = append(listOfSurveys, survey)
    }

    if len(listOfSurveys) == 0 {
        w.Header().Set("Content-Type", "application/json; charset=UTF-8")
        w.WriteHeader(http.StatusNotFound)

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

    tx, err := db.Begin()
    if err != nil {
        http.Error(w, "Error starting database transaction", http.StatusInternalServerError)
        return
    }
    defer tx.Rollback()

    stmt, err := db.Prepare("INSERT INTO " + viper.GetString("db_schema") + ".survey VALUES($1, $2, $3, $4, $5, $6)")
    if err != nil {
        http.Error(w, "SQL statement not prepared", http.StatusInternalServerError)
        return
    }

    defer stmt.Close()

    // Generate a UUID to uniquely identify the new survey
    newID, err := uuid.NewV4()
    if err != nil {
        http.Error(w, "Error generating random uuid", http.StatusInternalServerError)
        return
    }

    survey.ID = newID.String()

    _, err = stmt.Exec(survey.ID, survey.SurveyRef, survey.ShortName, survey.LongName, survey.LegalBasis, survey.SurveyMode)
    if err != nil {
        http.Error(w, "SQL statement error" + err.Error(), http.StatusInternalServerError)
        return
    }

    err = tx.Commit()
    if err != nil {
            http.Error(w, "Error committing database transaction", http.StatusInternalServerError)
            return
        }

    var js []byte
    js, err = json.Marshal(&survey)

    logger.Logger.Info("Successfully posted survey")
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusCreated)
    w.Write(js)
}

//Get survey using the parameter reference
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

    surveyRef := vars["surveyRef"]

    queryString := "SELECT id, survey_ref, short_name, long_name, legal_basis, survey_mode FROM " + viper.GetString("db_schema") + ".survey WHERE survey_ref = $1"

    rows, err := db.Query(queryString, surveyRef)

    if err != nil {
        http.Error(w, "get survey query failed", http.StatusInternalServerError)
        return
    }

    var listOfSurveys = []models.Survey{}

    for rows.Next(){
        survey := models.Survey{}

        err = rows.Scan(
            &survey.ID,
            &survey.SurveyRef,
            &survey.ShortName,
            &survey.LongName,
            &survey.LegalBasis,
            &survey.SurveyMode,
        )
        if err != nil {
                http.Error(w, "Error scanning database rows", http.StatusInternalServerError)
                return
            }


        listOfSurveys = append(listOfSurveys, survey)
    }

    if len(listOfSurveys) == 0 {
        w.Header().Set("Content-Type", "application/json; charset=UTF-8")
        w.WriteHeader(http.StatusNotFound)

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

    tx, err := db.Begin()
    if err != nil {
        http.Error(w, "Error starting database transaction", http.StatusInternalServerError)
        return
    }
    defer tx.Rollback()

    results := db.QueryRow("SELECT * FROM " + viper.GetString("db_schema") + ".survey WHERE survey_ref = $1" , params["surveyRef"])
    var tempSurvey models.Survey
    err = results.Scan(&tempSurvey.ID, &tempSurvey.SurveyRef, &tempSurvey.ShortName, &tempSurvey.LongName, &tempSurvey.LegalBasis, &tempSurvey.SurveyMode)
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "Survey reference not found", http.StatusNotFound)
            return
        }
        http.Error(w, "Check query failed", http.StatusInternalServerError)
        return
    }

    stmt, err := db.Prepare("DELETE FROM " + viper.GetString("db_schema") + ".survey WHERE survey_ref = $1")
    if err != nil {
        http.Error(w, "SQL statement not prepared", http.StatusInternalServerError)
        return
    }

    defer stmt.Close()

    _, err = stmt.Exec(params["surveyRef"])
    if err != nil {
        http.Error(w, "SQL statement error" + err.Error(), http.StatusInternalServerError)
        return
    }

    err = tx.Commit()
    if err != nil {
        http.Error(w, "Error committing database transaction", http.StatusInternalServerError)
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

    if survey.ShortName == "" && survey.LongName == "" && survey.LegalBasis == "" && survey.SurveyMode == "" {
        http.Error(w, "No values to update", http.StatusBadRequest)
        return
    }

    tx, err := db.Begin()
    if err != nil {
        http.Error(w, "Error starting database transaction", http.StatusInternalServerError)
        return
    }
    defer tx.Rollback()

    selectQuery := "SELECT * FROM " + viper.GetString("db_schema") + ".survey WHERE survey_ref = $1"

    results := db.QueryRow(selectQuery, params["surveyRef"])
    var tempSurvey models.Survey
    err = results.Scan(&tempSurvey.ID, &tempSurvey.SurveyRef, &tempSurvey.ShortName, &tempSurvey.LongName, &tempSurvey.LegalBasis, &tempSurvey.SurveyMode)
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "Survey reference not found", http.StatusNotFound)
            return
        }
        http.Error(w, "Check query failed", http.StatusInternalServerError)
        return
    }

    newShortName := "short_name"
    newLongName := "long_name"
    newLegalBasis := "legal_basis"
    newSurveyMode := "survey_mode"

    if survey.ShortName != "" {
        newShortName = survey.ShortName
    } else {
        newShortName = tempSurvey.ShortName
    }
    if survey.LongName != "" {
        newLongName = survey.LongName
    } else {
        newLongName = tempSurvey.LongName
    }
    if survey.LegalBasis != "" {
        newLegalBasis = survey.LegalBasis
    } else {
        newLegalBasis = tempSurvey.LegalBasis
    }
    if survey.SurveyMode != "" {
        newSurveyMode = survey.SurveyMode
    } else {
        newSurveyMode = tempSurvey.SurveyMode
    }

    updateQuery := "UPDATE " + viper.GetString("db_schema") + ".survey SET short_name = $1, long_name = $2, legal_basis = $3, survey_mode = $4 WHERE survey_ref = $5"

    stmt, err := db.Prepare(updateQuery)
    if err != nil {
        http.Error(w, "SQL statement not prepared", http.StatusInternalServerError)
        return
    }

    defer stmt.Close()

    survey.SurveyRef = params["surveyRef"]

    _, err = stmt.Exec(newShortName, newLongName, newLegalBasis, newSurveyMode, survey.SurveyRef)
    if err != nil {
        http.Error(w, "SQL statement error" + err.Error(), http.StatusInternalServerError)
        return
    }

    err = tx.Commit()
    if err != nil {
            http.Error(w, "Error committing database transaction", http.StatusInternalServerError)
            return
        }

    //redo the select query to show what the survey looks like now
    newResults := db.QueryRow(selectQuery, survey.SurveyRef)
    err = newResults.Scan(&survey.ID, &survey.SurveyRef, &survey.ShortName, &survey.LongName, &survey.LegalBasis, &survey.SurveyMode)
        if err != nil {
            http.Error(w, "Second search query failed", http.StatusInternalServerError)
            return
        }

    var js []byte
    js, err = json.Marshal(&survey)

    logger.Logger.Info("Successfully updated survey")
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusOK)
    w.Write(js)
}