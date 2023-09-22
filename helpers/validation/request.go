package validation

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type InvalidSchema struct {
	Field   string `json:"field"`
	Message any    `json:"message"`
}

func RequestBody(errorFields validator.ValidationErrors, payload interface{}) []InvalidSchema {
	var invalid []InvalidSchema

	for _, errorField := range errorFields {
		invalid = append(invalid, InvalidSchema{
			Field:   errorField.Field(),
			Message: fmt.Sprintf("invalid '%s' with value '%v'", errorField.Field(), errorField.Value()),
		})
	}

	return invalid
}
