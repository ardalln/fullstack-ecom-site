package handler

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator adapts go-playground/validator to Echo's Validator interface.
type Validator struct {
	validate *validator.Validate
}

func NewValidator() *Validator {
	v := validator.New()
	// Report errors with the JSON field name instead of the Go field name.
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return &Validator{validate: v}
}

func (v *Validator) Validate(i any) error {
	return v.validate.Struct(i)
}

func validationDetails(errs validator.ValidationErrors) map[string]string {
	details := make(map[string]string, len(errs))
	for _, fe := range errs {
		field := fe.Namespace()
		if i := strings.Index(field, "."); i >= 0 {
			field = field[i+1:] // drop the struct name
		}
		details[field] = validationMessage(fe)
	}
	return details
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		switch fe.Kind() {
		case reflect.String:
			return fmt.Sprintf("must be at least %s characters long", fe.Param())
		case reflect.Slice:
			return fmt.Sprintf("must contain at least %s item(s)", fe.Param())
		}
		return fmt.Sprintf("must be at least %s", fe.Param())
	case "max":
		switch fe.Kind() {
		case reflect.String:
			return fmt.Sprintf("must be at most %s characters long", fe.Param())
		case reflect.Slice:
			return fmt.Sprintf("must contain at most %s item(s)", fe.Param())
		}
		return fmt.Sprintf("must be at most %s", fe.Param())
	case "gt":
		return fmt.Sprintf("must be greater than %s", fe.Param())
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fe.Param())
	default:
		return fmt.Sprintf("failed validation %q", fe.Tag())
	}
}
