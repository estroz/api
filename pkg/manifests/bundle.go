package manifests

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/operator-framework/api/pkg/metadata/annotations"
	"github.com/operator-framework/api/pkg/metadata/dependencies"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
)

type Bundle struct {
	// Identifying information about this bundle.
	Name        string
	BundleImage string

	// Manifests
	Objects     []*unstructured.Unstructured
	CSV         *operatorsv1alpha1.ClusterServiceVersion
	V1beta1CRDs []*apiextensionsv1beta1.CustomResourceDefinition
	V1CRDs      []*apiextensionsv1.CustomResourceDefinition

	// Metadata
	Annotations  annotations.File
	Dependencies dependencies.File
}

func (b *Bundle) ObjectsToValidate() []interface{} {
	objs := []interface{}{}
	for _, crd := range b.V1CRDs {
		objs = append(objs, crd)
	}
	for _, crd := range b.V1beta1CRDs {
		objs = append(objs, crd)
	}
	objs = append(objs, b.CSV)
	objs = append(objs, b)

	return objs
}
