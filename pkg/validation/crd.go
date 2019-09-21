package validation

import (
	"strings"

	"github.com/operator-framework/api/pkg/validation/validator"

	"github.com/operator-framework/operator-registry/pkg/appregistry"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/validation"
	"k8s.io/apimachinery/pkg/conversion"
)

type crdValidator struct {
	crds map[appregistry.CRDKey]*v1beta1.CustomResourceDefinition
}

func NewCRDValidator(crds ...*v1beta1.CustomResourceDefinition) validator.Validator {
	val := crdValidator{crds: map[appregistry.CRDKey]*v1beta1.CustomResourceDefinition{}}
	for _, crd := range crds {
		val.crds[apiKey(crd)] = crd
	}
	return &val
}

func (v *crdValidator) Validate() (results []validator.ManifestResult) {
	for key, crd := range v.crds {
		result := validateCRD(crd)
		result.Name = key.String()
		results = append(results, result)
	}
	return results
}

func apiKey(v *v1beta1.CustomResourceDefinition) appregistry.CRDKey {
	// TODO: support multiple versions.
	key := appregistry.CRDKey{Name: v.GetName(), Kind: v.Spec.Names.Kind}
	key.Version = v.Spec.Version
	if key.Version == "" {
		if len(v.Spec.Versions) == 0 {
			// QUESTION: is this unique enough, or should we just return an error
			key.Version = "badVer" + key.String()
		} else {
			key.Version = v.Spec.Versions[0].Name
		}
	}
	return key
}

func (v crdValidator) Name() string {
	return "CustomResourceDefinition Validator"
}

func validateCRD(crd *v1beta1.CustomResourceDefinition) (results validator.ManifestResult) {
	unversionedCRD := apiextensions.CustomResourceDefinition{}
	err := Scheme.Converter().Convert(&crd, &unversionedCRD, conversion.SourceToDest, nil)
	if err != nil {
		results.Add(validator.ErrInvalidParse(err.Error(), nil))
		return results
	}
	errList := validation.ValidateCustomResourceDefinition(&unversionedCRD)
	for _, err := range errList {
		if !strings.Contains(err.Field, "openAPIV3Schema") && !strings.Contains(err.Field, "status") {
			results.Add(validator.NewError(validator.ErrorType(err.Type), err.Error(), err.Field, err.BadValue))
		}
	}
	return results
}
