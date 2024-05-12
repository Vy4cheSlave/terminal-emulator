package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vy4cheSlave/test-task-postgres/bash"
	mock_bash "github.com/Vy4cheSlave/test-task-postgres/bash/mock"
	mock_database "github.com/Vy4cheSlave/test-task-postgres/database/mock"
	"github.com/Vy4cheSlave/test-task-postgres/models"
	"github.com/golang/mock/gomock"
)

func TestRestApi_CreateNewCommandHandler(t *testing.T) {
	type mockDBBehavior func(*mock_database.MockDBWorker)
	type mockBashBehavior func(*mock_bash.MockBashCommandsWorker)

	testTable := []struct {
		name string
		inputBody string
		mockDBBehavior mockDBBehavior
		mockBashBehavior mockBashBehavior
		expectedStatusCode int
		expectedResponseIsError bool
	} {
		{
			name: `test1 IsError=false`,
			inputBody: `{"bash_strings": ["test1"]}`,
			mockBashBehavior: func(m *mock_bash.MockBashCommandsWorker) {
				m.EXPECT().ExecCommands(&bash.ReqCreateNewCommandBody{
					BashStrings: []string{"test1"},
				}).Return(
					&[]models.CommandsWithoutID{
						{
							Command: "test1",
							IsError: false,
							Log: "test1 is successful",
						},
					},
					nil,
				).AnyTimes()
			},
			mockDBBehavior: func(m *mock_database.MockDBWorker) {
				m.EXPECT().CreateNewCommandsQuery(
					[]models.CommandsWithoutID{
						{
							Command: "test1",
							IsError: false,
							Log: "test1 is successful",
						},
					},
					context.Background(),
				).Return(nil).AnyTimes()
			},
			expectedStatusCode: http.StatusOK,
			expectedResponseIsError: false,
		},
		{
			name: `test2 IsError=true`,
			inputBody: `{"bash_strings": ["test2"]}`,
			mockBashBehavior: func(m *mock_bash.MockBashCommandsWorker) {
				m.EXPECT().ExecCommands(&bash.ReqCreateNewCommandBody{
					BashStrings: []string{"test2"},
				}).Return(
					&[]models.CommandsWithoutID{
						{
							Command: "test2",
							IsError: true,
							Log: "test2 isn't successful",
						},
					},
					nil,
				).AnyTimes()
			},
			mockDBBehavior: func(m *mock_database.MockDBWorker) {
				m.EXPECT().CreateNewCommandsQuery(
					[]models.CommandsWithoutID{
						{
							Command: "test2",
							IsError: true,
							Log: "test2 isn't successful",
						},
					},
					context.Background(),
				).Return(nil).AnyTimes()
			},
			expectedStatusCode: http.StatusOK,
			expectedResponseIsError: true,
		},
		{
			name: `test3 isErrorOnChannel`,
			inputBody: `{"bash_strings": ["test3"]}`,
			mockBashBehavior: func(m *mock_bash.MockBashCommandsWorker) {
				m.EXPECT().ExecCommands(&bash.ReqCreateNewCommandBody{
					BashStrings: []string{"test3"},
				}).Return(
					&[]models.CommandsWithoutID{
						{
							Command: "test3",
							IsError: false,
							Log: "test3 is successful",
						},
					},
					fmt.Errorf("some error in channel"),
				).AnyTimes()
			},
			mockDBBehavior: func(m *mock_database.MockDBWorker) {
				m.EXPECT().CreateNewCommandsQuery(
					[]models.CommandsWithoutID{
						{
							Command: "test3",
							IsError: false,
							Log: "test3 is successful",
						},
					},
					context.Background(),
				).Return(nil).AnyTimes()
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponseIsError: false,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(*testing.T){
			// init dependences
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mDatabase := mock_database.NewMockDBWorker(ctrl)
			testCase.mockDBBehavior(mDatabase)
			mBash := mock_bash.NewMockBashCommandsWorker(ctrl)
			testCase.mockBashBehavior(mBash)

			restApi := RestApi{}

			// test request
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/bash/create-command", bytes.NewBufferString(
				testCase.inputBody,
			))
			handleFunc := restApi.CreateNewCommandHandler(mDatabase, mBash)
			handleFunc(w, r)

			if w.Result().StatusCode != testCase.expectedStatusCode {
				t.Errorf("expected status code %v but got %v", testCase.expectedStatusCode, w.Result().StatusCode)
			}
			defer w.Result().Body.Close()

			wBody, err := io.ReadAll(w.Result().Body)
			if err != nil {
				t.Error(err)
			}
			wBodyStruct := [1]models.CommandsWithoutID{}
			if err := json.Unmarshal(wBody, &wBodyStruct); err != nil {
				t.Error("writer body json unmarshall error")
				return
			}

			if testCase.expectedResponseIsError != wBodyStruct[0].IsError {
				t.Errorf("exptcted response body %v but got %v", testCase.expectedResponseIsError, wBodyStruct[0].IsError)
			}
		}) 
	}

}

