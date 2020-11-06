DROP SCHEMA IF EXISTS surveyv2 CASCADE;

CREATE SCHEMA surveyv2;

CREATE TABLE IF NOT EXISTS surveyv2.survey (
    survey_ref uuid PRIMARY KEY,
    short_name text,
    long_name text,
    legal_basis text,
    survey_mode text
);

CREATE TABLE IF NOT EXISTS surveyv2.collection_exercise (
    exercise_id serial PRIMARY KEY,
    survey_ref uuid NOT NULL,
    state text,
    exercise_uuid uuid NOT NULL,
    period_name text,
    mps timestamp,
    go_live timestamp,
    period_start timestamp,
    period_end timestamp,
    employment timestamp,
    return timestamp,
    FOREIGN KEY (survey_ref) REFERENCES surveyv2.survey (survey_ref)
);

CREATE TABLE IF NOT EXISTS surveyv2.collection_instrument (
    instrument_id serial PRIMARY KEY,
    survey_ref uuid NOT NULL,
    instrument_uuid uuid NOT NULL,
    type text,
    classifiers jsonb,
    seft_filename text,
    FOREIGN KEY (survey_ref) REFERENCES surveyv2.survey (survey_ref)
);

CREATE TABLE IF NOT EXISTS surveyv2.associated_instruments (
    exercise_id int NOT NULL,
    instrument_id int NOT NULL,
    PRIMARY KEY (exercise_id, instrument_id),
    FOREIGN KEY (exercise_id) REFERENCES surveyv2.collection_exercise (exercise_id),
    FOREIGN KEY (instrument_id) REFERENCES surveyv2.collection_instrument (instrument_id)
);

CREATE TABLE IF NOT EXISTS surveyv2.email (
    email_id serial PRIMARY KEY,
    exercise_id int NOT NULL,
    type text,
    time_scheduled timestamp,
    FOREIGN KEY (exercise_id) REFERENCES surveyv2.collection_exercise (exercise_id)
);