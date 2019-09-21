package manifests

import (
	internal "github.com/operator-framework/api/pkg/internal"
	"github.com/operator-framework/api/pkg/validation"
	"github.com/operator-framework/api/pkg/validation/validator"

	"github.com/operator-framework/operator-registry/pkg/registry"
	"github.com/sirupsen/logrus"
)

func GetManifestsDir(dir string) (registry.PackageManifest, []*registry.Bundle, []validator.ManifestResult) {
	// Parse manifest directory and populate a store containing all manifests
	// in dir.
	manifests, err := internal.ManifestsStoreForDir(dir)
	if err != nil {
		logrus.Fatal(err)
		return registry.PackageManifest{}, nil, nil
	}
	pkg := manifests.GetPackageManifest()
	bundles := manifests.GetBundles()
	// Validate manifests collectively in dir.
	val := validation.NewManifestsValidator(pkg, bundles...)
	results := val.Validate()
	return pkg, bundles, results
}
