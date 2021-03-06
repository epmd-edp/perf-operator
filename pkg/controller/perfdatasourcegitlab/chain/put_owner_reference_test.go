package chain

import (
	v1alpha12 "github.com/epmd-edp/codebase-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/perf-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

const (
	fakeName      = "fake-name"
	fakeNamespace = "fake-namespace"
)

func TestPutOwnerReference_PerfDataSourceContainsPerfServerOwnerReference(t *testing.T) {
	pds := &v1alpha1.PerfDataSourceGitLab{
		ObjectMeta: v1.ObjectMeta{
			OwnerReferences: []v1.OwnerReference{
				{
					Kind: "Codebase",
				},
			},
		},
	}
	ch := PutOwnerReference{}
	assert.NoError(t, ch.ServeRequest(pds))
}

func TestPutOwnerReference_ShouldSetOwnerReference(t *testing.T) {
	pds := &v1alpha1.PerfDataSourceGitLab{
		ObjectMeta: v1.ObjectMeta{
			Namespace: fakeNamespace,
		},
		Spec: v1alpha1.PerfDataSourceGitLabSpec{
			PerfServerName: fakeName,
		},
	}

	c := &v1alpha12.Codebase{
		ObjectMeta: v1.ObjectMeta{
			Name:      fakeName,
			Namespace: fakeNamespace,
		},
	}

	objs := []runtime.Object{
		pds, c,
	}

	s := scheme.Scheme
	s.AddKnownTypes(v1.SchemeGroupVersion, pds, c)

	ch := PutOwnerReference{
		scheme: s,
		client: fake.NewFakeClient(objs...),
	}
	assert.NoError(t, ch.ServeRequest(pds))
}

func TestPutOwnerReference_PerfServerShouldNotBeFound(t *testing.T) {
	pds := &v1alpha1.PerfDataSourceGitLab{
		ObjectMeta: v1.ObjectMeta{
			Namespace: fakeNamespace,
		},
		Spec: v1alpha1.PerfDataSourceGitLabSpec{
			PerfServerName: fakeName,
		},
	}

	ps := &v1alpha1.PerfServer{}

	objs := []runtime.Object{
		pds, ps,
	}

	s := scheme.Scheme
	s.AddKnownTypes(v1.SchemeGroupVersion, pds, ps)

	ch := PutOwnerReference{
		scheme: s,
		client: fake.NewFakeClient(objs...),
	}
	assert.Error(t, ch.ServeRequest(pds))
}
