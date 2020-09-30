package main

import (
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
		CollectionInstruments []CollectionInstrument `gorm:"foreignKey:SurveyRef"`
	}

	// LegalBasis represents the possible legal bases for a survey
	LegalBasis string

	// CollectionInstrument represents collection instrument information
	CollectionInstrument struct {
		InstrumentID   uint `gorm:"primaryKey;autoIncrement"`
		SurveyRef      string
		InstrumentUUID uuid.UUID `gorm:"type:uuid"`
		Type           string
		Classifiers    string
	}
)

const ()

// BeforeCreate will create a UUID for the CollectionInstrument
func (ci *CollectionInstrument) BeforeCreate(tx *gorm.DB) (err error) {
	ci.InstrumentUUID, err = uuid.NewV4()
	return
}
