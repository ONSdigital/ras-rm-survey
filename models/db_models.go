package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/ONSdigital/ras-rm-survey/logger"
	"github.com/google/uuid"
	"github.com/jinzhu/inflection"
	"github.com/stoewer/go-strcase"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
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
		ExerciseUUID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
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
		InstrumentUUID uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
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

// TableName sets the table name, including the schema
func (ns PostgresStrategy) TableName(table string) string {
	logger.Logger.Infow("Calling TableName", "table", table)
	_, file, no, ok := runtime.Caller(1)
	if ok {
		logger.Logger.Infof("called from %s#%d\n", file, no)
	}
	return ns.Schema + "." + inflection.Plural(strcase.SnakeCase(table))
}

// ColumnName sets the column name
func (ns PostgresStrategy) ColumnName(table, column string) string {
	return strcase.SnakeCase(column)
}

// JoinTableName sets the name of a join table, including the schema
func (ns PostgresStrategy) JoinTableName(table string) string {
	return ns.Schema + "." + inflection.Plural(strcase.SnakeCase(table))
}

// RelationshipFKName sets the name of a FK constraint
func (ns PostgresStrategy) RelationshipFKName(rel schema.Relationship) string {
	return ns.Schema + "." + strings.Replace(fmt.Sprintf("fk_%s_%s", rel.Schema.Table, strcase.SnakeCase(rel.Name)), ".", "_", -1)
}

// CheckerName sets the name of a check constraint
func (ns PostgresStrategy) CheckerName(table, column string) string {
	return ns.Schema + "." + strings.Replace(fmt.Sprintf("chk_%s_%s", table, column), ".", "_", -1)
}

// IndexName sets the name of an index
func (ns PostgresStrategy) IndexName(table, column string) string {
	return ns.Schema + "." + strings.Replace(fmt.Sprintf("idx_%s_%s", table, column), ".", "_", -1)
}
