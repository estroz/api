package validation

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/operator-framework/api/pkg/validation/errors"
	interfaces "github.com/operator-framework/api/pkg/validation/interfaces"

	operatorsv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/operator-framework/operator-registry/pkg/registry"
	"github.com/sirupsen/logrus"
)

type BundleValidator struct{}

func (f BundleValidator) GetFuncs(objs ...interface{}) (vals interfaces.ValidatorFuncs) {
	for _, obj := range objs {
		switch v := obj.(type) {
		case *registry.Bundle:
			vals = append(vals, func() errors.ManifestResult {
				return validateBundle(v)
			})
		}
	}
	return vals
}

func validateBundle(bundle *registry.Bundle) errors.ManifestResult {
	bcsv, err := bundle.ClusterServiceVersion()
	if err != nil {
		logrus.Fatal(err)
		return errors.ManifestResult{}
	}
	csv := mustBundleCSVToCSV(bcsv)
	result := validateOwnedCRDs(bundle, csv)
	result.Name = csv.Spec.Version.String()
	return result
}

func mustBundleCSVToCSV(bcsv *registry.ClusterServiceVersion) *operatorsv1alpha1.ClusterServiceVersion {
	spec := operatorsv1alpha1.ClusterServiceVersionSpec{}
	if err := json.Unmarshal(bcsv.Spec, &spec); err != nil {
		log.Fatalf("Error converting bundle CSV %q type: %v", bcsv.GetName(), err)
	}
	return &operatorsv1alpha1.ClusterServiceVersion{
		TypeMeta:   bcsv.TypeMeta,
		ObjectMeta: bcsv.ObjectMeta,
		Spec:       spec,
	}
}

func validateOwnedCRDs(bundle *registry.Bundle, csv *operatorsv1alpha1.ClusterServiceVersion) (result errors.ManifestResult) {
	ownedCrdNames := getOwnedCustomResourceDefintionNames(csv)
	bundleCrdNames, err := getBundleCRDNames(bundle)
	if err != (errors.Error{}) {
		result.Add(err)
		return result
	}

	// validating names
	for _, ownedCrd := range ownedCrdNames {
		if !bundleCrdNames[ownedCrd] {
			result.Add(errors.ErrInvalidBundle(fmt.Sprintf("owned crd (%s) not found in bundle %s", ownedCrd, bundle.Name), ownedCrd))
		} else {
			delete(bundleCrdNames, ownedCrd)
		}
	}
	// CRDs not defined in the CSV present in the bundle
	if len(bundleCrdNames) != 0 {
		for crd, _ := range bundleCrdNames {
			result.Add(errors.WarnInvalidBundle(fmt.Sprintf("`%s` crd present in bundle `%s` not defined in csv", crd, bundle.Name), crd))
		}
	}
	return result
}

func getOwnedCustomResourceDefintionNames(csv *operatorsv1alpha1.ClusterServiceVersion) []string {
	var names []string
	for _, ownedCrd := range csv.Spec.CustomResourceDefinitions.Owned {
		names = append(names, ownedCrd.Name)
	}
	return names
}

func getBundleCRDNames(bundle *registry.Bundle) (map[string]bool, errors.Error) {
	crds, err := bundle.CustomResourceDefinitions()
	if err != nil {
		logrus.Fatal(err)
		return nil, errors.Error{}
	}
	bundleCrdNames := make(map[string]bool)
	for _, crd := range crds {
		bundleCrdNames[crd.GetName()] = true
	}
	return bundleCrdNames, errors.Error{}
}
