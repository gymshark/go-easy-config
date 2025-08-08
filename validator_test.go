package config

import (
	"github.com/go-playground/validator/v10"
	"testing"
)

type TestStructRequiredIfAllSet struct {
	FieldA string
	FieldB string
	FieldC string `validate:"required_if_all_set=FieldA FieldB"`
}

type TestStructRequiredIfNoneSet struct {
	FieldA string
	FieldB string
	FieldC string `validate:"required_if_none_set=FieldA FieldB"`
}

type TestStructRequiredIfOneSet struct {
	FieldA string
	FieldB string
	FieldC string `validate:"required_if_one_set=FieldA FieldB"`
}

type TestStructRequiredIfNoneSetOrOneSet struct {
	FieldA string
	FieldB string
	FieldC string `validate:"required_if_none_set_or_one_set=FieldA FieldB"`
}

type TestStructRequiredIfAtMostOneSet struct {
	FieldA string
	FieldB string
	FieldC string `validate:"required_if_at_most_one_set=FieldA FieldB"`
}

type TestStructRequiredIfAtMostOneNotSet struct {
	FieldA string
	FieldB string
	FieldC string `validate:"required_if_at_most_one_not_set=FieldA FieldB"`
}

func getValidator() *validator.Validate {
	v := NewValidator()
	return &v
}

func TestRequiredIfAllSet(t *testing.T) {
	v := getValidator()
	// Should fail: FieldC required if FieldA and FieldB are set
	obj := TestStructRequiredIfAllSet{FieldA: "foo", FieldB: "bar"}
	err := v.Struct(obj)
	if err == nil {
		t.Errorf("Expected error for missing FieldC when FieldA and FieldB are set")
	}
	// Should pass: FieldC present
	obj.FieldC = "baz"
	err = v.Struct(obj)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRequiredIfNoneSet(t *testing.T) {
	v := getValidator()
	// Should fail: FieldC required if FieldA and FieldB are not set
	obj := TestStructRequiredIfNoneSet{}
	err := v.Struct(obj)
	if err == nil {
		t.Errorf("Expected error for missing FieldC when FieldA and FieldB are not set")
	}
	// Should pass: FieldC present
	obj.FieldC = "baz"
	err = v.Struct(obj)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRequiredIfOneSet(t *testing.T) {
	v := getValidator()
	// Should fail: FieldC required if one of FieldA or FieldB is set
	obj := TestStructRequiredIfOneSet{FieldA: "foo"}
	err := v.Struct(obj)
	if err == nil {
		t.Errorf("Expected error for missing FieldC when one of FieldA or FieldB is set")
	}
	// Should pass: FieldC present
	obj.FieldC = "baz"
	err = v.Struct(obj)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRequiredIfNoneSetOrOneSet(t *testing.T) {
	v := getValidator()
	// Should fail: FieldC required if none or one of FieldA/FieldB is set
	obj := TestStructRequiredIfNoneSetOrOneSet{}
	err := v.Struct(obj)
	if err == nil {
		t.Errorf("Expected error for missing FieldC when none of FieldA/FieldB is set")
	}
	obj.FieldA = "foo"
	err = v.Struct(obj)
	if err == nil {
		t.Errorf("Expected error for missing FieldC when only one of FieldA/FieldB is set")
	}
	// Should pass: FieldC present
	obj.FieldC = "baz"
	err = v.Struct(obj)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRequiredIfAtMostOneSet(t *testing.T) {
	v := getValidator()
	// Should fail: FieldC required if at most one of FieldA/FieldB is set
	obj := TestStructRequiredIfAtMostOneSet{FieldA: "foo"}
	err := v.Struct(obj)
	if err == nil {
		t.Errorf("Expected error for missing FieldC when at most one of FieldA/FieldB is set")
	}
	// Should pass: FieldC present
	obj.FieldC = "baz"
	err = v.Struct(obj)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRequiredIfAtMostOneNotSet(t *testing.T) {
	v := getValidator()
	// Should fail: FieldC required if at most one of FieldA/FieldB is not set
	obj := TestStructRequiredIfAtMostOneNotSet{FieldA: "foo", FieldB: "bar"}
	err := v.Struct(obj)
	if err == nil {
		t.Errorf("Expected error for missing FieldC when at most one of FieldA/FieldB is not set")
	}
	// Should pass: FieldC present
	obj.FieldC = "baz"
	err = v.Struct(obj)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