func TestRestApi_GettingListCommandsQuery(t *testing.T) {
	type mockDBBehavior func(*mock_database.MockDBWorker)

	testTable := []struct {
		name string
		mockDBBehavior mockDBBehavior
		expectedStatusCode int
	} {
		{
			name: `without db error`,
			mockDBBehavior: func(m *mock_database.MockDBWorker) {
				m.EXPECT().GettingListCommandsQuery(context.Background()).Return(
					&[]models.Commands{
						{},
					},
					nil,
				)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: `with db error`,
			mockDBBehavior: func(m *mock_database.MockDBWorker) {
				m.EXPECT().GettingListCommandsQuery(context.Background()).Return(
					nil,
					fmt.Errorf("some db error"),
				)
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(*testing.T){
			// init dependences
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mDatabase := mock_database.NewMockDBWorker(ctrl)
			testCase.mockDBBehavior(mDatabase)

			restApi := RestApi{}

			// test request
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/bash/get-commands", nil)
			handleFunc := restApi.GettingListCommandsHandler(mDatabase)
			handleFunc(w, r)

			if w.Result().StatusCode != testCase.expectedStatusCode {
				t.Errorf("expected status code %v but got %v", testCase.expectedStatusCode, w.Result().StatusCode)
			}
			defer w.Result().Body.Close()
		}) 
	}

}

func TestRestApi_GettingSingleCommandHandler(t *testing.T) {
	type mockDBBehavior func(*mock_database.MockDBWorker)

	testTable := []struct {
		name string
		pathValue string
		mockDBBehavior mockDBBehavior
		expectedStatusCode int
	} {
		{
			name: `default input`,
			pathValue: "5",
			mockDBBehavior: func(m *mock_database.MockDBWorker) {
				m.EXPECT().GettingSingleCommandQuery(uint(5), context.Background()).Return(
					&models.Commands{
						Id: 5,
						Command: "some command id 5",
						IsError: false,
						Log: "result some command with id 5",
					},
					nil,
				)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: `pathValue <= 0`,
			pathValue: "0",
			mockDBBehavior: func(m *mock_database.MockDBWorker) {},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name: `pathValue is not number`,
			pathValue: "not_number",
			mockDBBehavior: func(m *mock_database.MockDBWorker) {},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name: `db querry error`,
			pathValue: "6",
			mockDBBehavior: func(m *mock_database.MockDBWorker) {
				m.EXPECT().GettingSingleCommandQuery(uint(6), context.Background()).Return(
					nil,
					fmt.Errorf("some db error"),
				)
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(*testing.T){
			// init dependences
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mDatabase := mock_database.NewMockDBWorker(ctrl)
			testCase.mockDBBehavior(mDatabase)

			restApi := RestApi{}

			// test request
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/bash/get-commands/", nil)
			r.SetPathValue("id", testCase.pathValue)
			handleFunc := restApi.GettingSingleCommandHandler(mDatabase)
			handleFunc(w, r)

			if w.Result().StatusCode != testCase.expectedStatusCode {
				t.Errorf("expected status code %v but got %v", testCase.expectedStatusCode, w.Result().StatusCode)
			}
			defer w.Result().Body.Close()
		}) 
	}

}