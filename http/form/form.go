package form

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/socle-lab/pkg/util"
)

type FieldTagValidation struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
}

func ExtractValidationErrors(err error) []FieldTagValidation {
	var out []FieldTagValidation

	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			out = append(out, FieldTagValidation{
				Field: e.Field(),
				Tag:   e.Tag(),
			})
		}
	}

	return out
}

func BindForm[T any](r *http.Request, dst *T) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	// Use reflection to fill the struct
	v := reflect.ValueOf(dst).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		formKey := field.Tag.Get("form")

		if formKey == "" {
			formKey = util.ToSnakeCase(field.Name)
		}

		value := r.FormValue(formKey)
		if value == "" {
			continue
		}

		f := v.Field(i)
		if !f.CanSet() {
			continue
		}

		switch f.Kind() {
		case reflect.String:
			f.SetString(value)

		case reflect.Int, reflect.Int64:
			n, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid int for %s: %v", formKey, err)
			}
			f.SetInt(n)

		case reflect.Float64:
			n, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("invalid float for %s: %v", formKey, err)
			}
			f.SetFloat(n)

		case reflect.Bool:
			b := value == "1" || strings.ToLower(value) == "true" || value == "on"
			f.SetBool(b)
		}
	}

	return nil
}
