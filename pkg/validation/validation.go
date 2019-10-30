package validation

import (
	interfaces "github.com/operator-framework/api/pkg/validation/interfaces"
	"github.com/operator-framework/api/pkg/validation/internal"
)

func DefaultPackageManifestValidators(vals ...interfaces.Validator) interfaces.Validators {
	return append(vals, internal.PackageManifestValidator{})
}

func DefaultClusterServiceVersionValidators(vals ...interfaces.Validator) interfaces.Validators {
	return append(vals, internal.CSVValidator{})
}

func DefaultCustomResourceDefinitionValidators(vals ...interfaces.Validator) interfaces.Validators {
	return append(vals, internal.CRDValidator{})
}

func DefaultBundleValidators(vals ...interfaces.Validator) interfaces.Validators {
	return append(vals, internal.BundleValidator{})
}

func DefaultManifestsValidators(vals ...interfaces.Validator) interfaces.Validators {
	return append(vals, internal.ManifestsValidator{})
}

func DefaultValidators(vals ...interfaces.Validator) interfaces.Validators {
	return append(vals,
		internal.PackageManifestValidator{},
		internal.CSVValidator{},
		internal.CRDValidator{},
		internal.BundleValidator{},
		internal.ManifestsValidator{},
	)
}
