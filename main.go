package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/ONSdigital/ras-rm-survey/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	//"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
)

var db *sql.DB

const (
  host     = "localhost"
  port     = 5432
  user     = "adamwilkie"
  password = ""
  dbname   = "test_db"
)

func main() {
	viper.AutomaticEnv()
	setDefaults()
	err := logger.ConfigureLogger()
	if err != nil {
		log.Fatalln("Couldn't set up a logger, exiting", err)
	}

	logger.Logger.Info("Starting ras-rm-survey...")

/*

	dbURI := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", viper.GetString("db_host"), viper.GetString("db_port"), viper.GetString("db_name"), viper.GetString("db_username"), viper.GetString("db_password"))
	db, err = sql.Open("postgres", dbURI)
	if err != nil {
		logger.Logger.Fatal("Couldn't connect to postgres, " + err.Error())
	}

	dbMigrate()

*/

//localhost database stuff

  psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", host, port, user, dbname)

  db, err := sql.Open("postgres", psqlInfo)
  if err != nil {
  logger.Logger.Fatal("Couldn't connect to postgres, " + err.Error())
    panic(err)
  }
  //defer db.Close()

  err = db.Ping()
  if err != nil {
  logger.Logger.Fatal("Database ping failed, " + err.Error())
    panic(err)
  }

  fmt.Println("Connected to localhost database!")
//end of localhost database stuff

    //router := httprouter.New()
	router := mux.NewRouter()
	handleEndpoints(router)
	logger.Logger.Info("ras-rm-survey started")
	http.ListenAndServe(":8080", router)
}

func dbMigrate() {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal("Database connection for migration failed", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://db-migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatal("Database migration failed", err)
	}
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("Database migration failed ", err)
	}
}
