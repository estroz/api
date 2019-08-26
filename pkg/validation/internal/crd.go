package validation

import (
	"strings"

	"github.com/operator-framework/api/pkg/validation/errors"
	interfaces "github.com/operator-framework/api/pkg/validation/interfaces"

	"github.com/operator-framework/operator-registry/pkg/appregistry"
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

func (f CRDValidator) GetFuncs(objs ...interface{}) (funcs interfaces.ValidatorFuncs) {
	for _, obj := range objs {
		switch v := obj.(type) {
		case *apiextv1beta.CustomResourceDefinition:
			funcs = append(funcs, func() errors.ManifestResult {
				return validateCRD(v)
			})
		}
	}
	return funcs
}

func apiKey(v *apiextv1beta.CustomResourceDefinition) appregistry.CRDKey {
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

func validateCRD(crd interface{}) errors.ManifestResult {
	unversionedCRD := apiextensions.CustomResourceDefinition{}
	err := Scheme.Converter().Convert(&crd, &unversionedCRD, conversion.SourceToDest, nil)
	if err != nil {
		return errors.ManifestResult{Errors: []errors.Error{errors.ErrInvalidParse(err.Error(), nil)}}
	}
	result := validateCRD(&unversionedCRD)
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
