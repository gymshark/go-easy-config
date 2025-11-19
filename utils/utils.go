// Package utils provides utility functions for configuration handling.
package utils

import "reflect"

// IsConfigFullyPopulated checks if all exported fields in a configuration struct are non-zero.
// This is used by InterpolatingChainLoader with ShortCircuit enabled to determine when to stop loading.
func IsConfigFullyPopulated[T any](c *T) bool {
	if c == nil {
		return false
	}
	v := reflect.ValueOf(c)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)
		if structField.PkgPath != "" { // skip unexported fields
			continue
		}
		if IsZero(field) {
			return false
		}
	}
	return true
}

// IsZero determines if a reflect.Value represents a zero value for its type.
// This provides more comprehensive zero-checking than reflect.Value.IsZero(),
// including proper handling of interfaces and arrays.
func IsZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Map, reflect.Slice, reflect.Ptr:
		return v.IsNil()
	case reflect.Interface:
		if v.IsNil() {
			return true
		}
		// If the underlying value is a nil pointer, treat as zero
		underlying := v.Elem()
		if (underlying.Kind() == reflect.Ptr || underlying.Kind() == reflect.Interface) && underlying.IsNil() {
			return true
		}
		// For other types, recursively check if the underlying value is zero
		return IsZero(underlying)
	case reflect.Array:
		// Array is zero if all elements are zero
		for i := 0; i < v.Len(); i++ {
			if !IsZero(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Complex64, reflect.Complex128:
		return v.Complex() == 0
	case reflect.String:
		return v.String() == ""
	}
	return false
}
