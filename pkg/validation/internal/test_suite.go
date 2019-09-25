package internal

import (
	"reflect"
	"testing"

	"github.com/operator-framework/api/pkg/validation/errors"
)

type validatorFuncTest struct {
	description       string
	wantErr           bool
	errors            []errors.Error
	numErrs, numWarns int
}

func (c validatorFuncTest) check(t *testing.T, result errors.ManifestResult) {
	if c.wantErr {
		if !result.HasError() && !result.HasWarn() {
			t.Errorf("%s: expected error %#v, got nil", c.description, c.errors)
		} else {
			numErrors, numWarns := len(result.Errors), len(result.Warnings)
			if numErrors != c.numErrs || numWarns != c.numWarns {
				t.Errorf("%s: expected %d errors and %d warnings, got: %d and %d",
					c.description, c.numErrs, c.numWarns, numErrors, numWarns)
			}
			allErrors := append(result.Errors, result.Warnings...)
			if !deepEqualErrors(c.errors, allErrors) {
				t.Errorf("%s:\nexpected:\n\t%#v\ngot:\n\t%#v", c.description, c.errors, allErrors)
			}
		}
	} else if result.HasError() || result.HasWarn() {
		t.Errorf("%s: expected no errors, got errors:\n%v", c.description, result)
	}
}

func deepEqualErrors(errs1, errs2 []errors.Error) bool {
	// Do string matching on error types for test purposes.
	for i, err := range errs1 {
		if badErr, ok := err.BadValue.(error); ok && badErr != nil {
			errs1[i].BadValue = badErr.Error()
		}
	}
	for i, err := range errs2 {
		if badErr, ok := err.BadValue.(error); ok && badErr != nil {
			errs2[i].BadValue = badErr.Error()
		}
	}
	return reflect.DeepEqual(errs1, errs2)
}
