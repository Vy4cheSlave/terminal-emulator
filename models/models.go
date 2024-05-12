package models

type Commands struct {
	Id uint `json:"id"`
	Command string `json:"command"`
	IsError bool `json:"is_error"`
	Log string `json:"log"`
}

type CommandsWithoutID struct {
	Command string `json:"command"`
	IsError bool `json:"is_error"`
	Log string `json:"log"`
}