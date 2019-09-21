package manifests

import (
	internal "github.com/operator-framework/api/pkg/internal"
	"github.com/operator-framework/api/pkg/validation"
	"github.com/operator-framework/api/pkg/validation/validator"

	"github.com/operator-framework/operator-registry/pkg/registry"
	"github.com/sirupsen/logrus"
)

func GetManifestsDir(manifestDirectory string) (registry.PackageManifest, []*registry.Bundle, []validator.ManifestResult) {
	// parse manifest directory
	manifests, err := internal.ManifestsStoreForDir(manifestDirectory)
	if err != nil {
		logrus.Fatal(err)
		return registry.PackageManifest{}, nil, nil
	}
	pkg := manifests.GetPackageManifest()
	bundles := manifests.GetBundles()
	// validate bundle
	val := validation.NewBundleValidator(pkg, bundles...)
	results := val.Validate()
	return pkg, bundles, results
}
