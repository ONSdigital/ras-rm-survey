package models

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
	    ID                      string      `json:"id"`
        SurveyRef               string      `json:"surveyRef"`
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