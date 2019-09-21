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
	bundles []*registry.Bundle
}

func NewBundleValidator(bundles ...*registry.Bundle) validator.Validator {
	return &bundleValidator{
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
		crds, err := bundle.CustomResourceDefinitions()
		if err != nil {
			logrus.Fatal(err)
			return nil
		}
		vals := validator.NewValidators(NewCSVValidator(csv), NewCRDValidator(crds...))
		results = append(results, vals.ValidateAll()...)
	}
	result := v.validateBundle()
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
	for _, bundle := range v.bundles {
		bcsv, err := bundle.ClusterServiceVersion()
		if err != nil {
			logrus.Fatal(err)
			return validator.ManifestResult{}
		}
		csv := mustBundleCSVToCSV(bcsv)
		results.Add(validateOwnedCRDs(bundle, csv)...)
	}
	return results
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
