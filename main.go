package main

import (
	// std
	"fmt"
	"log"
	"net/http"

	// local
	"github.com/Vy4cheSlave/test-task-postgres/bash"
	"github.com/Vy4cheSlave/test-task-postgres/database"
	_ "github.com/Vy4cheSlave/test-task-postgres/docs"
	"github.com/Vy4cheSlave/test-task-postgres/handlers"

	// web
	"github.com/swaggo/http-swagger/v2"
)

const (
	// host=test-task-db
	DATABASE_URL string = "user=psql password=psql host=test-task-db port=5432 dbname=test-task-db"
	NUMBER_ATTEMPTS_TO_CONNECT_TO_DB uint = 5
	PORT string = "8080"
)

//	@title			bash API
//	@version		1.0
//	@license.name	Apache 2.0
func main() {
	var dbInstance database.DBWorker
	dbInstance, err := database.ConnectToDB(DATABASE_URL, NUMBER_ATTEMPTS_TO_CONNECT_TO_DB)
	if err != nil {
		log.Fatalln(err)
	}
	defer dbInstance.Close()
	log.Println("succsesfully connect to database")

	restApi := handlers.RestApi{}
	sh := bash.BashCommands{}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /swagger/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://localhost:%v/swagger/doc.json", PORT)),
		))
	mux.HandleFunc("POST /bash/create-command", 
		restApi.CreateNewCommandHandler(dbInstance, sh))
	mux.HandleFunc("GET /bash/get-commands/{id}", 
		restApi.GettingSingleCommandHandler(dbInstance))
	mux.HandleFunc("GET /bash/get-commands", 
		restApi.GettingListCommandsHandler(dbInstance))

	log.Printf("starting listen and serve Url = localhost:%v\n", PORT)
	if err := http.ListenAndServe(":" + PORT, mux); err != nil {
		log.Fatalln(err)
	}
}