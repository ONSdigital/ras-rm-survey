package models

import (
	"strconv"
	"time"
)

type (
	// Info represents the return values for GET /info
	Info struct {
		Name       string `json:"name"`
		AppVersion string `json:"appVersion"`
	}

	// Health represents the return values for GET /health
	Health struct {
		Database string `json:"database"`
		RabbitMQ string `json:"rabbitmq"`
	}

    // Error represents any erroneous response
    Error struct {
        Error string `json:"error"`
    }

	Survey struct {
        Reference               string      `json:"reference"`
        ShortName               string      `json:"shortName"`
        LongName                string      `json:"longName"`
        LegalBasis              string      `json:"legalBasis"`
        SurveyMode              string      `json:"surveyMode"`
    //    CollectionInstruments   []string    `json:"collectionInstruments"`  //This is a placeholder until CIs are integrated
    }

    RESTError struct {
    	Code      string `json:"code"`
    	Message   string `json:"message"`
    	Timestamp string `json:"timestamp"`
    }
)

    // NewRESTError returns a RESTError with the Timestamp field pre-populated.
    func NewRESTError(code string, message string) RESTError {
    	ts := strconv.Itoa(int(time.Now().UTC().Unix()))
    	return RESTError{Code: code, Message: message, Timestamp: ts}
    }
