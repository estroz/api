package internal

import (
	"reflect"
	"testing"

	"github.com/operator-framework/api/pkg/validation/errors"
)

type validatorFuncTest struct {
	description       string
	wantErr, wantWarn bool
	errors            []errors.Error
	numErrs, numWarns int
}

func (c validatorFuncTest) check(t *testing.T, result errors.ManifestResult) {
	if c.wantErr {
		if !result.HasError() {
			t.Errorf("%s: expected errors %#v, got nil", c.description, c.errors)
		} else {
			numErrors := len(result.Errors)
			if numErrors != c.numErrs {
				t.Errorf("%s: expected %d errors, got: %d",
					c.description, c.numErrs, numErrors)
			}
			errs, _ := splitErrorsWarnings(c.errors)
			if !deepEqualErrors(errs, result.Errors) {
				t.Errorf("%s:\nexpected:\n\t%#v\ngot:\n\t%#v", c.description, errs, result.Errors)
			}
		}
	}
	if c.wantWarn {
		if !result.HasWarn() {
			t.Errorf("%s: expected warnings %#v, got nil", c.description, c.errors)
		} else {
			numWarns := len(result.Warnings)
			if numWarns != c.numWarns {
				t.Errorf("%s: expected %d warnings, got: %d",
					c.description, c.numWarns, numWarns)
			}
			_, warns := splitErrorsWarnings(c.errors)
			if !deepEqualErrors(warns, result.Warnings) {
				t.Errorf("%s:\nexpected:\n\t%#v\ngot:\n\t%#v", c.description, warns, result.Warnings)
			}
		}
	}
	if !c.wantErr && !c.wantWarn && (result.HasError() || result.HasWarn()) {
		t.Errorf("%s: expected no errors or warnings, got:\n%v", c.description, result)
	}
}

func splitErrorsWarnings(all []errors.Error) (errs, warns []errors.Error) {
	for _, a := range all {
		if a.Level == errors.LevelError {
			errs = append(errs, a)
		} else {
			warns = append(warns, a)
		}
	}
	return
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
