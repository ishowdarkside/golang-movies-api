package validator

import (
	"regexp"
	"slices"
)

var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

type Validator struct {
	Errors map[string]string
}

func (v *Validator) New() *Validator {

	return &Validator{Errors: make(map[string]string)}

}

func (v *Validator) Valid() bool {

	return len(v.Errors) == 0
}

func (v *Validator) AddError(key string, value string) {

	if _, exists := v.Errors[key]; !exists {

		v.Errors[key] = value
	}

}

func (v *Validator) Check(ok bool, key string, value string) {

	if !ok {

		v.AddError(key, value)

	}
}

func PermittedValue[T comparable](value T, permittedValues ...T) bool {

	return slices.Contains(permittedValues, value)

}

func Matches(val string, regex *regexp.Regexp) bool {

	return regex.MatchString(val)

}

func Unique[T comparable](values []T) bool {

	uniqueValues := make(map[T]bool)

	for _, value := range values {

		uniqueValues[value] = true
	}

	return len(values) == len(uniqueValues)

}
