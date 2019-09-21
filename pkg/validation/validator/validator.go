package validator

// Validator is an interface for implementing a validator of a single
// Kubernetes object type. Ideally each Validator will check one aspect of
// an object, or perform several steps that have a common theme or goal.
type Validator interface {
	// Validate should run validation logic on an arbitrary object, and return
	// a one ManifestResult for each object that did not pass validation.
	// TODO: use pointers
	Validate() []ManifestResult
	// Name should return a succinct name for this validator.
	Name() string
}

// Validators contains a list of Validators to be executed sequentially.
// TODO: add configurable logger.
type Validators []Validator

// NewValidatorSet creates a Validators containing vs.
func NewValidatorSet(vs ...Validator) Validators {
	vals := Validators{}
	vals.AddValidators(vs...)
	return vals
}

// AddValidators adds each unique Validator in vs to the receiver.
func (vals Validators) AddValidators(vs ...Validator) {
	for _, v := range vs {
		vals = append(vals, v)
	}
}

// ValidateAll runs each Validator in the receiver and returns all results.
func (vals Validators) ValidateAll() (allResults []ManifestResult) {
	for _, val := range vals {
		results := val.Validate()
		allResults = append(allResults, results...)
	}
	return allResults
}
