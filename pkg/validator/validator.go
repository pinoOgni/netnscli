package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

const (
// validationTagEndpointPort = "endpoint_port"
)

// TODO fix it, it does not work
// ValidateConfiguration takes a configuration and validates it using the tags defined in the struct
func ValidateConfiguration(configuration interface{}) error {
	// Validate the Config struct
	validate := validator.New(validator.WithRequiredStructEnabled())

	// if err := RegisterValidations(validate); err != nil {
	// return err
	// }
	err := validate.Struct(configuration)
	if err != nil {
		// Handle validation errors
		for _, err := range err.(validator.ValidationErrors) {
			if err.Param() != "" {
				return fmt.Errorf("validation error on field %s. Tag %s, %s", err.Field(), err.Tag(), err.Param())
			} else {
				return fmt.Errorf("validation error on field %s. Tag: %s", err.Field(), err.Tag())
			}
		}
	}
	return nil
}

// RegisterValidations registers custom validations
func RegisterValidations(validate *validator.Validate) error {
	// if err := validate.RegisterValidation(validationTagEndpointPort, validateEndpointPort); err != nil {
	// return fmt.Errorf("cannot register validatation validationTagEndpointPort: %v", err)
	// }
	return nil
}
