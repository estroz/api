package internal

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/operator-framework/api/pkg/validation/errors"
	"github.com/operator-framework/operator-registry/pkg/registry"

	"github.com/blang/semver"
	operatorsv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8svalidation "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type CSVValidator struct{}

func (f CSVValidator) Validate(objs ...interface{}) (results []errors.ManifestResult) {
	for _, obj := range objs {
		switch v := obj.(type) {
		case *operatorsv1alpha1.ClusterServiceVersion:
			results = append(results, validateCSV(v))
		case *registry.ClusterServiceVersion:
			results = append(results, validateCSVRegistry(v))
		}
	}
	return results
}

func validateCSVRegistry(bcsv *registry.ClusterServiceVersion) (result errors.ManifestResult) {
	csv, err := bundleCSVToCSV(bcsv)
	if err != (errors.Error{}) {
		result.Add(err)
		return result
	}
	return validateCSV(csv)
}

// Iterates over the given CSV. Returns a ManifestResult type object.
func validateCSV(csv *operatorsv1alpha1.ClusterServiceVersion) errors.ManifestResult {
	result := errors.ManifestResult{Name: csv.GetName()}
	// Ensure CSV names are of the correct format.
	if _, err := parseNameVersion(csv); err != (errors.Error{}) {
		result.Add(err)
	}
	if csv.Spec.Replaces != "" {
		if _, err := parseReplacesVersion(csv); err != (errors.Error{}) {
			result.Add(err)
		}
	}
	// validate example annotations ("alm-examples", "olm.examples").
	result.Add(validateExamplesAnnotations(csv)...)
	// validate installModes
	result.Add(validateInstallModes(csv)...)
	// check missing optional/mandatory fields.
	result.Add(checkFields(csv)...)
	return result
}

func parseCSVNameFormat(name string) (string, semver.Version, error) {
	splitName := strings.SplitN(name, ".", 2)
	if len(splitName) != 2 {
		return "", semver.Version{}, fmt.Errorf("%q must have format: {operator name}.vX.Y.Z", name)
	}
	if violations := k8svalidation.IsDNS1123Subdomain(splitName[0]); len(violations) != 0 {
		return "", semver.Version{}, fmt.Errorf("%q is invalid:\n%s", splitName[0], violations)
	}
	nameVer, err := semver.Parse(splitName[1])
	if err != nil {
		return "", semver.Version{}, fmt.Errorf("%q contains an invalid semver %q", name, splitName[1])
	}
	return splitName[0], nameVer, errors.Error{}
}

func parseCSVNameVersion(name, fieldPath, csvName string) (semver.Version, errors.Error) {
	_, replacesVer, err := parseCSVNameFormat(name)
	if err != nil {
		return semver.Version{}, errors.ErrInvalidCSV(fmt.Sprintf("%s %s", fieldPath, err), csvName)
	}
	return replacesVer, errors.Error{}
}

func parseNameVersion(csv *operatorsv1alpha1.ClusterServiceVersion) (semver.Version, errors.Error) {
	return parseCSVNameVersion(csv.GetName(), "metadata.name", csv.GetName())
}

func parseReplacesVersion(csv *operatorsv1alpha1.ClusterServiceVersion) (semver.Version, errors.Error) {
	return parseCSVNameVersion(csv.Spec.Replaces, "spec.replaces", csv.GetName())
}

// checkFields runs checkMissingFields and returns its errors.
func checkFields(csv *operatorsv1alpha1.ClusterServiceVersion) (errs []errors.Error) {
	result := errors.ManifestResult{}
	checkMissingFields(&result, reflect.ValueOf(csv), "")
	return append(result.Errors, result.Warnings...)
}

