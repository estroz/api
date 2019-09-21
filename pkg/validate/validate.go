package validate

import (
	operatorsv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes/scheme"
)

var Scheme = scheme.Scheme

func init() {
	if err := operatorsv1alpha1.AddToScheme(Scheme); err != nil {
		panic(err)
	}
	if err := apiextv1beta1.AddToScheme(Scheme); err != nil {
		panic(err)
	}
}
