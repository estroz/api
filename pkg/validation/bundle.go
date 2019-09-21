package validation

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/operator-framework/api/pkg/validation/validator"

	olmapiv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/operator-framework/operator-registry/pkg/registry"
	"github.com/sirupsen/logrus"
)

type bundleValidator struct {
	pkg     registry.PackageManifest
	bundles map[string]*registry.Bundle
}

func NewBundleValidator(pkg registry.PackageManifest, bundles ...*registry.Bundle) validator.Validator {
	// TODO: define addObjects for bundle.go
	return &bundleValidator{}
}

func (v *bundleValidator) Validate() (results []validator.ManifestResult) {
	// validate individual bundle files
	for _, bundle := range v.bundles {
		bcsv, err := bundle.ClusterServiceVersion()
		if err != nil {
			logrus.Fatal(err)
			return nil
		}
		csv := mustBundleCSVToCSV(bcsv)
		validators := []validator.Validator{NewCSVValidator(csv)}
		crds, err := bundle.CustomResourceDefinitions()
		if err != nil {
			logrus.Fatal(err)
			return nil
		}
		validators = append(validators, NewCRDValidator(crds...))
		for _, val := range validators {
			results = append(results, val.Validate()...)
		}
	}
	val := NewPackageManifestValidator(v.pkg)
	results = append(results, val.Validate()...)

	result := v.validateBundle()
	result.Name = v.pkg.PackageName
	results = append(results, result)
	return results
}

func (v bundleValidator) Name() string {
	return "Bundle Validator"
}

func mustBundleCSVToCSV(bcsv *registry.ClusterServiceVersion) *olmapiv1alpha1.ClusterServiceVersion {
	spec := olmapiv1alpha1.ClusterServiceVersionSpec{}
	if err := json.Unmarshal(bcsv.Spec, &spec); err != nil {
		log.Fatalf("Error converting bundle CSV %q type: %v", bcsv.GetName(), err)
	}
	return &olmapiv1alpha1.ClusterServiceVersion{
		TypeMeta:   bcsv.TypeMeta,
		ObjectMeta: bcsv.ObjectMeta,
		Spec:       spec,
	}
}

func (v bundleValidator) validateBundle() validator.ManifestResult {
	manifestResult := validator.ManifestResult{}
	csvReplacesMap := make(map[string]string)
	var csvsInBundle []string
	for _, bundle := range v.bundles {
		bcsv, err := bundle.ClusterServiceVersion()
		if err != nil {
			logrus.Fatal(err)
			return validator.ManifestResult{}
		}
		csv := mustBundleCSVToCSV(bcsv)
		csvsInBundle = append(csvsInBundle, csv.GetName())
		csvReplacesMap[csv.GetName()] = csv.Spec.Replaces
		if csv.GetName() == csv.Spec.Replaces {
			manifestResult.Add(validator.WarnInvalidCSV("`spec.replaces` field matches its own `metadata.name`. It should contain `metadata.name` of the old CSV to be replaced", csv.GetName()))
		}
		manifestResult = validateOwnedCRDs(bundle, csv, manifestResult)
	}
	manifestResult = checkReplacesForCSVs(csvReplacesMap, csvsInBundle, manifestResult)
	manifestResult = checkDefaultChannelInBundle(v.pkg, csvsInBundle, manifestResult)
	return manifestResult
}

func checkDefaultChannelInBundle(pkg registry.PackageManifest, csvsInBundle []string, manifestResult validator.ManifestResult) validator.ManifestResult {
	for _, channel := range pkg.Channels {
		if !isStringPresent(csvsInBundle, channel.CurrentCSVName) {
			manifestResult.Add(validator.ErrInvalidBundle(fmt.Sprintf("currentCSV `%s` for channel name `%s` in package `%s` not found in bundle", channel.CurrentCSVName, channel.Name, pkg.PackageName), channel.CurrentCSVName))
		}
	}
	return manifestResult
}

func validateOwnedCRDs(bundle *registry.Bundle, csv *olmapiv1alpha1.ClusterServiceVersion, manifestResult validator.ManifestResult) validator.ManifestResult {
	ownedCrdNames := getOwnedCustomResourceDefintionNames(csv)
	bundleCrdNames, err := getBundleCRDNames(bundle)
	if err != (validator.Error{}) {
		manifestResult.Add(err)
		return manifestResult
	}

	// validating names
	for _, ownedCrd := range ownedCrdNames {
		if !bundleCrdNames[ownedCrd] {
			manifestResult.Add(validator.ErrInvalidBundle(fmt.Sprintf("owned crd (%s) not found in bundle %s", ownedCrd, bundle.Name), ownedCrd))
		} else {
			delete(bundleCrdNames, ownedCrd)
		}
	}
	// CRDs not defined in the CSV present in the bundle
	if len(bundleCrdNames) != 0 {
		for crd, _ := range bundleCrdNames {
			manifestResult.Add(validator.WarnInvalidBundle(fmt.Sprintf("`%s` crd present in bundle `%s` not defined in csv", crd, bundle.Name), crd))
		}
	}
	return manifestResult
}

func getOwnedCustomResourceDefintionNames(csv *olmapiv1alpha1.ClusterServiceVersion) []string {
	var names []string
	for _, ownedCrd := range csv.Spec.CustomResourceDefinitions.Owned {
		names = append(names, ownedCrd.Name)
	}
	return names
}

func getBundleCRDNames(bundle *registry.Bundle) (map[string]bool, validator.Error) {
	crds, err := bundle.CustomResourceDefinitions()
	if err != nil {
		logrus.Fatal(err)
		return nil, validator.Error{}
	}
	bundleCrdNames := make(map[string]bool)
	for _, crd := range crds {
		bundleCrdNames[crd.GetName()] = true
	}
	return bundleCrdNames, validator.Error{}
}

// checkReplacesForCSVs generates an error if value of the `replaces` field in the
// csv does not match the `metadata.Name` field of the old csv to be replaced.
// It also generates a warning if the `replaces` field of a csv is empty.
func checkReplacesForCSVs(csvReplacesMap map[string]string, csvsInBundle []string, manifestResult validator.ManifestResult) validator.ManifestResult {
	for pathCSV, replaces := range csvReplacesMap {
		if replaces == "" {
			manifestResult.Add(validator.WarnInvalidCSV("`spec.replaces` field not present. If this csv replaces an old version, populate this field with the `metadata.Name` of the old csv", pathCSV))
		} else if !isStringPresent(csvsInBundle, replaces) {
			manifestResult.Add(validator.ErrInvalidCSV(fmt.Sprintf("`%s` mentioned in the `spec.replaces` field ofnot present in the manifest", replaces), pathCSV))
		}
	}
	return manifestResult
}

func isStringPresent(list []string, val string) bool {
	for _, str := range list {
		if val == str {
			return true
		}
	}
	return false
}
