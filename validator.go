package main

import (
	"github.com/go-playground/validator/v10"
)

type ErrorResponse struct {
	FailedField string
	Tag         string
	Value       string
}

func ValidateStruct(requestConversations YFRequestConversations) []*ErrorResponse {
	var errors []*ErrorResponse
	validate := validator.New()
	err := validate.Struct(requestConversations)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element ErrorResponse
			element.FailedField = err.StructNamespace()
			element.Tag = err.Tag()
			element.Value = err.Param()
			errors = append(errors, &element)
		}
	}
	return errors
}

// Running a test with the following curl commands
// curl -X POST -H "Content-Type: application/json" --data "{\"name\":\"john\",\"isactive\":\"True\"}" http://localhost:8080/register/user

// Results in
// [{"FailedField":"User.Email","Tag":"required","Value":""},{"FailedField":"User.Job.Salary","Tag":"required","Value":""},{"FailedField":"User.Job.Type","Tag":"required","Value":""}]⏎