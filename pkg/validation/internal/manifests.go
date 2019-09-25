package internal

import (
	"fmt"

	"github.com/operator-framework/api/pkg/validation/errors"
	interfaces "github.com/operator-framework/api/pkg/validation/interfaces"

	"github.com/operator-framework/operator-registry/pkg/registry"
)

type ManifestsValidator struct{}

func (f ManifestsValidator) GetFuncs(objs ...interface{}) (funcs interfaces.ValidatorFuncs) {
	var pkg *registry.PackageManifest
	bundles := []*registry.Bundle{}
	for _, obj := range objs {
		switch v := obj.(type) {
		case *registry.PackageManifest:
			if pkg == nil {
				pkg = v
			}
		case *registry.Bundle:
			bundles = append(bundles, v)
		}
	}
	if pkg != nil && len(bundles) > 0 {
		funcs = append(funcs, func() errors.ManifestResult {
			return validateManifests(pkg, bundles)
		})
	}
	return funcs
}

func validateManifests(pkg *registry.PackageManifest, bundles []*registry.Bundle) (result errors.ManifestResult) {
	csvNames, replacesNames := map[string]struct{}{}, map[string]string{}
	for _, bundle := range bundles {
		bcsv, err := bundle.ClusterServiceVersion()
		if err != nil {
			result.Add(errors.ErrInvalidParse("error getting bundle CSV", err))
			return result
		}
		csv, rerr := bundleCSVToCSV(bcsv)
		if rerr != (errors.Error{}) {
			result.Add(rerr)
			return result
		}
		if csv.Spec.Replaces == "" {
			result.Add(errors.WarnInvalidCSV("`spec.replaces` field not present. If this csv replaces an old version, populate this field with the `metadata.Name` of the old csv", csv.GetName()))
		} else if csv.GetName() == csv.Spec.Replaces {
			result.Add(errors.WarnInvalidCSV("`spec.replaces` field matches its own `metadata.name`. It should contain `metadata.name` of the old CSV to be replaced", csv.GetName()))
		} else {
			csvNames[csv.GetName()] = struct{}{}
			replacesNames[csv.Spec.Replaces] = csv.GetName()
		}
	}
	for replaces, sourceCSV := range replacesNames {
		if _, csvExists := csvNames[replaces]; !csvExists {
			result.Add(errors.ErrInvalidCSV(fmt.Sprintf("%q mentioned in the `spec.replaces` field is not present in manifests", replaces), sourceCSV))
		}
	}
	result.Add(checkDefaultChannelInBundle(pkg, csvNames)...)
	return result
}

func checkDefaultChannelInBundle(pkg *registry.PackageManifest, csvNames map[string]struct{}) (errs []errors.Error) {
	for _, channel := range pkg.Channels {
		if _, csvExists := csvNames[channel.CurrentCSVName]; !csvExists {
			errs = append(errs, errors.ErrInvalidBundle(fmt.Sprintf("currentCSV %q for channel name %q in package %q not found in bundle", channel.CurrentCSVName, channel.Name, pkg.PackageName), channel.CurrentCSVName))
		}
	}
	return errs
}
