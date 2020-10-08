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
)
