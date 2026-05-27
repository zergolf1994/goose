package goose

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ── Validation Engine ────────────────────────────────────────
//
// Supported goose validation tags:
//
//   goose:"required"              → field must not be zero-value
//   goose:"enum:active|inactive"  → string must be one of the values
//   goose:"min:0"                 → minimum value (int/float) or min length (string)
//   goose:"max:100"               → maximum value (int/float) or max length (string)
//   goose:"minlength:2"           → minimum string length
//   goose:"maxlength:100"         → maximum string length
//   goose:"match:^[a-z]+$"        → string must match regex pattern

// ValidationError represents a single field validation error.
type ValidationError struct {
	Field   string // bson field name
	Message string // human-readable error message
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors struct {
	Errors []ValidationError
}

func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return ""
	}
	var msgs []string
	for _, err := range e.Errors {
		msgs = append(msgs, err.Error())
	}
	return "validation failed: " + strings.Join(msgs, "; ")
}

// HasErrors returns true if there are any validation errors.
func (e *ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// Validate checks a document against its goose struct tag validators.
// Returns nil if valid, or *ValidationErrors with all failures.
func Validate(doc interface{}) error {
	errs := &ValidationErrors{}
	validateStruct(doc, errs)
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// ValidateField validates a single field by its bson name.
func ValidateField(doc interface{}, bsonFieldName string) error {
	errs := &ValidationErrors{}
	validateSingleField(doc, bsonFieldName, errs)
	if errs.HasErrors() {
		return errs
	}
	return nil
}

func validateStruct(doc interface{}, errs *ValidationErrors) {
	v := reflect.ValueOf(doc)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)

		// Recurse into embedded structs
		if structField.Anonymous && field.Kind() == reflect.Struct {
			if field.CanAddr() {
				validateStruct(field.Addr().Interface(), errs)
			}
			continue
		}

		gooseTag := structField.Tag.Get("goose")
		if gooseTag == "" {
			continue
		}

		bsonTag := structField.Tag.Get("bson")
		bsonName := strings.Split(bsonTag, ",")[0]
		if bsonName == "" {
			bsonName = structField.Name
		}

		validateFieldValue(field, bsonName, gooseTag, errs)
	}
}

func validateSingleField(doc interface{}, bsonFieldName string, errs *ValidationErrors) {
	v := reflect.ValueOf(doc)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)

		if structField.Anonymous && field.Kind() == reflect.Struct {
			if field.CanAddr() {
				validateSingleField(field.Addr().Interface(), bsonFieldName, errs)
			}
			continue
		}

		bsonTag := structField.Tag.Get("bson")
		bsonName := strings.Split(bsonTag, ",")[0]
		if bsonName == "" {
			bsonName = structField.Name
		}
		if bsonName != bsonFieldName {
			continue
		}

		gooseTag := structField.Tag.Get("goose")
		if gooseTag != "" {
			validateFieldValue(field, bsonName, gooseTag, errs)
		}
	}
}

