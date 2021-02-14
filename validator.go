package main

import (
	"fmt"
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

func validateTimeRange(startTime int64, endTime int64,maxTimeRange int64)[]*ErrorResponse{
	var errors []*ErrorResponse
	if (endTime - startTime) < 0 {
		var element ErrorResponse
		element.FailedField = "search_parms"
		element.Tag = "invalid-time-range"
		element.Value = "end time is smaller than start time"
		errors = append(errors, &element)
		return errors
	}
	if (endTime - startTime) > maxTimeRange{
		var element ErrorResponse
		element.FailedField = "search_parms"
		element.Tag = "invalid-time-range"
		element.Value = fmt.Sprintf("time range exceeded maximum %d",maxTimeRange)
		errors = append(errors, &element)
		return errors
	}
	return errors
}

// Running a test with the following curl commands
// curl -X POST -H "Content-Type: application/json" --data "{\"name\":\"john\",\"isactive\":\"True\"}" http://localhost:8080/register/user

// Results in
// [{"FailedField":"User.Email","Tag":"required","Value":""},{"FailedField":"User.Job.Salary","Tag":"required","Value":""},{"FailedField":"User.Job.Type","Tag":"required","Value":""}]⏎