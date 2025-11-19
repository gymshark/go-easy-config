// Package aws provides helper functions for handling mixed tag scenarios
// where structs contain both secret tags and other types of tags.
package aws

import (
	"fmt"
	"reflect"
)

// hasSecretTags checks if the struct has any fields with secret tags
func hasSecretTags(c interface{}) bool {
	v := reflect.ValueOf(c)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // skip unexported fields
			continue
		}
		if field.Tag.Get("secret") != "" {
			return true
		}
	}
	return false
}

// createSecretOnlyStruct creates a new struct containing only fields with secret tags
func createSecretOnlyStruct(c interface{}) (interface{}, map[string]int, error) {
	v := reflect.ValueOf(c)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("expected struct, got %T", c)
	}

	t := v.Type()
	var fields []reflect.StructField
	fieldMap := make(map[string]int) // maps temp struct field index to original struct field index

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // skip unexported fields
			continue
		}
		if field.Tag.Get("secret") != "" {
			fieldMap[field.Name] = i
			fields = append(fields, field)
		}
	}

	if len(fields) == 0 {
		return nil, nil, nil // No secret fields
	}

	// Create new struct type with only secret fields
	newType := reflect.StructOf(fields)
	newStruct := reflect.New(newType).Interface()

	return newStruct, fieldMap, nil
}

// copySecretValues copies values from the temporary struct back to the original struct
func copySecretValues(original, temp interface{}, fieldMap map[string]int) error {
	origVal := reflect.ValueOf(original)
	if origVal.Kind() == reflect.Ptr {
		origVal = origVal.Elem()
	}

	tempVal := reflect.ValueOf(temp)
	if tempVal.Kind() == reflect.Ptr {
		tempVal = tempVal.Elem()
	}

	tempType := tempVal.Type()
	for i := 0; i < tempVal.NumField(); i++ {
		tempField := tempVal.Field(i)
		fieldName := tempType.Field(i).Name

		origIndex, exists := fieldMap[fieldName]
		if !exists {
			continue
		}

		origField := origVal.Field(origIndex)
		if origField.CanSet() && !tempField.IsZero() {
			origField.Set(tempField)
		}
	}

	return nil
}
