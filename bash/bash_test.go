package bash

import (
	"fmt"
	"sync"
	"testing"

	"github.com/Vy4cheSlave/test-task-postgres/models"
	// mock_bash "github.com/Vy4cheSlave/test-task-postgres/bash/mock"
)

var bash BashCommands = BashCommands{}

func TestRunSubprocess(t *testing.T) {
	var wg sync.WaitGroup
	outputCommand := make(chan models.CommandsWithoutID, 1)
	errorChan := make(chan struct{}, 1)

	var tests = []struct {
        inputString string
        want bool
    }{
        {"echo hello world!!!", false},
        {"pws", true},
    }
	
	for _, tt := range tests {
        testname := fmt.Sprintf("%v", tt.inputString)
        t.Run(testname, func(t *testing.T) {
			wg.Add(1)
            bash.RunSubprocess(&wg, &tt.inputString, outputCommand, errorChan)
			select {
			case <- errorChan:
				t.Errorf("Subprocess execution error")
			case resultValue := <- outputCommand:
				if resultValue.IsError != tt.want {
					t.Errorf("Subprocess error: the behavior of the function does not meet expectations\ngot %v, want %v", resultValue.IsError, tt.want)
				}
			}
        })
    }
}

func TestExecCommands(t *testing.T) {
	var tests = []struct {
		testName string
        inputStruct *ReqCreateNewCommandBody
        wantResult *models.CommandsWithoutID
    }{
        {
			"inputStructNil",
			nil,
			nil,
		},
        {
			"inputStructNotNil",
			&ReqCreateNewCommandBody {
				BashStrings: []string{"echo hello world!!!"},
			},
			&models.CommandsWithoutID{
				Command: "echo hello world!!!",
				IsError: false,
				Log: "hello world!!!\n",
			},
		},
    }

	for _, tt := range tests {
        testname := fmt.Sprintf("%v", tt.testName)
        t.Run(testname, func(t *testing.T) {
            result, _ := bash.ExecCommands(tt.inputStruct)
			if tt.inputStruct != nil {
				if (*result)[0].Command != tt.wantResult.Command || (*result)[0].IsError != tt.wantResult.IsError || (*result)[0].Log != tt.wantResult.Log {
					t.Errorf("Subprocess error: the behavior of the function does not meet expectations\ngot %v, want %v", result, tt.wantResult)
				}
			} else {
				if result != nil {
					t.Errorf("Subprocess error: the behavior of the function does not meet expectations\ngot %v, want %v", result, tt.wantResult)
				}
			}
        })
    }
}