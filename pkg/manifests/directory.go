package manifests

import (
	"fmt"

	internal "github.com/operator-framework/api/pkg/internal"
	"github.com/operator-framework/api/pkg/validation"
	"github.com/operator-framework/api/pkg/validation/errors"

	"github.com/operator-framework/operator-registry/pkg/registry"
)

func GetManifestsDir(dir string) (*registry.PackageManifest, []*registry.Bundle, []errors.ManifestResult) {
	manifests, err := internal.ManifestsStoreForDir(dir)
	if err != nil {
		result := errors.ManifestResult{}
		result.Add(errors.ErrInvalidParse(fmt.Sprintf("parse manifests from %q", dir), err))
		return nil, nil, []errors.ManifestResult{result}
	}
	pkg := manifests.GetPackageManifest()
	bundles := manifests.GetBundles()
	objs := []interface{}{}
	for _, obj := range bundles {
		objs = append(objs, obj)
	}
	results := validation.DefaultValidators().Apply(objs...)
	return pkg, bundles, results
}
