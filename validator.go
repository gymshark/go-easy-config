package config

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

func NewValidator() validator.Validate {
	validate := validator.New()

	// Field must be set if all listed fields are set
	_ = validate.RegisterValidation("required_if_all_set", func(fl validator.FieldLevel) bool {
		return !fl.Field().IsZero() || !allFieldsSet(fl.Param(), fl)
	})

	// Field must be set if none of the listed fields are set
	_ = validate.RegisterValidation("required_if_none_set", func(fl validator.FieldLevel) bool {
		return !fl.Field().IsZero() || !noneFieldsSet(fl.Param(), fl)
	})

	// Field must be set if at least one of the listed fields is set
	_ = validate.RegisterValidation("required_if_one_set", func(fl validator.FieldLevel) bool {
		return !fl.Field().IsZero() || !oneFieldSet(fl.Param(), fl)
	})

	// Field must be set if none or only one of the listed fields is set
	_ = validate.RegisterValidation("required_if_none_set_or_one_set", func(fl validator.FieldLevel) bool {
		return !fl.Field().IsZero() || !(noneFieldsSet(fl.Param(), fl) || oneFieldSet(fl.Param(), fl))
	})

	// Field must be set if at most one of the listed fields is set
	_ = validate.RegisterValidation("required_if_at_most_one_set", func(fl validator.FieldLevel) bool {
		return !fl.Field().IsZero() || !atMostOneFieldSet(fl.Param(), fl)
	})

	// Field must be set if at most one of the listed fields is not set
	_ = validate.RegisterValidation("required_if_at_most_one_not_set", func(fl validator.FieldLevel) bool {
		return !fl.Field().IsZero() || !atMostOneFieldNotSet(fl.Param(), fl)
	})

	return *validate
}

func allFieldsSet(param string, fl validator.FieldLevel) bool {
	fields := strings.Fields(param)
	for _, name := range fields {
		f := fl.Parent().FieldByName(name)
		if f.IsZero() {
			return false
		}
	}
	return true
}

func noneFieldsSet(param string, fl validator.FieldLevel) bool {
	fields := strings.Fields(param)
	for _, name := range fields {
		f := fl.Parent().FieldByName(name)
		if !f.IsZero() {
			return false
		}
	}
	return true
}

func oneFieldSet(param string, fl validator.FieldLevel) bool {
	fields := strings.Fields(param)
	count := 0
	for _, name := range fields {
		f := fl.Parent().FieldByName(name)
		if !f.IsZero() {
			count++
		}
	}
	return count == 1
}

func atMostOneFieldSet(param string, fl validator.FieldLevel) bool {
	fields := strings.Fields(param)
	count := 0
	for _, name := range fields {
		f := fl.Parent().FieldByName(name)
		if !f.IsZero() {
			count++
		}
	}
	return count <= 1
}

func atMostOneFieldNotSet(param string, fl validator.FieldLevel) bool {
	fields := strings.Fields(param)
	count := 0
	for _, name := range fields {
		f := fl.Parent().FieldByName(name)
		if f.IsZero() {
			count++
		}
	}
	return count <= 1
}
