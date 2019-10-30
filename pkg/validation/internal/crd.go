package internal

import (
	"strings"

	"github.com/operator-framework/api/pkg/validation/errors"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextv1beta "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/validation"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/client-go/kubernetes/scheme"
)

var Scheme = scheme.Scheme

func init() {
	if err := apiextensions.AddToScheme(Scheme); err != nil {
		panic(err)
	}
	if err := apiextv1beta.AddToScheme(Scheme); err != nil {
		panic(err)
	}
}

type CRDValidator struct{}

func (f CRDValidator) Validate(objs ...interface{}) (results []errors.ManifestResult) {
	for _, obj := range objs {
		switch v := obj.(type) {
		case *apiextv1beta.CustomResourceDefinition:
			results = append(results, validateCRD(v))
		}
	}
	return results
}

func validateCRD(crd interface{}) (result errors.ManifestResult) {
	unversionedCRD := apiextensions.CustomResourceDefinition{}
	err := Scheme.Converter().Convert(&crd, &unversionedCRD, conversion.SourceToDest, nil)
	if err != nil {
		result.Add(errors.ErrInvalidParse("error converting versioned crd to unversioned crd", err))
		return result
	}
	result = validateCRD(&unversionedCRD)
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
