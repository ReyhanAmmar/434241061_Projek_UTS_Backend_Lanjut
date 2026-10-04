package helper

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
)

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New()

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})

	if err := v.RegisterValidation(
		"maxbytes",
		func(fl validator.FieldLevel) bool {
			limit, err := strconv.Atoi(fl.Param())
			if err != nil || fl.Field().Kind() != reflect.String {
				return false
			}
			return len(fl.Field().String()) <= limit
		},
	); err != nil {
		panic(err)
	}

	if err := v.RegisterValidation(
		"strongpassword",
		func(fl validator.FieldLevel) bool {
			return passwordStrength(fl.Field().String()) == ""
		},
	); err != nil {
		panic(err)
	}

	return v
}

func ValidateStruct(value any) map[string]string {
	err := validate.Struct(value)
	if err == nil {
		return nil
	}

	var invalid *validator.InvalidValidationError
	if errors.As(err, &invalid) {
		return map[string]string{
			"_": "objek yang divalidasi tidak sah",
		}
	}

	var fieldErrors validator.ValidationErrors
	if !errors.As(err, &fieldErrors) {
		return map[string]string{
			"_": "validasi gagal",
		}
	}

	result := make(map[string]string, len(fieldErrors))

	for _, fieldErr := range fieldErrors {
		if _, exists := result[fieldErr.Field()]; !exists {
			result[fieldErr.Field()] = messageFor(fieldErr)
		}
	}

	return result
}

func messageFor(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "wajib diisi"

	case "email":
		return "format email tidak valid"

	case "min":
		if fe.Kind() == reflect.String {
			return "minimal " + fe.Param() + " karakter"
		}
		return "nilai minimal " + fe.Param()

	case "max":
		if fe.Kind() == reflect.String {
			return "maksimal " + fe.Param() + " karakter"
		}
		return "nilai maksimal " + fe.Param()

	case "maxbytes":
		return "maksimal " + fe.Param() + " byte"

	case "len":
		return "harus terdiri dari " + fe.Param() + " karakter"

	case "hexadecimal":
		return "harus berupa karakter heksadesimal"

	case "strongpassword":
		if value, ok := fe.Value().(string); ok {
			if message := passwordStrength(value); message != "" {
				return message
			}
		}
		return "password tidak memenuhi syarat"

	default:
		return "tidak memenuhi aturan " + fe.Tag()
	}
}

func passwordStrength(password string) string {
	if utf8.RuneCountInString(password) < 8 {
		return "minimal 8 karakter"
	}

	var hasLetter, hasDigit bool

	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}

	if !hasLetter || !hasDigit {
		return "harus memuat huruf dan angka"
	}

	weak := map[string]bool{
		"password1":   true,
		"12345678":    true,
		"qwerty123":   true,
		"admin123":    true,
		"password123": true,
	}

	if weak[strings.ToLower(password)] {
		return "password terlalu umum"
	}

	return ""
}