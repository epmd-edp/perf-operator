package util

import (
	"context"
	"github.com/epmd-edp/perf-operator/pkg/apis/edp/v1alpha1"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetSecret(client client.Client, name, namespace string) (s *coreV1.Secret, err error) {
	err = client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, s)
	if err != nil {
		return nil, err
	}
	return
}

func GetOwnerReference(ownerKind string, ors []metav1.OwnerReference) *metav1.OwnerReference {
	if len(ors) == 0 {
		return nil
	}
	for _, o := range ors {
		if o.Kind == ownerKind {
			return &o
		}
	}
	return nil
}

func GetPerfServerCr(c client.Client, name, namespace string) (ps *v1alpha1.PerfServer, err error) {
	nsn := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	if err := c.Get(context.TODO(), nsn, ps); err != nil {
		return nil, err
	}
	return
}