// validateExamplesAnnotations compares alm/olm example annotations with provided APIs given
// by Spec.CustomResourceDefinitions.Owned and Spec.APIServiceDefinitions.Owned.
func validateExamplesAnnotations(csv *operatorsv1alpha1.ClusterServiceVersion) (errs []errors.Error) {
	annotations := csv.ObjectMeta.GetAnnotations()
	// Return right away if no examples annotations are found.
	if len(annotations) == 0 {
		errs = append(errs, errors.WarnInvalidCSV("annotations not found", csv.GetName()))
		return errs
	}
	// Expect either `alm-examples` or `olm.examples` but not both
	// If both are present, `alm-examples` will be used
	var examplesString string
	almExamples, almOK := annotations["alm-examples"]
	olmExamples, olmOK := annotations["olm.examples"]
	if !almOK && !olmOK {
		errs = append(errs, errors.WarnInvalidCSV("example annotations not found", csv.GetName()))
		return errs
	} else if almOK {
		if olmOK {
			errs = append(errs, errors.WarnInvalidCSV("both `alm-examples` and `olm.examples` are present. Checking only `alm-examples`", csv.GetName()))
		}
		examplesString = almExamples
	} else {
		examplesString = olmExamples
	}
	us := []unstructured.Unstructured{}
	dec := yaml.NewYAMLOrJSONDecoder(strings.NewReader(examplesString), 8)
	if err := dec.Decode(&us); err != nil && err != io.EOF {
		errs = append(errs, errors.ErrInvalidParse("error decoding example CustomResource", err))
		return errs
	}
	parsed := map[schema.GroupVersionKind]struct{}{}
	for _, u := range us {
		parsed[u.GetObjectKind().GroupVersionKind()] = struct{}{}
	}

	providedAPIs, aerrs := getProvidedAPIs(csv)
	errs = append(errs, aerrs...)

	errs = append(errs, matchGVKProvidedAPIs(parsed, providedAPIs)...)
	return errs
}

func getProvidedAPIs(csv *operatorsv1alpha1.ClusterServiceVersion) (provided map[schema.GroupVersionKind]struct{}, errs []errors.Error) {
	provided = map[schema.GroupVersionKind]struct{}{}
	for _, owned := range csv.Spec.CustomResourceDefinitions.Owned {
		parts := strings.SplitN(owned.Name, ".", 2)
		if len(parts) < 2 {
			errs = append(errs, errors.ErrInvalidParse(fmt.Sprintf("couldn't parse plural.group from crd name: %s", owned.Name), nil))
			continue
		}
		provided[newGVK(parts[1], owned.Version, owned.Kind)] = struct{}{}
	}

	for _, api := range csv.Spec.APIServiceDefinitions.Owned {
		provided[newGVK(api.Group, api.Version, api.Kind)] = struct{}{}
	}

	return provided, errs
}

func newGVK(g, v, k string) schema.GroupVersionKind {
	return schema.GroupVersionKind{Group: g, Version: v, Kind: k}
}

func matchGVKProvidedAPIs(examples map[schema.GroupVersionKind]struct{}, providedAPIs map[schema.GroupVersionKind]struct{}) (errs []errors.Error) {
	for key := range examples {
		if _, ok := providedAPIs[key]; !ok {
			errs = append(errs, errors.ErrInvalidOperation(fmt.Sprintf("couldn't match %v in provided APIs list: %v", key, providedAPIs), key))
		}
	}
	return errs
}

func validateInstallModes(csv *operatorsv1alpha1.ClusterServiceVersion) (errs []errors.Error) {
	if len(csv.Spec.InstallModes) == 0 {
		errs = append(errs, errors.ErrInvalidCSV("install modes not found", csv.GetName()))
		return errs
	}

	installModeSet := operatorsv1alpha1.InstallModeSet{}
	anySupported := false
	for _, installMode := range csv.Spec.InstallModes {
		if _, ok := installModeSet[installMode.Type]; ok {
			errs = append(errs, errors.ErrInvalidCSV("duplicate install modes present", csv.GetName()))
		} else if installMode.Supported {
			anySupported = true
		}
	}

	// all installModes should not be `false`
	if !anySupported {
		errs = append(errs, errors.ErrInvalidCSV("none of InstallModeTypes are supported", csv.GetName()))
	}
	return errs
}
