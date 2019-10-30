package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/operator-framework/api/pkg/validation/errors"
	"github.com/operator-framework/operator-registry/pkg/registry"

	operatorsv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	// validate example annotations ("alm-examples", "olm.examples").
	result.Add(validateExamplesAnnotations(csv)...)
	// validate installModes
	result.Add(validateInstallModes(csv)...)
	// check missing optional/mandatory fields.
	result.Add(checkFields(csv)...)
	return result
}

// checkFields runs checkMissingFields and returns its errors.
func checkFields(csv *operatorsv1alpha1.ClusterServiceVersion) (errs []errors.Error) {
	result := errors.ManifestResult{}
	checkMissingFields(&result, reflect.ValueOf(csv), "")
	return append(result.Errors, result.Warnings...)
}

// Recursive function that traverses a nested struct passed in as reflect value, and reports for errors/warnings
// in case of null struct field values.
func checkMissingFields(result *errors.ManifestResult, v reflect.Value, parentStructName string) {
	if v.Kind() != reflect.Struct {
		return
	}
	typ := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		fieldType := typ.Field(i)

		tag := fieldType.Tag.Get("json")
		// Ignore fields that are subsets of a primitive field.
		if tag == "" {
			continue
		}

		// Omitted field tags will contain ",omitempty", and ignored tags will
		// match "-" exactly, respectively.
		isOptionalField := strings.Contains(tag, ",omitempty") || tag == "-"
		emptyVal := isEmptyValue(fieldValue)

		newParentStructName := fieldType.Name
		if parentStructName != "" {
			newParentStructName = parentStructName + "." + newParentStructName
		}

		switch fieldValue.Kind() {
		case reflect.Struct:
			updateLog(result, "struct", newParentStructName, emptyVal, isOptionalField)
			if !emptyVal {
				checkMissingFields(result, fieldValue, newParentStructName)
			}
		default:
			updateLog(result, "field", newParentStructName, emptyVal, isOptionalField)
		}
	}
}

// Returns updated error log with missing optional/mandatory field/struct objects.
func updateLog(result *errors.ManifestResult, typeName string, newParentStructName string, emptyVal bool, isOptionalField bool) {
	if !emptyVal {
		return
	}
	if isOptionalField {
		// TODO: update the value field (typeName).
		result.Add(errors.WarnFieldMissing("", newParentStructName, typeName))
	} else if newParentStructName != "Status" {
		// TODO: update the value field (typeName).
		result.Add(errors.ErrFieldMissing("", newParentStructName, typeName))
	}
}

// Uses reflect package to check if the value of the object passed is null, returns a boolean accordingly.
// TODO: replace with reflect.Kind.IsZero() in go 1.13
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		// Check if the value for 'Spec.InstallStrategy.StrategySpecRaw' field is present. This field is a RawMessage value type. Without a value, the key is explicitly set to 'null'.
		if fieldValue, ok := v.Interface().(json.RawMessage); ok {
			valString := string(fieldValue)
			if valString == "null" {
				return true
			}
		}
		return v.Len() == 0
	// Currently the only CSV field with integer type is containerPort. Operator Verification Library raises a warning if containerPort field is missisng or if its value is 0.
	// It is an optional field so the user can ignore the warning saying this field is missing if they intend to use port 0.
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		for i, n := 0, v.NumField(); i < n; i++ {
			if !isEmptyValue(v.Field(i)) {
				return false
			}
		}
		return true
	default:
		panic(fmt.Sprintf("%v kind is not supported.", v.Kind()))
	}
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
