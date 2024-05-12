package bash

import (
	// std
	"fmt"
	"os/exec"
	"sync"
	"log"
	"io"
	// local
	"github.com/Vy4cheSlave/test-task-postgres/models"
)

//go:generate mockgen -source=bash.go -destination=mock/mock.go

type BashCommandsWorker interface {
	ExecCommands(*ReqCreateNewCommandBody) (*[]models.CommandsWithoutID, error)
	RunSubprocess(*sync.WaitGroup, *string, chan<- models.CommandsWithoutID, chan<- struct{}) 
}

type BashCommands struct{}

type ReqCreateNewCommandBody struct {
	BashStrings []string `json:"bash_strings"`
}

func (sh BashCommands) ExecCommands(inputStruct *ReqCreateNewCommandBody) (*[]models.CommandsWithoutID, error) {
	if inputStruct == nil {
		return nil, fmt.Errorf("func parameter error: the function parameter is nil")
	}

	var wg sync.WaitGroup
	outputCommand := make(chan models.CommandsWithoutID, len(inputStruct.BashStrings))
	errorChan := make(chan struct{})
	for _, commandString := range inputStruct.BashStrings {
		wg.Add(1)
		go sh.RunSubprocess(&wg, &commandString, outputCommand, errorChan)
	}
	wg.Wait()
	close(outputCommand)

	isErrorOnChannel := false
	sliceCommands := make([]models.CommandsWithoutID, 0)
	for range 2 {
		select {
		case <-errorChan:
			isErrorOnChannel = true
		default:
			for elem := range outputCommand {
				sliceCommands = append(sliceCommands, elem)
			}
		}
	}
	if isErrorOnChannel {
		return &sliceCommands, fmt.Errorf("run subprocess error")
	} else {
		return &sliceCommands, nil
	}
}

func (bash BashCommands) RunSubprocess(wg *sync.WaitGroup, input *string, output chan<- models.CommandsWithoutID, errorChan chan<- struct{}) {
	grepCmd := exec.Command("sh", "-c", *input)
	grepOut, err := grepCmd.StdoutPipe()
	if err != nil {
		log.Printf("stdout error: %v", err)
		errorChan <- struct{}{}
		wg.Done()
		return
	}
	grepErr, err := grepCmd.StderrPipe()
	if err != nil {
		log.Printf("stderr error: %v", err)
		errorChan <- struct{}{}
		wg.Done()
		return
	}

	grepCmd.Start()
	grepOutBytes, err := io.ReadAll(grepOut)
	if err != nil {
		log.Printf("stdout read error: %v", err)
		errorChan <- struct{}{}
		wg.Done()
		return
	}
	grepErrBytes, err := io.ReadAll(grepErr)
	if err != nil {
		log.Printf("stderr read error: %v", err)
		errorChan <- struct{}{}
		wg.Done()
		return
	}
	grepCmd.Wait()

	if len(grepOutBytes) != 0 {
		output <- models.CommandsWithoutID{Command: *input, IsError: false, Log: string(grepOutBytes)}
	} else {
		output <- models.CommandsWithoutID{Command: *input, IsError: true, Log: string(grepErrBytes)}
	}
	wg.Done()
}