// Package validation provides default Validator's that can be run with a list
// of arbitrary objects. The defaults exposed here consist of all Validator's
// implemented by this validation library.
//
// Each default Validator runs an independent set of validation functions on
// a set of objects. To run all implemented Validator's, use WithDefaultValidators().
// The Validator will not be run on objects not of the appropriate type.

package validation

import (
	interfaces "github.com/operator-framework/api/pkg/validation/interfaces"
	"github.com/operator-framework/api/pkg/validation/internal"
)

// WithDefaultPackageManifestValidators returns a package manifest Validator
// that can be run directly with Apply(objs...). Optionally, any additional
// Validator's can be added to the returned Validators set.
func WithDefaultPackageManifestValidators(vals ...interfaces.Validator) interfaces.Validators {
	return append(vals, internal.ValidatePackageManifest)
}

// WithDefaultClusterServiceVersionValidators returns a ClusterServiceVersion
// Validator that can be run directly with Apply(objs...). Optionally, any
// additional Validator's can be added to the returned Validators set.
func WithDefaultClusterServiceVersionValidators(vals ...interfaces.Validator) interfaces.Validators {
	return append(vals, internal.Validatev1alpha1CSV)
}

// WithDefaultCustomResourceDefinitionValidators returns a
// CustomResourceDefinition Validator that can be run directly with
// Apply(objs...). Optionally, any additional Validator's can be added to the
// returned Validators set.
func WithDefaultCustomResourceDefinitionValidators(vals ...interfaces.Validator) interfaces.Validators {
	return append(vals, internal.Validatev1beta1CRD)
}

// WithDefaultBundleValidators returns a bundle Validator that can be run
// directly with Apply(objs...). Optionally, any additional Validator's can
// be added to the returned Validators set.
func WithDefaultBundleValidators(vals ...interfaces.Validator) interfaces.Validators {
	return append(vals, internal.ValidateBundle)
}

// WithDefaultManifestsValidators returns a manifests Validator that can be run
// directly with Apply(objs...). Optionally, any additional Validator's can be
// added to the returned Validators set.
func WithDefaultManifestsValidators(vals ...interfaces.Validator) interfaces.Validators {
	return append(vals, internal.ValidateManifests)
}

// WithDefaultValidators returns all default Validator's, which can be run
// directly with Apply(objs...). Optionally, any additional Validator's can
// be added to the returned Validators set.
func WithDefaultValidators(vals ...interfaces.Validator) interfaces.Validators {
	return append(vals,
		internal.ValidatePackageManifest,
		internal.Validatev1alpha1CSV,
		internal.Validatev1beta1CRD,
		internal.ValidateBundle,
		internal.ValidateManifests,
	)
}
