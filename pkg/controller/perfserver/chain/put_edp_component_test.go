package chain

import (
	"fmt"
	edpApi "github.com/epmd-edp/edp-component-operator/pkg/apis/v1/v1alpha1"
	"github.com/epmd-edp/perf-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

const (
	fakeNamespace = "fake-namespace"
	dirPath       = "/usr/local/configs/img"
	fileName      = "perf.svg"
)

func TestPutEdpComponent_ShouldCreateEdpComponent(t *testing.T) {
	createDirIfNotExist()

	edpComp := &edpApi.EDPComponent{}

	psr := &v1alpha1.PerfServer{
		ObjectMeta: v1.ObjectMeta{
			Name:      fakeName,
			Namespace: fakeNamespace,
		},
	}

	objs := []runtime.Object{
		edpComp, psr,
	}
	s := scheme.Scheme
	s.AddKnownTypes(v1.SchemeGroupVersion, edpComp, psr)

	ch := PutEdpComponent{
		scheme: s,
		client: fake.NewFakeClient(objs...),
	}

	assert.NoError(t, ch.ServeRequest(psr))

	removeFile()
}

func createDirIfNotExist() {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			panic(err)
		}
	}
	f, err := os.OpenFile(fmt.Sprintf("%v/%v", dirPath, fileName), os.O_RDONLY|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	f.Close()
}

func removeFile() {
	if err := os.Remove(fmt.Sprintf("%v/%v", dirPath, fileName)); err != nil {
		panic(err)
	}
}

func TestPutEdpComponent_EdpComponentAlreadyExists(t *testing.T) {
	edpComp := &edpApi.EDPComponent{
		ObjectMeta: v1.ObjectMeta{
			Name:      fakeName,
			Namespace: fakeNamespace,
		},
	}

	objs := []runtime.Object{
		edpComp,
	}
	s := scheme.Scheme
	s.AddKnownTypes(v1.SchemeGroupVersion, edpComp)

	ch := PutEdpComponent{
		scheme: s,
		client: fake.NewFakeClient(objs...),
	}

	psr := &v1alpha1.PerfServer{
		ObjectMeta: v1.ObjectMeta{
			Name:      fakeName,
			Namespace: fakeNamespace,
		},
	}

	assert.NoError(t, ch.ServeRequest(psr))
}

func TestPutEdpComponent_SchemeDoesntContainEdpComponent(t *testing.T) {
	s := scheme.Scheme
	s.AddKnownTypes(v1.SchemeGroupVersion)

	ch := PutEdpComponent{
		scheme: s,
		client: fake.NewFakeClient([]runtime.Object{}...),
	}

	psr := &v1alpha1.PerfServer{
		ObjectMeta: v1.ObjectMeta{
			Name:      fakeName,
			Namespace: fakeNamespace,
		},
	}

	assert.Error(t, ch.ServeRequest(psr))
}

func TestPutEdpComponent_IconDoesntExist(t *testing.T) {
	edpComp := &edpApi.EDPComponent{}

	psr := &v1alpha1.PerfServer{
		ObjectMeta: v1.ObjectMeta{
			Name:      fakeName,
			Namespace: fakeNamespace,
		},
	}

	objs := []runtime.Object{
		edpComp, psr,
	}
	s := scheme.Scheme
	s.AddKnownTypes(v1.SchemeGroupVersion, edpComp, psr)

	ch := PutEdpComponent{
		scheme: s,
		client: fake.NewFakeClient(objs...),
	}

	assert.Error(t, ch.ServeRequest(psr))
}
