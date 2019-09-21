package validation

import (
	"fmt"

	"github.com/operator-framework/api/pkg/validation/validator"

	"github.com/operator-framework/operator-registry/pkg/registry"
	"github.com/sirupsen/logrus"
)

type manifestsValidator struct {
	pkg     registry.PackageManifest
	bundles []*registry.Bundle
}

func NewManifestsValidator(pkg registry.PackageManifest, bundles ...*registry.Bundle) validator.Validator {
	return &manifestsValidator{
		pkg:     pkg,
		bundles: bundles,
	}
}

func (v *manifestsValidator) Validate() []validator.ManifestResult {
	vals := validator.NewValidators(NewBundleValidator(v.bundles...), NewPackageManifestValidator(v.pkg))
	results := vals.ValidateAll()

	result := v.validateManifestsCollective()
	result.Name = v.pkg.PackageName
	results = append(results, result)
	return results
}

func (v manifestsValidator) Name() string {
	return "Manifests Validator"
}

func (v manifestsValidator) validateManifestsCollective() (results validator.ManifestResult) {
	csvNames, replacesNames := map[string]struct{}{}, map[string]string{}
	for _, bundle := range v.bundles {
		bcsv, err := bundle.ClusterServiceVersion()
		if err != nil {
			logrus.Fatal(err)
			return validator.ManifestResult{}
		}
		csv := mustBundleCSVToCSV(bcsv)
		if csv.Spec.Replaces == "" {
			results.Add(validator.WarnInvalidCSV("`spec.replaces` field not present. If this csv replaces an old version, populate this field with the `metadata.Name` of the old csv", csv.GetName()))
		} else if csv.GetName() == csv.Spec.Replaces {
			results.Add(validator.WarnInvalidCSV("`spec.replaces` field matches its own `metadata.name`. It should contain `metadata.name` of the old CSV to be replaced", csv.GetName()))
		} else {
			csvNames[csv.GetName()] = struct{}{}
			replacesNames[csv.Spec.Replaces] = csv.GetName()
		}
	}
	for replaces, sourceCSV := range replacesNames {
		if _, csvExists := csvNames[replaces]; !csvExists {
			results.Add(validator.ErrInvalidCSV(fmt.Sprintf("`%s` mentioned in the `spec.replaces` field is not present in manifests", replaces), sourceCSV))
		}
	}
	results.Add(checkDefaultChannelInBundle(v.pkg, csvNames)...)
	return results
}

func checkDefaultChannelInBundle(pkg registry.PackageManifest, csvNames map[string]struct{}) (errs []validator.Error) {
	for _, channel := range pkg.Channels {
		if _, csvExists := csvNames[channel.CurrentCSVName]; !csvExists {
			errs = append(errs, validator.ErrInvalidBundle(fmt.Sprintf("currentCSV `%s` for channel name `%s` in package `%s` not found in bundle", channel.CurrentCSVName, channel.Name, pkg.PackageName), channel.CurrentCSVName))
		}
	}
	return errs
}
