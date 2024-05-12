package handlers

import (
	// std
	"encoding/json"
	"fmt"
	"strconv"
	"context"
	"io"
	"log"
	"net/http"
	// "sync"
	// "os/exec"
	// local
	"github.com/Vy4cheSlave/test-task-postgres/database"
	_ "github.com/Vy4cheSlave/test-task-postgres/docs"
	"github.com/Vy4cheSlave/test-task-postgres/bash"
)

//go:generate mockgen -source=handlers.go -destination=mock/mock.go

type RestApiWorker interface {
	CreateNewCommandHandler(database.DBWorker, bash.BashCommandsWorker) func(http.ResponseWriter, *http.Request)
	GettingSingleCommandHandler(database.DBWorker) func(http.ResponseWriter, *http.Request)
	GettingListCommandsHandler(database.DBWorker) func(http.ResponseWriter, *http.Request)
}

type RestApi struct{}

func closeHandlerWithErr(w http.ResponseWriter, err error) {
	log.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
}

//	@Tags		/bash/
//	@Accept		json
//	@Produce	json
//	@Param		new_command	body	bash.ReqCreateNewCommandBody	true	"input bash string"
//	@Router		/bash/create-command [post]
func (restApi RestApi) CreateNewCommandHandler(db database.DBWorker, sh bash.BashCommandsWorker) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var inputStruct bash.ReqCreateNewCommandBody
		defer r.Body.Close()
		buf, err := io.ReadAll(r.Body)
		if err != nil {
			closeHandlerWithErr(w, fmt.Errorf("read request body error: %v", err))
			return
		}
		if err := json.Unmarshal(buf, &inputStruct); err != nil {
			closeHandlerWithErr(w, fmt.Errorf("json unmarshal error: %v", err))
			return
		}

		isErrorOnChannel := false
		sliceCommands, err := sh.ExecCommands(&inputStruct)
		if err != nil {
			log.Println(err)
			isErrorOnChannel = true
		} 
		if sliceCommands == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err = db.CreateNewCommandsQuery(*sliceCommands, context.Background()); err != nil {
			log.Printf("database query error: %v\n", err)
		}

		if isErrorOnChannel {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		w.Header().Set("сontent-type", "application/json; charset=UTF-8")
		if err := json.NewEncoder(w).Encode(sliceCommands); err != nil {
			log.Printf("json encode error: %v\n", err)
		}
	}
}

//	@Tags		/bash/
//	@Produce	json
//	@Router		/bash/get-commands [get]
func (restApi RestApi) GettingListCommandsHandler(db database.DBWorker) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		commands, err := db.GettingListCommandsQuery(context.Background()) 
		if err != nil {
			closeHandlerWithErr(w, fmt.Errorf("database query error: %v", err))
			return
		} else {
			w.WriteHeader(http.StatusOK)
		}
		w.Header().Set("сontent-type", "application/json; charset=UTF-8")
		json.NewEncoder(w).Encode(commands)
	}
}

//	@Tags		/bash/
//	@Produce	json
//	@Param		id	path	uint	true	"uint without 0"	minimum(1)
//	@Router		/bash/get-commands/{id} [get]
func (restApi RestApi) GettingSingleCommandHandler(db database.DBWorker) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		pathVal, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			closeHandlerWithErr(w, fmt.Errorf("pathValue is not a number: %v", err))
			return
		}
		if pathVal <= 0 {
			closeHandlerWithErr(w, fmt.Errorf("pathValue isn't positive"))
			return
		}
		command, err := db.GettingSingleCommandQuery(uint(pathVal), context.Background())
		if err != nil {
			closeHandlerWithErr(w, fmt.Errorf("database query error: %v", err))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("сontent-type", "application/json; charset=UTF-8")
		json.NewEncoder(w).Encode(command)
	}
}