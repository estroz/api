package validation

import (
	interfaces "github.com/operator-framework/api/pkg/validation/interfaces"
	valinternal "github.com/operator-framework/api/pkg/validation/internal"
)

func DefaultPackageManifestValidators(vals ...interfaces.Validator) interfaces.Validators {
	return append(vals, valinternal.PackageManifestValidator{})
}

func DefaultClusterServiceVersionValidators(vals ...interfaces.Validator) interfaces.Validators {
	return append(vals, valinternal.CSVValidator{})
}

func DefaultCustomResourceDefinitionValidators(vals ...interfaces.Validator) interfaces.Validators {
	return append(vals, valinternal.CRDValidator{})
}

func DefaultBundleValidators(vals ...interfaces.Validator) interfaces.Validators {
	return append(vals, valinternal.BundleValidator{})
}

func DefaultManifestsValidators(vals ...interfaces.Validator) interfaces.Validators {
	return append(vals, valinternal.ManifestsValidator{})
}

func DefaultValidators(vals ...interfaces.Validator) interfaces.Validators {
	return append(vals,
		valinternal.PackageManifestValidator{},
		valinternal.CSVValidator{},
		valinternal.CRDValidator{},
		valinternal.BundleValidator{},
		valinternal.ManifestsValidator{},
	)
}
