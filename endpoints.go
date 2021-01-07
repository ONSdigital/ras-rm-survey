package main

import (
	"encoding/json"
	"io/ioutil"
	"fmt"
	"net/http"
	"time"
	"database/sql"
	"strings"

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

    fmt.Println(db)

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
            sb.WriteString(" AND s.survey_ref='")
            sb.WriteString(queryParams.Get("surveyRef"))
            sb.WriteString("'")
        case "shortName":
            sb.WriteString(" AND s.short_name='")
            sb.WriteString(queryParams.Get("shortName"))
            sb.WriteString("'")
        case "longName":
            sb.WriteString(" AND s.long_name='")
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

    queryString := "SELECT s.survey_ref, s.short_name, s.long_name, s.legal_basis, s.survey_mode FROM surveyv2.survey" + sb.String()

    rows, err := db.Query(queryString)

    //put collection exercise here???
    //survey = models.Survey{CollectionExercise: ["Collection exercise goes here"]}

    //data := rows.Scan(&survey.Reference, &survey.ShortName, &survey.LongName, &survey.LegalBasis, &survey.SurveyMode)

    if err == sql.ErrNoRows {
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

    if err != nil {
        http.Error(w, "get survey query failed", http.StatusInternalServerError)
        return
    }

    data, err := json.Marshal(rows)
    if err != nil {
        http.Error(w, "Failed to marshal survey JSON", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusOK)
    w.Write(data)

    //vars := mux.Vars(r)
    //ref, found := vars["survey_ref"]
    //sName, found := vars["short_name"]
    //lName, found := vars["long_name"]
    //survey := new(Survey)
    //surveyRow := api.GetSurveyStmt.QueryRow(ref, sName, lName)
    //err := surveyRow.Scan()

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

    stmt, err := db.Prepare("INSERT INTO test_schema.survey VALUES($#)")
    if err != nil {
        http.Error(w, "SQL statement not prepared", http.StatusInternalServerError)
        return
    }

    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Couldn't read message body", http.StatusInternalServerError)
        return
    }

    keyVal := make(map[string]string)
    json.Unmarshal(body, &keyVal)
    shortName := keyVal["shortName"]
    longName := keyVal["longName"]
    legalBasis := keyVal["legalBasis"]
    surveyMode := keyVal["surveyMode"]

    // Generate a UUID to uniquely identify the new survey
    surveyRef, err := uuid.NewV4()
    if err != nil {
        http.Error(w, "Error generating random uuid", http.StatusInternalServerError)
        return
    }

    _, err = stmt.Exec(surveyRef, shortName, longName, legalBasis, surveyMode)
    if err != nil {
        http.Error(w, "SQL statement error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusCreated)
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
            sb.WriteString(" AND s.survey_ref='")
            sb.WriteString(queryParams.Get("surveyRef"))
            sb.WriteString("'")
        case "shortName":
            sb.WriteString(" AND s.short_name='")
            sb.WriteString(queryParams.Get("shortName"))
            sb.WriteString("'")
        case "longName":
            sb.WriteString(" AND s.long_name='")
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

    ref := queryParams.Get("surveyRef")

    queryString := "SELECT s.survey_ref, s.short_name, s.long_name, s.legal_basis, s.survey_mode FROM surveyv2.survey WHERE survey_ref = $1"

    rows, err := db.Query(queryString, ref)

    //put collection exercise here???
    //survey = models.Survey{CollectionExercise: ["Collection exercise goes here"]}

    //data := rows.Scan(&survey.Reference, &survey.ShortName, &survey.LongName, &survey.LegalBasis, &survey.SurveyMode)

    if err == sql.ErrNoRows {
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

    if err != nil {
        http.Error(w, "get survey query failed", http.StatusInternalServerError)
        return
    }

    data, err := json.Marshal(rows)
    if err != nil {
        http.Error(w, "Failed to marshal survey JSON", http.StatusInternalServerError)
        return
    }

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

    _, err = stmt.Exec(params["surveyRef"])
    if err != nil {
        http.Error(w, "SQL statement error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusOK)
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

    keyVal := make(map[string]string)
    json.Unmarshal(body, &keyVal)
    shortName := keyVal["shortName"]
    longName := keyVal["longName"]
    legalBasis := keyVal["legalBasis"]
    surveyMode := keyVal["surveyMode"]

    _, err = stmt.Exec(shortName, longName, legalBasis, surveyMode, params["survey_ref"])
    if err != nil {
        http.Error(w, "SQL statement error", http.StatusInternalServerError)
        return
    }

}