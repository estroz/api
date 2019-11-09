package internal

import (
	"strings"

	"github.com/operator-framework/api/pkg/validation/errors"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/install"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/validation"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/client-go/kubernetes/scheme"
)

var Scheme = scheme.Scheme

func init() {
	install.Install(Scheme)
}

type crdv1beta1Validator func(*v1beta1.CustomResourceDefinition) errors.ManifestResult

func (f crdv1beta1Validator) Validate(objs ...interface{}) (results []errors.ManifestResult) {
	for _, obj := range objs {
		switch v := obj.(type) {
		case *v1beta1.CustomResourceDefinition:
			results = append(results, f(v))
		}
	}
	return results
}

// Validatev1beta1CRD validates the given v1beta1 CRD object.
var Validatev1beta1CRD crdv1beta1Validator = func(crd *v1beta1.CustomResourceDefinition) (result errors.ManifestResult) {
	unversionedCRD := apiextensions.CustomResourceDefinition{}
	err := Scheme.Converter().Convert(&crd, &unversionedCRD, conversion.SourceToDest, nil)
	if err != nil {
		result.Add(errors.ErrInvalidParse("error converting versioned crd to unversioned crd", err))
		return result
	}
	result = validateCRDUnversioned(&unversionedCRD)
	result.Name = unversionedCRD.GetName()
	return result
}

func validateCRDUnversioned(crd *apiextensions.CustomResourceDefinition) (result errors.ManifestResult) {
	errList := validation.ValidateCustomResourceDefinition(crd)
	for _, err := range errList {
		if !strings.Contains(err.Field, "openAPIV3Schema") && !strings.Contains(err.Field, "status") {
			result.Add(errors.NewError(errors.ErrorType(err.Type), err.Error(), err.Field, err.BadValue))
		}
	}
	return result
}