func validateFieldValue(field reflect.Value, bsonName string, gooseTag string, errs *ValidationErrors) {
	parts := strings.Split(gooseTag, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		switch {
		case part == "required":
			if isZeroValue(field) {
				errs.Errors = append(errs.Errors, ValidationError{
					Field:   bsonName,
					Message: "is required",
				})
			}

		case strings.HasPrefix(part, "enum:"):
			if field.Kind() == reflect.String {
				val := field.String()
				allowed := strings.Split(part[len("enum:"):], "|")
				if val != "" && !contains(allowed, val) {
					errs.Errors = append(errs.Errors, ValidationError{
						Field:   bsonName,
						Message: fmt.Sprintf("must be one of [%s], got '%s'", strings.Join(allowed, ", "), val),
					})
				}
			}

		case strings.HasPrefix(part, "min:"):
			minStr := part[len("min:"):]
			validateMin(field, bsonName, minStr, errs)

		case strings.HasPrefix(part, "max:"):
			maxStr := part[len("max:"):]
			validateMax(field, bsonName, maxStr, errs)

		case strings.HasPrefix(part, "minlength:"):
			if field.Kind() == reflect.String {
				val := field.String()
				minLen, err := strconv.Atoi(part[len("minlength:"):])
				if err == nil && len(val) > 0 && len(val) < minLen {
					errs.Errors = append(errs.Errors, ValidationError{
						Field:   bsonName,
						Message: fmt.Sprintf("must be at least %d characters, got %d", minLen, len(val)),
					})
				}
			}

		case strings.HasPrefix(part, "maxlength:"):
			if field.Kind() == reflect.String {
				val := field.String()
				maxLen, err := strconv.Atoi(part[len("maxlength:"):])
				if err == nil && len(val) > maxLen {
					errs.Errors = append(errs.Errors, ValidationError{
						Field:   bsonName,
						Message: fmt.Sprintf("must be at most %d characters, got %d", maxLen, len(val)),
					})
				}
			}

		case strings.HasPrefix(part, "match:"):
			if field.Kind() == reflect.String {
				val := field.String()
				pattern := part[len("match:"):]
				if val != "" {
					re, err := regexp.Compile(pattern)
					if err == nil && !re.MatchString(val) {
						errs.Errors = append(errs.Errors, ValidationError{
							Field:   bsonName,
							Message: fmt.Sprintf("must match pattern '%s'", pattern),
						})
					}
				}
			}
		}
	}
}

func validateMin(field reflect.Value, bsonName string, minStr string, errs *ValidationErrors) {
	switch field.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		min, err := strconv.ParseInt(minStr, 10, 64)
		if err == nil && field.Int() < min {
			errs.Errors = append(errs.Errors, ValidationError{
				Field:   bsonName,
				Message: fmt.Sprintf("must be at least %d, got %d", min, field.Int()),
			})
		}
	case reflect.Float64, reflect.Float32:
		min, err := strconv.ParseFloat(minStr, 64)
		if err == nil && field.Float() < min {
			errs.Errors = append(errs.Errors, ValidationError{
				Field:   bsonName,
				Message: fmt.Sprintf("must be at least %g, got %g", min, field.Float()),
			})
		}
	case reflect.String:
		// min on string = minlength
		min, err := strconv.Atoi(minStr)
		if err == nil && len(field.String()) > 0 && len(field.String()) < min {
			errs.Errors = append(errs.Errors, ValidationError{
				Field:   bsonName,
				Message: fmt.Sprintf("must be at least %d characters, got %d", min, len(field.String())),
			})
		}
	}
}

func validateMax(field reflect.Value, bsonName string, maxStr string, errs *ValidationErrors) {
	switch field.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		max, err := strconv.ParseInt(maxStr, 10, 64)
		if err == nil && field.Int() > max {
			errs.Errors = append(errs.Errors, ValidationError{
				Field:   bsonName,
				Message: fmt.Sprintf("must be at most %d, got %d", max, field.Int()),
			})
		}
	case reflect.Float64, reflect.Float32:
		max, err := strconv.ParseFloat(maxStr, 64)
		if err == nil && field.Float() > max {
			errs.Errors = append(errs.Errors, ValidationError{
				Field:   bsonName,
				Message: fmt.Sprintf("must be at most %g, got %g", max, field.Float()),
			})
		}
	case reflect.String:
		// max on string = maxlength
		max, err := strconv.Atoi(maxStr)
		if err == nil && len(field.String()) > max {
			errs.Errors = append(errs.Errors, ValidationError{
				Field:   bsonName,
				Message: fmt.Sprintf("must be at most %d characters, got %d", max, len(field.String())),
			})
		}
	}
}

func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Struct:
		if v.Type() == reflect.TypeOf(time.Time{}) {
			t := v.Interface().(time.Time)
			return t.IsZero()
		}
		return false
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map:
		return v.IsNil()
	default:
		return false
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
