package validation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/operator-framework/api/pkg/validation/errors"
	interfaces "github.com/operator-framework/api/pkg/validation/interfaces"

	operatorsv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type CSVValidator struct{}

func (f CSVValidator) GetFuncs(objs ...interface{}) (funcs interfaces.ValidatorFuncs) {
	for _, obj := range objs {
		switch v := obj.(type) {
		case *operatorsv1alpha1.ClusterServiceVersion:
			funcs = append(funcs, func() errors.ManifestResult {
				return validateCSV(v)
			})
		}
	}
	return funcs
}

// Iterates over the given CSV. Returns a ManifestResult type object.
func validateCSV(csv *operatorsv1alpha1.ClusterServiceVersion) errors.ManifestResult {
	result := errors.ManifestResult{Name: csv.GetName()}

	// validate example annotations ("alm-examples", "olm.examples").
	result.Add(validateExamplesAnnotations(csv)...)

	// validate installModes
	result.Add(validateInstallModes(csv)...)

	// check missing optional/mandatory fields.
	fieldValue := reflect.ValueOf(csv)

	switch fieldValue.Kind() {
	case reflect.Struct:
		checkMissingFields(&result, fieldValue, "")
		return result
	}
	return result
}

// Recursive function that traverses a nested struct passed in as reflect value, and reports for errors/warnings
// in case of null struct field values.
func checkMissingFields(log *errors.ManifestResult, v reflect.Value, parentStructName string) {

	for i := 0; i < v.NumField(); i++ {

		fieldValue := v.Field(i)

		tag := v.Type().Field(i).Tag.Get("json")
		// Ignore fields that are subsets of a primitive field.
		if tag == "" {
			continue
		}

		fields := strings.Split(tag, ",")
		isOptionalField := containsStrict(fields, "omitempty")
		emptyVal := isEmptyValue(fieldValue)

		newParentStructName := ""
		if parentStructName == "" {
			newParentStructName = v.Type().Field(i).Name
		} else {
			newParentStructName = parentStructName + "." + v.Type().Field(i).Name
		}

		switch fieldValue.Kind() {
		case reflect.Struct:
			updateLog(log, "struct", newParentStructName, emptyVal, isOptionalField)
			if !emptyVal {
				checkMissingFields(log, fieldValue, newParentStructName)
			}
		default:
			updateLog(log, "field", newParentStructName, emptyVal, isOptionalField)
		}
	}
}

// Returns updated error log with missing optional/mandatory field/struct objects.
func updateLog(log *errors.ManifestResult, typeName string, newParentStructName string, emptyVal bool, isOptionalField bool) {

	if emptyVal && isOptionalField {
		// TODO: update the value field (typeName).
		log.Add(errors.WarnFieldMissing("", newParentStructName, typeName))
	} else if emptyVal && !isOptionalField {
		if newParentStructName != "Status" {
			// TODO: update the value field (typeName).
			log.Add(errors.ErrFieldMissing("", newParentStructName, typeName))
		}
	}
}

// Takes in a string slice and checks if a string (x) is present in the slice.
// Return true if the string is present in the slice.
func containsStrict(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// Uses reflect package to check if the value of the object passed is null, returns a boolean accordingly.
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
	var examples []apiextv1beta1.CustomResourceDefinition
	var annotationsExamples string
	annotations := csv.ObjectMeta.GetAnnotations()
	// Return right away if no examples annotations are found.
	if len(annotations) == 0 {
		errs = append(errs, errors.WarnInvalidCSV("annotations not found", csv.GetName()))
		return errs
	}
	// Expect either `alm-examples` or `olm.examples` but not both
	// If both are present, `alm-examples` will be used
	if value, ok := annotations["alm-examples"]; ok {
		annotationsExamples = value
		if _, ok = annotations["olm.examples"]; ok {
			// both `alm-examples` and `olm.examples` are present
			errs = append(errs, errors.WarnInvalidCSV("both `alm-examples` and `olm.examples` are present. Checking only `alm-examples`", csv.GetName()))
		}
	} else {
		annotationsExamples = annotations["olm.examples"]
	}

	// Can't find examples annotations, simply return
	if annotationsExamples == "" {
		errs = append(errs, errors.WarnInvalidCSV("example annotations not found", csv.GetName()))
		return errs
	}

	if err := json.Unmarshal([]byte(annotationsExamples), &examples); err != nil {
		errs = append(errs, errors.ErrInvalidParse(fmt.Sprintf("parsing example annotations to %T type:  %s ", examples, err), nil))
		return errs
	}

	providedAPIs, aerrs := getProvidedAPIs(csv)
	errs = append(errs, aerrs...)

	parsedExamples, perrs := parseExamplesAnnotations(examples)
	if len(perrs) != 0 {
		errs = append(errs, perrs...)
		return errs
	}

	errs = append(errs, matchGVKProvidedAPIs(parsedExamples, providedAPIs)...)
	return errs
}

func getProvidedAPIs(csv *operatorsv1alpha1.ClusterServiceVersion) (provided map[schema.GroupVersionKind]struct{}, errs []errors.Error) {
	provided = map[schema.GroupVersionKind]struct{}{}
	for _, owned := range csv.Spec.CustomResourceDefinitions.Owned {
		parts := strings.SplitN(owned.Name, ".", 2)
		if len(parts) < 2 {
			errs = append(errs, errors.ErrInvalidParse(fmt.Sprintf("couldn't parse plural.group from crd name: %s", owned.Name), owned.Name))
			continue
		}
		provided[schema.GroupVersionKind{Group: parts[1], Version: owned.Version, Kind: owned.Kind}] = struct{}{}
	}

	for _, api := range csv.Spec.APIServiceDefinitions.Owned {
		provided[schema.GroupVersionKind{Group: api.Group, Version: api.Version, Kind: api.Kind}] = struct{}{}
	}

	return provided, errs
}

func parseExamplesAnnotations(examples []apiextv1beta1.CustomResourceDefinition) (parsed map[schema.GroupVersionKind]struct{}, errs []errors.Error) {
	parsed = map[schema.GroupVersionKind]struct{}{}
	for _, value := range examples {
		parts := strings.SplitN(value.APIVersion, "/", 2)
		if len(parts) < 2 {
			errs = append(errs, errors.ErrInvalidParse(fmt.Sprintf("couldn't parse group/version from crd kind: %s", value.Kind), value.Kind))
			continue
		}
		parsed[schema.GroupVersionKind{Group: parts[0], Version: parts[1], Kind: value.Kind}] = struct{}{}
	}

	return parsed, errs
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
	// var installModeSet operatorsv1alpha1.InstallModeSet
	installModeSet := make(operatorsv1alpha1.InstallModeSet)
	for _, installMode := range csv.Spec.InstallModes {
		if _, ok := installModeSet[installMode.Type]; ok {
			errs = append(errs, errors.ErrInvalidCSV("duplicate install modes present", csv.GetName()))
		} else {
			installModeSet[installMode.Type] = installMode.Supported
		}
	}

	// installModes not found, return with a warning
	if len(installModeSet) == 0 {
		errs = append(errs, errors.WarnInvalidCSV("install modes not found", csv.GetName()))
		return errs
	}

	// all installModes should not be `false`
	if checkAllFalseForInstallModeSet(installModeSet) {
		errs = append(errs, errors.ErrInvalidCSV("none of InstallModeTypes are supported", csv.GetName()))
	}
	return errs
}

func checkAllFalseForInstallModeSet(installModeSet operatorsv1alpha1.InstallModeSet) bool {
	for _, isSupported := range installModeSet {
		if isSupported {
			return false
		}
	}
	return true
}
