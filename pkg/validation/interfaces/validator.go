package validator

import (
	"sync"

	"github.com/operator-framework/api/pkg/validation/errors"
)

// ValidatorFunc returns a ManifestResult containing errors and warnings from
// validating some object. These are typically closures or methods.
type ValidatorFunc func() errors.ManifestResult

// ValidationFuncs is a set of validation functions.
type ValidatorFuncs []ValidatorFunc

// Validator is an interface for validating arbitrary objects. A Validator
// returns a set of functions that each validate some underlying object.
// This allows the implementer to easily break down each validation step
// into discrete units if they wish, and perform either lazy or active
// validation.
type Validator interface {
	// GetFuncs takes a list of arbitrary objects and returns a set of functions
	// that each validate some object from the list.
	GetFuncs(...interface{}) ValidatorFuncs
}

// Validators is a set of Validator's that can be run via Apply.
type Validators []Validator

// Apply collects validator functions from each Validator in vals by collecting
// the appropriate functions for each obj in objs, then invokes them,
// collecting the results. Use Apply in code that:
// - Uses more than one Validator in one call.
// - Want active validation.
func (vals Validators) Apply(objs ...interface{}) (results []errors.ManifestResult) {
	if len(vals) == 0 {
		return nil
	}
	if len(vals) == 1 {
		return vals[0].GetFuncs(objs...).runP()
	}
	queue := make(chan []errors.ManifestResult)
	wg := &sync.WaitGroup{}
	wg.Add(len(vals))
	for _, val := range vals {
		go func(v Validator) {
			funcResults := []errors.ManifestResult{}
			for _, result := range v.GetFuncs(objs...).runP() {
				if result.HasError() || result.HasWarn() {
					funcResults = append(funcResults, result)
				}
			}
			queue <- funcResults
		}(val)
	}
	go func() {
		for funcResults := range queue {
			results = append(results, funcResults...)
			wg.Done()
		}
	}()
	wg.Wait()
	return results
}

// runP runs all funcs in parallel.
func (funcs ValidatorFuncs) runP() (results []errors.ManifestResult) {
	if len(funcs) == 0 {
		return nil
	}
	if len(funcs) == 1 {
		results = append(results, funcs[0]())
	}
	queue := make(chan errors.ManifestResult, 1)
	wg := &sync.WaitGroup{}
	wg.Add(len(funcs))
	for _, validate := range funcs {
		go func(f ValidatorFunc) {
			queue <- f()
		}(validate)
	}
	go func() {
		for result := range queue {
			if result.HasError() || result.HasWarn() {
				results = append(results, result)
			}
			wg.Done()
		}
	}()
	wg.Wait()
	return results
}
