package validator

import "github.com/go-playground/validator/v10"

type Validator struct {
	*validator.Validate
}

func NewValidator() (Validator, error) {
	return Validator{
		validator.New(),
	}, nil
}
