package validation

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/operator-framework/api/pkg/validation/validator"

	olmapiv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/operator-framework/operator-registry/pkg/appregistry"
	"github.com/operator-framework/operator-registry/pkg/registry"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

type bundleValidator struct {
	pkg     registry.PackageManifest
	bundles []*registry.Bundle
}

func NewBundleValidator(pkg registry.PackageManifest, bundles ...*registry.Bundle) validator.Validator {
	return &bundleValidator{
		pkg:     pkg,
		bundles: bundles,
	}
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

func (v bundleValidator) validateBundle() (results validator.ManifestResult) {
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
			results.Add(validator.WarnInvalidCSV("`spec.replaces` field matches its own `metadata.name`. It should contain `metadata.name` of the old CSV to be replaced", csv.GetName()))
		}
		results.Add(validateOwnedCRDs(bundle, csv)...)
	}
	results.Add(checkReplacesForCSVs(csvReplacesMap, csvsInBundle)...)
	results.Add(checkDefaultChannelInBundle(v.pkg, csvsInBundle))
	return results
}

func checkDefaultChannelInBundle(pkg registry.PackageManifest, csvsInBundle []string) validator.Error {
	for _, channel := range pkg.Channels {
		if !isStringPresent(csvsInBundle, channel.CurrentCSVName) {
			return validator.ErrInvalidBundle(fmt.Sprintf("currentCSV `%s` for channel name `%s` in package `%s` not found in bundle", channel.CurrentCSVName, channel.Name, pkg.PackageName), channel.CurrentCSVName)
		}
	}
	return validator.Error{}
}

func validateOwnedCRDs(bundle *registry.Bundle, csv *olmapiv1alpha1.ClusterServiceVersion) (errs []validator.Error) {
	ownedCrdNames := getOwnedCustomResourceDefintionNames(csv)
	bundleCrdNames, err := getBundleCRDNames(bundle)
	if err != (validator.Error{}) {
		return []validator.Error{err}
	}

	// validating names
	for _, ownedCrd := range ownedCrdNames {
		if !bundleCrdNames[ownedCrd] {
			errs = append(errs, validator.ErrInvalidBundle(fmt.Sprintf("owned crd (%s) not found in bundle %s", ownedCrd, bundle.Name), ownedCrd))
		} else {
			delete(bundleCrdNames, ownedCrd)
		}
	}
	// CRDs not defined in the CSV present in the bundle
	if len(bundleCrdNames) != 0 {
		for crd, _ := range bundleCrdNames {
			errs = append(errs, validator.WarnInvalidBundle(fmt.Sprintf("`%s` crd present in bundle `%s` not defined in csv", crd, bundle.Name), crd))
		}
	}
	return errs
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
func checkReplacesForCSVs(csvReplacesMap map[string]string, csvsInBundle []string) (errs []validator.Error) {
	for pathCSV, replaces := range csvReplacesMap {
		if replaces == "" {
			errs = append(errs, validator.WarnInvalidCSV("`spec.replaces` field not present. If this csv replaces an old version, populate this field with the `metadata.Name` of the old csv", pathCSV))
		} else if !isStringPresent(csvsInBundle, replaces) {
			errs = append(errs, validator.ErrInvalidCSV(fmt.Sprintf("`%s` mentioned in the `spec.replaces` field ofnot present in the manifest", replaces), pathCSV))
		}
	}
	return errs
}

func isStringPresent(list []string, val string) bool {
	for _, str := range list {
		if val == str {
			return true
		}
	}
	return false
}

// validateBundle ensures all objects in bundle have the correct data.
func validateBundle(bundle *registry.Bundle) (err error) {
	bcsv, err := bundle.ClusterServiceVersion()
	if err != nil {
		return err
	}
	csv := mustBundleCSVToCSV(bcsv)
	crds, err := bundle.CustomResourceDefinitions()
	if err != nil {
		return err
	}
	crdMap := map[string]struct{}{}
	for _, crd := range crds {
		for _, k := range getCRDKeys(crd) {
			crdMap[k.String()] = struct{}{}
		}
	}
	// If at least one CSV has an owned CRD it must be present.
	if len(csv.Spec.CustomResourceDefinitions.Owned) > 0 && len(crds) == 0 {
		return errors.Errorf("bundled CSV has an owned CRD but no CRD's are present in bundle dir")
	}
	// Ensure all CRD's referenced in each CSV exist in BundleDir.
	for _, o := range csv.Spec.CustomResourceDefinitions.Owned {
		key := getCRDDescKey(o)
		if _, hasCRD := crdMap[key.String()]; !hasCRD {
			return errors.Errorf("bundle dir does not contain owned CRD %q from CSV %q", key, csv.GetName())
		}
	}
	if !hasSupportedInstallMode(csv) {
		return errors.Errorf("at least one installMode must be marked \"supported\" in CSV %q", csv.GetName())
	}
	return nil
}

// hasSupportedInstallMode returns true if a csv supports at least one
// installMode.
func hasSupportedInstallMode(csv *olmapiv1alpha1.ClusterServiceVersion) bool {
	for _, mode := range csv.Spec.InstallModes {
		if mode.Supported {
			return true
		}
	}
	return false
}

// getCRDKeys returns a key uniquely identifying crd.
func getCRDDescKey(crd olmapiv1alpha1.CRDDescription) appregistry.CRDKey {
	return appregistry.CRDKey{
		Kind:    crd.Kind,
		Name:    crd.Name,
		Version: crd.Version,
	}
}

// getCRDKeys returns a set of keys uniquely identifying crd per version.
// getCRDKeys assumes at least one of spec.version, spec.versions is non-empty.
func getCRDKeys(crd *v1beta1.CustomResourceDefinition) (keys []appregistry.CRDKey) {
	if crd.Spec.Version != "" && len(crd.Spec.Versions) == 0 {
		return []appregistry.CRDKey{{
			Kind:    crd.Spec.Names.Kind,
			Name:    crd.GetName(),
			Version: crd.Spec.Version,
		},
		}
	}
	for _, v := range crd.Spec.Versions {
		keys = append(keys, appregistry.CRDKey{
			Kind:    crd.Spec.Names.Kind,
			Name:    crd.GetName(),
			Version: v.Name,
		})
	}
	return keys
}
