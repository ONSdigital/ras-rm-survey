package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type (
	// Survey represents survey information
	Survey struct {
		SurveyRef             string `gorm:"primaryKey"`
		ShortName             string
		LongName              string
		LegalBasis            string
		SurveyMode            string
		CollectionExercises   []CollectionExercise   `gorm:"foreignKey:SurveyRef"`
		CollectionInstruments []CollectionInstrument `gorm:"foreignKey:SurveyRef"`
	}

	// CollectionExercise represeents collection exercise information
	CollectionExercise struct {
		gorm.Model
		SurveyRef             string
		State                 string
		ExerciseUUID          uuid.UUID `gorm:"type:uuid"`
		PeriodName            string
		MPS                   sql.NullTime
		GoLive                sql.NullTime
		PeriodStart           sql.NullTime
		PeriodEnd             sql.NullTime
		Employment            sql.NullTime
		Return                sql.NullTime
		Emails                []Email                `gorm:"foreignKey:ExerciseID"`
		CollectionInstruments []CollectionInstrument `gorm:"many2many:associated_instruments;"`
	}

	// CollectionInstrument represents collection instrument information
	CollectionInstrument struct {
		gorm.Model
		SurveyRef      string
		InstrumentUUID uuid.UUID `gorm:"type:uuid"`
		Type           string
		Classifiers    JSONB
		SeftFilename   sql.NullString
	}

	// Email represents email trigger dates for a collection exercise
	Email struct {
		gorm.Model
		ExerciseID    uint
		Type          string
		TimeScheduled time.Time
	}

	// PostgresStrategy is a naming strategy that respects Postgres schemas
	PostgresStrategy struct {
		Schema string
	}

	// JSONB allows conversion into a PSQL JSONB column
	JSONB json.RawMessage
)

// Scan implements the driver.Scanner interface for JSONB
func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := json.RawMessage{}
	err := json.Unmarshal(bytes, &result)
	*j = JSONB(result)
	return err
}

// Value implements the driver.Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// BeforeCreate will create a UUID for the CollectionExercise
func (ce *CollectionExercise) BeforeCreate(tx *gorm.DB) (err error) {
	ce.ExerciseUUID, err = uuid.NewV4()
	return
}

// BeforeCreate will create a UUID for the CollectionInstrument
func (ci *CollectionInstrument) BeforeCreate(tx *gorm.DB) (err error) {
	ci.InstrumentUUID, err = uuid.NewV4()
	return
}
